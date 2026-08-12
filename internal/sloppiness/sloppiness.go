// Package sloppiness turns conservative, observable code smells into a bounded
// repair brief. It never edits source code and never starts another assessment.
package sloppiness

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	SchemaVersion = "1.0.0"
	MaxPasses     = 2
)

var sourceExtensions = map[string]bool{
	".cs": true, ".go": true, ".java": true, ".js": true, ".jsx": true,
	".py": true, ".rb": true, ".rs": true, ".swift": true, ".ts": true,
	".tsx": true,
}

var ignoredDirectories = map[string]bool{
	".git": true, ".idea": true, ".next": true, ".venv": true, "build": true,
	"coverage": true, "dist": true, "node_modules": true, "target": true,
	"vendor": true,
}

type Report struct {
	SchemaVersion  string           `json:"schema_version"`
	Root           string           `json:"root"`
	Score          int              `json:"score"`
	Interpretation string           `json:"interpretation"`
	Lines          int              `json:"source_lines"`
	SourceFiles    int              `json:"source_files"`
	TestFiles      int              `json:"test_files"`
	Dimensions     []Dimension      `json:"dimensions"`
	Findings       []Finding        `json:"findings"`
	Cycle          RemediationCycle `json:"remediation_cycle"`
}

type Dimension struct {
	Name         string `json:"name"`
	Score        int    `json:"score"`
	EvidenceCount int   `json:"evidence_count"`
}

type Finding struct {
	ID           string `json:"id"`
	Rule         string `json:"rule"`
	Dimension    string `json:"dimension"`
	Severity     string `json:"severity"`
	Path         string `json:"path"`
	Line         int    `json:"line,omitempty"`
	Evidence     string `json:"evidence"`
	Consequence  string `json:"consequence"`
	Instruction  string `json:"instruction"`
	Verification string `json:"verification"`
}

type RemediationCycle struct {
	Pass               int      `json:"pass"`
	MaxPasses          int      `json:"max_passes"`
	Status             string   `json:"status"`
	PersistentFindings []string `json:"persistent_findings,omitempty"`
	NextAction         string   `json:"next_action"`
}

type sourceFile struct {
	path       string
	relative   string
	lines      []string
	nonblank   int
	isTest     bool
}

// Assess scans root once. Supplying a previous report makes this the second and
// final remediation pass; remaining findings are escalated rather than looped.
func Assess(root string, previous *Report) (Report, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return Report{}, fmt.Errorf("resolve root: %w", err)
	}

	files, err := collectFiles(absoluteRoot)
	if err != nil {
		return Report{}, err
	}

	report := Report{
		SchemaVersion:  SchemaVersion,
		Root:           absoluteRoot,
		Interpretation: "The score prioritizes detected evidence; it is not proof that code is clean or correct.",
		Findings:       make([]Finding, 0),
	}
	for _, file := range files {
		report.Lines += file.nonblank
		if file.isTest {
			report.TestFiles++
		} else {
			report.SourceFiles++
		}
	}

	report.Findings = append(report.Findings, largeFileFindings(files)...)
	report.Findings = append(report.Findings, debtMarkerFindings(files)...)
	report.Findings = append(report.Findings, duplicateBlockFindings(files)...)
	if report.SourceFiles >= 3 && report.TestFiles == 0 {
		report.Findings = append(report.Findings, Finding{
			Rule: "tests.absent", Dimension: "tests", Severity: "high", Path: ".",
			Evidence: fmt.Sprintf("%d source files and no recognized test files", report.SourceFiles),
			Consequence: "Agents cannot prove behavior is preserved while changing the system.",
			Instruction: "Add black-box tests for the highest-risk public behavior before restructuring production code.",
			Verification: "Run the repository test command and rerun clean-code slop; at least one recognized test file must be present.",
		})
	}

	sort.Slice(report.Findings, func(i, j int) bool {
		if severityWeight(report.Findings[i].Severity) != severityWeight(report.Findings[j].Severity) {
			return severityWeight(report.Findings[i].Severity) > severityWeight(report.Findings[j].Severity)
		}
		if report.Findings[i].Path != report.Findings[j].Path {
			return report.Findings[i].Path < report.Findings[j].Path
		}
		if report.Findings[i].Line != report.Findings[j].Line {
			return report.Findings[i].Line < report.Findings[j].Line
		}
		return report.Findings[i].Rule < report.Findings[j].Rule
	})
	for index := range report.Findings {
		report.Findings[index].ID = fmt.Sprintf("SLOP-%03d", index+1)
	}
	report.Score, report.Dimensions = score(report.Lines, report.Findings)
	report.Cycle = cycle(previous, report.Findings)
	return report, nil
}

func Load(reader io.Reader) (Report, error) {
	var report Report
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return Report{}, fmt.Errorf("decode previous sloppiness report: %w", err)
	}
	if report.SchemaVersion != SchemaVersion {
		return Report{}, fmt.Errorf("unsupported previous report schema %q", report.SchemaVersion)
	}
	return report, nil
}

func LoadFile(path string) (Report, error) {
	file, err := os.Open(path)
	if err != nil {
		return Report{}, fmt.Errorf("open previous sloppiness report: %w", err)
	}
	defer file.Close()
	return Load(file)
}

func collectFiles(root string) ([]sourceFile, error) {
	var files []sourceFile
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && ignoredDirectories[entry.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !sourceExtensions[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if !utf8.Valid(content) || generated(content) {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		lines := scanLines(string(content))
		files = append(files, sourceFile{
			path: path, relative: filepath.ToSlash(relative), lines: lines,
			nonblank: countNonblank(lines), isTest: isTestPath(relative),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan source files: %w", err)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].relative < files[j].relative })
	return files, nil
}

func scanLines(content string) []string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func generated(content []byte) bool {
	head := strings.ToLower(string(content))
	if len(head) > 2048 {
		head = head[:2048]
	}
	return strings.Contains(head, "code generated") && strings.Contains(head, "do not edit")
}

func isTestPath(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	clean := "/" + strings.ToLower(filepath.ToSlash(path)) + "/"
	return strings.Contains(clean, "/test/") || strings.Contains(clean, "/tests/") ||
		strings.HasSuffix(name, "_test.go") || strings.Contains(name, ".test.") ||
		strings.Contains(name, ".spec.") || strings.HasPrefix(name, "test_") ||
		strings.HasSuffix(name, "tests.cs")
}

func countNonblank(lines []string) int {
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func largeFileFindings(files []sourceFile) []Finding {
	var findings []Finding
	for _, file := range files {
		if file.isTest || file.nonblank <= 500 {
			continue
		}
		findings = append(findings, Finding{
			Rule: "responsibility.large-file", Dimension: "responsibility", Severity: "high", Path: file.relative, Line: 1,
			Evidence: fmt.Sprintf("%d nonblank lines exceed the conservative 500-line review threshold", file.nonblank),
			Consequence: "Unrelated reasons to change are likely coupled, increasing review and regression cost.",
			Instruction: "Identify responsibilities that change for different reasons and extract one at a time without changing public behavior.",
			Verification: "Run existing tests after each extraction, then rerun clean-code slop; document an exception if the file must remain large.",
		})
	}
	return findings
}

func debtMarkerFindings(files []sourceFile) []Finding {
	var findings []Finding
	markers := []string{"TODO", "FIXME", "HACK"}
	for _, file := range files {
		for index, line := range file.lines {
			upper := strings.ToUpper(line)
			for _, marker := range markers {
				if !strings.Contains(upper, marker) {
					continue
				}
				findings = append(findings, Finding{
					Rule: "intent.unbounded-debt-marker", Dimension: "intent", Severity: "low", Path: file.relative, Line: index + 1,
					Evidence: fmt.Sprintf("%s marker: %s", marker, truncate(strings.TrimSpace(line), 120)),
					Consequence: "The future change has no explicit owner, boundary, or completion evidence.",
					Instruction: "Implement and remove the marker, or replace it with a tracked issue reference and a precise failure condition.",
					Verification: "Rerun clean-code slop and confirm the marker is gone or points to a bounded tracked issue.",
				})
				break
			}
		}
	}
	return findings
}

type blockOccurrence struct {
	path string
	line int
	text string
}

func duplicateBlockFindings(files []sourceFile) []Finding {
	const window = 6
	seen := make(map[string]blockOccurrence)
	reported := make(map[string]bool)
	var findings []Finding
	for _, file := range files {
		if file.isTest {
			continue
		}
		for index := 0; index+window <= len(file.lines); index++ {
			normalized, ok := normalizeBlock(file.lines[index : index+window])
			if !ok {
				continue
			}
			digestBytes := sha256.Sum256([]byte(normalized))
			digest := hex.EncodeToString(digestBytes[:])
			first, exists := seen[digest]
			if !exists {
				seen[digest] = blockOccurrence{path: file.relative, line: index + 1, text: normalized}
				continue
			}
			if first.path == file.relative || reported[digest] {
				continue
			}
			reported[digest] = true
			findings = append(findings, Finding{
				Rule: "duplication.exact-block", Dimension: "duplication", Severity: "medium", Path: file.relative, Line: index + 1,
				Evidence: fmt.Sprintf("6-line exact normalized block also appears at %s:%d", first.path, first.line),
				Consequence: "A single policy change can require synchronized edits in multiple places.",
				Instruction: "Name the shared policy and extract it behind the narrowest existing boundary; keep callers responsible for their distinct behavior.",
				Verification: "Run focused tests for both call sites, then rerun clean-code slop and confirm the duplicate block is absent.",
			})
		}
	}
	return findings
}

func normalizeBlock(lines []string) (string, bool) {
	var normalized []string
	substantive := 0
	length := 0
	for _, line := range lines {
		trimmed := strings.Join(strings.Fields(line), " ")
		if trimmed == "" || trimmed == "{" || trimmed == "}" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			return "", false
		}
		normalized = append(normalized, trimmed)
		length += len(trimmed)
		if strings.IndexFunc(trimmed, func(r rune) bool { return r >= 'A' && r <= 'z' }) >= 0 {
			substantive++
		}
	}
	return strings.Join(normalized, "\n"), substantive >= 4 && length >= 120
}

func score(lines int, findings []Finding) (int, []Dimension) {
	points := make(map[string]int)
	counts := make(map[string]int)
	for _, finding := range findings {
		points[finding.Dimension] += severityWeight(finding.Severity)
		counts[finding.Dimension]++
	}
	denominator := lines
	if denominator < 1000 {
		denominator = 1000
	}
	names := make([]string, 0, len(points))
	for name := range points {
		names = append(names, name)
	}
	sort.Strings(names)
	var dimensions []Dimension
	total := 0
	for _, name := range names {
		dimensionScore := clamp(points[name] * 4000 / denominator)
		dimensions = append(dimensions, Dimension{Name: name, Score: dimensionScore, EvidenceCount: counts[name]})
		total += points[name]
	}
	return clamp(total * 4000 / denominator), dimensions
}

func severityWeight(severity string) int {
	switch severity {
	case "high":
		return 8
	case "medium":
		return 4
	default:
		return 1
	}
}

func clamp(value int) int {
	if value > 100 {
		return 100
	}
	if value < 0 {
		return 0
	}
	return value
}

func cycle(previous *Report, findings []Finding) RemediationCycle {
	result := RemediationCycle{Pass: 1, MaxPasses: MaxPasses}
	if len(findings) == 0 {
		result.Status = "DONE"
		result.NextAction = "No repair batch is required. Preserve the report as evidence."
		return result
	}
	if previous == nil {
		result.Status = "REPAIR"
		result.NextAction = "Apply one bounded repair batch, run focused tests, then assess once more with --previous."
		return result
	}

	result.Pass = MaxPasses
	previousFingerprints := make(map[string]bool)
	for _, finding := range previous.Findings {
		previousFingerprints[fingerprint(finding)] = true
	}
	for _, finding := range findings {
		if previousFingerprints[fingerprint(finding)] {
			result.PersistentFindings = append(result.PersistentFindings, finding.ID)
		}
	}
	result.Status = "ESCALATE"
	result.NextAction = "Stop rewriting. Give this report, the previous report, tests, and diff to an independent agent or human for a policy decision."
	return result
}

func fingerprint(finding Finding) string {
	return strings.Join([]string{finding.Rule, finding.Path, fmt.Sprint(finding.Line), finding.Evidence}, "|")
}

func truncate(value string, maximum int) string {
	if len(value) <= maximum {
		return value
	}
	return value[:maximum-3] + "..."
}
