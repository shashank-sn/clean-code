package study

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Task struct {
	ID     string   `json:"id"`
	Model  string   `json:"model"`
	Tools  []string `json:"tools"`
	Limit  int      `json:"limit"`
	Oracle string   `json:"oracle"`
}
type Manifest struct {
	SchemaVersion         string `json:"schema_version"`
	StudyID               string `json:"study_id"`
	Repository            string `json:"repository"`
	Revision              string `json:"revision"`
	CaseCorpusDigest      string `json:"case_corpus_digest"`
	OracleCorpusDigest    string `json:"oracle_corpus_digest"`
	ModelConfigDigest     string `json:"model_config_digest"`
	PreregistrationDigest string `json:"preregistration_digest"`
	MinimumPairs          int    `json:"minimum_pairs"`
	Tasks                 []Task `json:"tasks"`
}
type Outcome struct {
	Repository         string   `json:"repository"`
	Revision           string   `json:"revision"`
	RunID              string   `json:"run_id"`
	RequestDigest      string   `json:"request_digest"`
	ArtifactDigest     string   `json:"artifact_digest"`
	StartedAt          string   `json:"started_at"`
	FinishedAt         string   `json:"finished_at"`
	TaskID             string   `json:"task_id"`
	Arm                string   `json:"arm"`
	Model              string   `json:"model"`
	Tools              []string `json:"tools"`
	Limit              int      `json:"limit"`
	Oracle             string   `json:"oracle"`
	CaseCorpusDigest   string   `json:"case_corpus_digest"`
	OracleCorpusDigest string   `json:"oracle_corpus_digest"`
	ModelConfigDigest  string   `json:"model_config_digest"`
	Status             string   `json:"status"`
	FalsePositives     int      `json:"false_positives"`
	CorrectSilence     bool     `json:"correct_silence"`
}
type Results struct {
	SchemaVersion  string    `json:"schema_version"`
	StudyID        string    `json:"study_id"`
	ManifestDigest string    `json:"manifest_digest"`
	Outcomes       []Outcome `json:"outcomes"`
}
type Report struct {
	SchemaVersion  string    `json:"schema_version"`
	Pairs          int       `json:"pairs"`
	ExecutionValid bool      `json:"execution_valid"`
	ClaimAllowed   bool      `json:"claim_allowed"`
	Outcomes       []Outcome `json:"outcomes"`
	Limitations    []string  `json:"limitations"`
}

var repoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
var revisionPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func Digest(body []byte) string { sum := sha256.Sum256(body); return hex.EncodeToString(sum[:]) }
func ParseManifest(body []byte) (Manifest, error) {
	var m Manifest
	if err := strict(body, &m); err != nil {
		return m, err
	}
	if m.SchemaVersion != "1.0.0" || !repoPattern.MatchString(m.Repository) || !revisionPattern.MatchString(m.Revision) || !digestPattern.MatchString(m.CaseCorpusDigest) || !digestPattern.MatchString(m.OracleCorpusDigest) || !digestPattern.MatchString(m.ModelConfigDigest) || !digestPattern.MatchString(m.PreregistrationDigest) {
		return m, errors.New("study manifest identity is invalid")
	}
	return m, nil
}
func ParseResults(body []byte) (Results, error) {
	var r Results
	err := strict(body, &r)
	return r, err
}
func LoadManifest(path string) (Manifest, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	return ParseManifest(body)
}
func VerifyInputs(m Manifest, cases, oracle, config, preregistration []byte) error {
	if Digest(cases) != m.CaseCorpusDigest {
		return errors.New("case corpus digest does not match preregistration")
	}
	if Digest(oracle) != m.OracleCorpusDigest {
		return errors.New("oracle corpus digest does not match preregistration")
	}
	if Digest(config) != m.ModelConfigDigest {
		return errors.New("model config digest does not match preregistration")
	}
	if Digest(preregistration) != m.PreregistrationDigest {
		return errors.New("preregistration digest does not match manifest")
	}
	return nil
}
func strict(body []byte, target any) error {
	d := json.NewDecoder(strings.NewReader(string(body)))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	var x any
	if err := d.Decode(&x); !errors.Is(err, io.EOF) {
		return errors.New("unexpected trailing JSON")
	}
	return nil
}
func Score(m Manifest, r Results) (Report, error) {
	if r.SchemaVersion != "1.0.0" || m.StudyID == "" || m.StudyID != r.StudyID || m.MinimumPairs < 1 {
		return Report{}, errors.New("study metadata is invalid")
	}
	tasks := map[string]Task{}
	for _, t := range m.Tasks {
		if t.ID == "" || t.Model == "" || len(t.Tools) == 0 || t.Limit < 1 || t.Oracle == "" || tasks[t.ID].ID != "" {
			return Report{}, errors.New("preregistered task is invalid")
		}
		tasks[t.ID] = t
	}
	type pairIDs struct{ runs, artifacts map[string]bool }
	ids := map[string]*pairIDs{}
	arms := map[string]map[string]bool{}
	failed := false
	for _, o := range r.Outcomes {
		start, e1 := time.Parse(time.RFC3339, o.StartedAt)
		finish, e2 := time.Parse(time.RFC3339, o.FinishedAt)
		task := tasks[o.TaskID]
		if task.ID == "" || o.Repository != m.Repository || o.Revision != m.Revision || o.RunID == "" || !digestPattern.MatchString(o.RequestDigest) || !digestPattern.MatchString(o.ArtifactDigest) || e1 != nil || e2 != nil || finish.Before(start) || o.FalsePositives < 0 || o.Model != task.Model || o.Limit != task.Limit || o.Oracle != task.Oracle || o.CaseCorpusDigest != m.CaseCorpusDigest || o.OracleCorpusDigest != m.OracleCorpusDigest || o.ModelConfigDigest != m.ModelConfigDigest || !same(o.Tools, task.Tools) || !one(o.Arm, "control", "workflow") || !one(o.Status, "PASS", "FAIL", "TIMEOUT") {
			return Report{}, errors.New("raw outcome is invalid")
		}
		if ids[o.TaskID] == nil {
			ids[o.TaskID] = &pairIDs{map[string]bool{}, map[string]bool{}}
		}
		if ids[o.TaskID].runs[o.RunID] || ids[o.TaskID].artifacts[o.ArtifactDigest] {
			return Report{}, errors.New("paired arms must use distinct run and artifact identities")
		}
		ids[o.TaskID].runs[o.RunID] = true
		ids[o.TaskID].artifacts[o.ArtifactDigest] = true
		if o.Status != "PASS" {
			failed = true
		}
		if arms[o.TaskID] == nil {
			arms[o.TaskID] = map[string]bool{}
		}
		if arms[o.TaskID][o.Arm] {
			return Report{}, errors.New("duplicate study arm")
		}
		arms[o.TaskID][o.Arm] = true
	}
	pairs := 0
	for id := range tasks {
		if arms[id]["control"] && arms[id]["workflow"] {
			pairs++
		}
	}
	executionValid := pairs >= m.MinimumPairs && pairs == len(tasks) && !failed
	limits := []string{"descriptive pilot has no preregistered superiority threshold; comparative performance claims are blocked"}
	if !executionValid {
		limits = []string{"insufficient, unbalanced, or failing paired cases; execution qualification and performance claims are blocked"}
	}
	sort.Slice(r.Outcomes, func(i, j int) bool {
		if r.Outcomes[i].TaskID == r.Outcomes[j].TaskID {
			return r.Outcomes[i].Arm < r.Outcomes[j].Arm
		}
		return r.Outcomes[i].TaskID < r.Outcomes[j].TaskID
	})
	return Report{"1.0.0", pairs, executionValid, false, r.Outcomes, limits}, nil
}
func same(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func one(v string, a ...string) bool {
	for _, x := range a {
		if v == x {
			return true
		}
	}
	return false
}
