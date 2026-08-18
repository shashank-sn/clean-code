package benchmark

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const maxFullFlowManifestBytes int64 = 2 << 20

type FullFlowManifest struct {
	SchemaVersion      string          `json:"schema_version"`
	TaskID             string          `json:"task_id"`
	TaskPath           string          `json:"task_path"`
	Outcomes           []FullFlowEntry `json:"outcomes"`
	Rubric             []RubricItem    `json:"rubric"`
	ReviewerRubric     []RubricItem    `json:"reviewer_rubric,omitempty"`
	Reviewer           *ReviewerInput  `json:"reviewer,omitempty"`
}

type FullFlowEntry struct {
	WorkflowID string `json:"workflow_id"`
	Label      string `json:"label"`
	PackageDir string `json:"package_dir"`
}

type RubricItem struct {
	ID          string  `json:"id"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description,omitempty"`
}

type ReviewerInput struct {
	Model   string             `json:"model,omitempty"`
	Scores  map[string]float64 `json:"scores,omitempty"`
	Notes   string             `json:"notes,omitempty"`
	Blinded bool               `json:"blinded,omitempty"`
}

type FullFlowReport struct {
	SchemaVersion string              `json:"schema_version"`
	TaskID        string              `json:"task_id"`
	TaskPath      string              `json:"task_path"`
	Outcomes      []FullFlowOutcome   `json:"outcomes"`
	Rubric        []RubricItem        `json:"rubric"`
	Winner        string              `json:"winner"`
	Summary       string              `json:"summary"`
	Reviewer      *ReviewerInput      `json:"reviewer,omitempty"`
}

type FullFlowOutcome struct {
	WorkflowID     string             `json:"workflow_id"`
	Label          string             `json:"label"`
	PackageDir     string             `json:"package_dir"`
	TestsPassed    bool               `json:"tests_passed"`
	TestOutput     string             `json:"test_output,omitempty"`
	Metrics        CodeMetrics        `json:"metrics"`
	AutoScores     map[string]float64 `json:"auto_scores"`
	ReviewerScores map[string]float64 `json:"reviewer_scores,omitempty"`
	AutoTotal      float64            `json:"auto_total"`
	ReviewerTotal  float64            `json:"reviewer_total,omitempty"`
	Combined       float64            `json:"combined_score"`
}

type CodeMetrics struct {
	ProductionLines int `json:"production_lines"`
	TestLines       int `json:"test_lines"`
	TestFunctions   int `json:"test_functions"`
	ProductionFuncs int `json:"production_funcs"`
	MaxFuncLines    int `json:"max_func_lines"`
	AvgFuncLines    int `json:"avg_func_lines"`
	HasFuzzTest     bool `json:"has_fuzz_test"`
}

var funcDeclPattern = regexp.MustCompile(`^func\s+(\([^)]*\)\s+)?(\w+)\s*\(`)

func LoadFullFlowManifest(path string) (FullFlowManifest, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return FullFlowManifest{}, fmt.Errorf("inspect full-flow manifest: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return FullFlowManifest{}, errors.New("full-flow manifest must be a regular file")
	}
	if info.Size() > maxFullFlowManifestBytes {
		return FullFlowManifest{}, fmt.Errorf("full-flow manifest exceeds %d bytes", maxFullFlowManifestBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return FullFlowManifest{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxFullFlowManifestBytes+1))
	decoder.DisallowUnknownFields()
	var manifest FullFlowManifest
	if err := decoder.Decode(&manifest); err != nil {
		return FullFlowManifest{}, fmt.Errorf("parse full-flow manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return FullFlowManifest{}, errors.New("parse full-flow manifest: unexpected trailing JSON value")
	}
	if err := ValidateFullFlowManifest(manifest); err != nil {
		return FullFlowManifest{}, err
	}
	return manifest, nil
}

func ValidateFullFlowManifest(manifest FullFlowManifest) error {
	if manifest.SchemaVersion != "1.0.0" || manifest.TaskID == "" {
		return errors.New("full-flow manifest requires schema version 1.0.0 and task_id")
	}
	if len(manifest.Outcomes) < 2 || len(manifest.Rubric) == 0 {
		return errors.New("full-flow manifest requires at least two outcomes and rubric items")
	}
	seen := map[string]bool{}
	for _, entry := range manifest.Outcomes {
		if entry.WorkflowID == "" || seen[entry.WorkflowID] {
			return errors.New("outcome workflow_id values must be present and unique")
		}
		seen[entry.WorkflowID] = true
		if strings.TrimSpace(entry.PackageDir) == "" {
			return fmt.Errorf("outcome %q missing package_dir", entry.WorkflowID)
		}
	}
	seenRubric := map[string]bool{}
	for _, item := range manifest.Rubric {
		if item.ID == "" || seenRubric[item.ID] {
			return errors.New("rubric ids must be present and unique")
		}
		seenRubric[item.ID] = true
		if item.Weight <= 0 {
			return fmt.Errorf("rubric item %q weight must be positive", item.ID)
		}
	}
	return nil
}

func RunFullFlow(repoRoot string, manifest FullFlowManifest) (FullFlowReport, error) {
	report := FullFlowReport{
		SchemaVersion: "1.0.0",
		TaskID:        manifest.TaskID,
		TaskPath:      manifest.TaskPath,
		Rubric:        append([]RubricItem{}, manifest.Rubric...),
		Reviewer:      manifest.Reviewer,
	}
	for _, entry := range manifest.Outcomes {
		pkgDir := entry.PackageDir
		if !filepath.IsAbs(pkgDir) {
			pkgDir = filepath.Join(repoRoot, pkgDir)
		}
		metrics, err := analyzePackage(pkgDir)
		if err != nil {
			return FullFlowReport{}, fmt.Errorf("analyze %s: %w", entry.WorkflowID, err)
		}
		passed, output := runPackageTests(pkgDir)
		auto := scoreMetrics(metrics, passed, manifest.Rubric)
		autoTotal := weightedTotal(auto, manifest.Rubric)
		reviewerMap := reviewerScoresFor(manifest.Reviewer, entry.WorkflowID)
		reviewerTotal := weightedTotal(reviewerMap, manifest.ReviewerRubric)
		combined := autoTotal
		if reviewerTotal > 0 {
			combined = (autoTotal + reviewerTotal) / 2.0
		}
		report.Outcomes = append(report.Outcomes, FullFlowOutcome{
			WorkflowID:     entry.WorkflowID,
			Label:          entry.Label,
			PackageDir:     entry.PackageDir,
			TestsPassed:    passed,
			TestOutput:     trimOutput(output),
			Metrics:        metrics,
			AutoScores:     auto,
			ReviewerScores: reviewerMap,
			AutoTotal:      autoTotal,
			ReviewerTotal:  reviewerTotal,
			Combined:       combined,
		})
	}
	report.Winner = pickWinner(report.Outcomes)
	report.Summary = buildFullFlowSummary(report)
	return report, nil
}

func analyzePackage(dir string) (CodeMetrics, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return CodeMetrics{}, err
	}
	metrics := CodeMetrics{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return CodeMetrics{}, err
		}
		lines := strings.Split(string(body), "\n")
		isTest := strings.HasSuffix(entry.Name(), "_test.go")
		if isTest {
			metrics.TestLines += countCodeLines(lines)
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "func Test") {
					metrics.TestFunctions++
				}
				if strings.HasPrefix(trimmed, "func Fuzz") {
					metrics.HasFuzzTest = true
				}
				if strings.Contains(trimmed, "{name:") {
					metrics.TestFunctions++
				}
			}
			continue
		}
		metrics.ProductionLines += countCodeLines(lines)
		funcLines := measureFunctions(lines)
		for _, length := range funcLines {
			metrics.ProductionFuncs++
			if length > metrics.MaxFuncLines {
				metrics.MaxFuncLines = length
			}
		}
	}
	if metrics.ProductionFuncs > 0 {
		total := 0
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			body, _ := os.ReadFile(filepath.Join(dir, entry.Name()))
			for _, length := range measureFunctions(strings.Split(string(body), "\n")) {
				total += length
			}
		}
		metrics.AvgFuncLines = total / metrics.ProductionFuncs
	}
	return metrics, nil
}

func countCodeLines(lines []string) int {
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		count++
	}
	return count
}

func measureFunctions(lines []string) []int {
	var lengths []int
	inFunc := false
	depth := 0
	start := 0
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inFunc && funcDeclPattern.MatchString(trimmed) && !strings.HasPrefix(trimmed, "func Test") && !strings.HasPrefix(trimmed, "func Fuzz") {
			inFunc = true
			start = index
			depth = 0
		}
		if !inFunc {
			continue
		}
		depth += strings.Count(line, "{")
		depth -= strings.Count(line, "}")
		if inFunc && depth == 0 && strings.Contains(line, "}") {
			lengths = append(lengths, countCodeLines(lines[start:index+1]))
			inFunc = false
		}
	}
	return lengths
}

func runPackageTests(dir string) (bool, string) {
	cmd := exec.Command("go", "test", "-race", "./...")
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	output := stdout.String() + stderr.String()
	return err == nil, output
}

func scoreMetrics(metrics CodeMetrics, testsPassed bool, rubric []RubricItem) map[string]float64 {
	scores := map[string]float64{}
	for _, item := range rubric {
		switch item.ID {
		case "tests_pass":
			scores[item.ID] = boolScore(testsPassed)
		case "test_coverage_breadth":
			scores[item.ID] = clamp01(float64(metrics.TestFunctions) / 6.0)
		case "function_size":
			if metrics.MaxFuncLines <= 15 {
				scores[item.ID] = 1.0
			} else if metrics.MaxFuncLines <= 30 {
				scores[item.ID] = 0.6
			} else {
				scores[item.ID] = 0.3
			}
		case "decomposition":
			if metrics.ProductionFuncs >= 3 && metrics.MaxFuncLines <= 20 {
				scores[item.ID] = 1.0
			} else if metrics.ProductionFuncs >= 2 {
				scores[item.ID] = 0.7
			} else {
				scores[item.ID] = 0.4
			}
		case "simplicity":
			ratio := float64(metrics.ProductionLines)
			if ratio <= 35 {
				scores[item.ID] = 1.0
			} else if ratio <= 55 {
				scores[item.ID] = 0.7
			} else {
				scores[item.ID] = 0.4
			}
		case "safety_hardening":
			scores[item.ID] = boolScore(metrics.HasFuzzTest)
		default:
			scores[item.ID] = 0.5
		}
	}
	return scores
}

func reviewerScoresFor(reviewer *ReviewerInput, workflowID string) map[string]float64 {
	if reviewer == nil || len(reviewer.Scores) == 0 {
		return nil
	}
	// When blinded, reviewer scores are keyed by label A/B in manifest file
	// mapped to workflow ids via separate fields - use workflow_id suffix keys
	prefix := workflowID + "."
	result := map[string]float64{}
	for key, value := range reviewer.Scores {
		if strings.HasPrefix(key, prefix) {
			result[strings.TrimPrefix(key, prefix)] = clamp01(value)
		}
	}
	if len(result) == 0 {
		// fallback: shared reviewer dimension keys apply to both weighted externally
		return nil
	}
	return result
}

func combineScores(auto, reviewer map[string]float64, rubric []RubricItem) float64 {
	return weightedTotal(auto, rubric)
}

func weightedTotal(scores map[string]float64, rubric []RubricItem) float64 {
	if len(rubric) == 0 || len(scores) == 0 {
		return 0
	}
	totalWeight := 0.0
	totalScore := 0.0
	for _, item := range rubric {
		score, ok := scores[item.ID]
		if !ok {
			continue
		}
		totalWeight += item.Weight
		totalScore += clamp01(score) * item.Weight
	}
	if totalWeight == 0 {
		return 0
	}
	return totalScore / totalWeight
}

func pickWinner(outcomes []FullFlowOutcome) string {
	if len(outcomes) == 0 {
		return ""
	}
	winner := outcomes[0].WorkflowID
	best := outcomes[0].Combined
	for _, outcome := range outcomes[1:] {
		if outcome.Combined > best {
			best = outcome.Combined
			winner = outcome.WorkflowID
		}
	}
	return winner
}

func buildFullFlowSummary(report FullFlowReport) string {
	if len(report.Outcomes) < 2 {
		return "insufficient outcomes"
	}
	a, b := report.Outcomes[0], report.Outcomes[1]
	if len(report.Outcomes) > 1 && report.Outcomes[1].Combined > report.Outcomes[0].Combined {
		a, b = report.Outcomes[1], report.Outcomes[0]
	}
	return fmt.Sprintf(
		"%s wins full-flow benchmark (%.2f vs %.2f). %s tests=%v max_func_lines=%d tests=%d; %s tests=%v max_func_lines=%d tests=%d.",
		report.Winner,
		maxCombined(report.Outcomes),
		minCombined(report.Outcomes),
		a.WorkflowID, a.TestsPassed, a.Metrics.MaxFuncLines, a.Metrics.TestFunctions,
		b.WorkflowID, b.TestsPassed, b.Metrics.MaxFuncLines, b.Metrics.TestFunctions,
	)
}

func maxCombined(outcomes []FullFlowOutcome) float64 {
	best := outcomes[0].Combined
	for _, o := range outcomes[1:] {
		if o.Combined > best {
			best = o.Combined
		}
	}
	return best
}

func minCombined(outcomes []FullFlowOutcome) float64 {
	low := outcomes[0].Combined
	for _, o := range outcomes[1:] {
		if o.Combined < low {
			low = o.Combined
		}
	}
	return low
}

func boolScore(ok bool) float64 {
	if ok {
		return 1.0
	}
	return 0.0
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func trimOutput(output string) string {
	const limit = 2048
	if len(output) <= limit {
		return output
	}
	return output[:limit] + "...(truncated)"
}
