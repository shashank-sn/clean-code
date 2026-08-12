package benchmark

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
)

const maxManifestBytes int64 = 10 << 20

type Manifest struct {
	SchemaVersion string `json:"schema_version"`
	Cases         []Case `json:"cases"`
}

type Case struct {
	ID       string   `json:"id"`
	Class    string   `json:"class"`
	Expected []string `json:"expected_findings"`
	Observed []string `json:"observed_findings"`
}

type Report struct {
	SchemaVersion  string  `json:"schema_version"`
	Cases          int     `json:"cases"`
	TruePositives  int     `json:"true_positives"`
	FalsePositives int     `json:"false_positives"`
	FalseNegatives int     `json:"false_negatives"`
	CleanControls  int     `json:"clean_controls"`
	CorrectSilence int     `json:"correct_silence"`
	Precision      float64 `json:"precision"`
	Recall         float64 `json:"recall"`
}

func Load(path string) (Manifest, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("inspect benchmark manifest: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Manifest{}, errors.New("benchmark manifest must be a regular file")
	}
	if info.Size() > maxManifestBytes {
		return Manifest{}, fmt.Errorf("benchmark manifest exceeds %d bytes", maxManifestBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return Manifest{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxManifestBytes+1))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse benchmark manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Manifest{}, errors.New("parse benchmark manifest: unexpected trailing JSON value")
	}
	if err := Validate(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func Validate(manifest Manifest) error {
	if manifest.SchemaVersion != "1.0.0" || len(manifest.Cases) == 0 {
		return errors.New("benchmark requires schema version 1.0.0 and at least one case")
	}
	ids := map[string]bool{}
	for _, testCase := range manifest.Cases {
		if testCase.ID == "" || ids[testCase.ID] {
			return errors.New("benchmark case ids must be present and unique")
		}
		ids[testCase.ID] = true
		if testCase.Class != "seeded-defect" && testCase.Class != "clean-control" {
			return fmt.Errorf("benchmark case %q has invalid class", testCase.ID)
		}
		if testCase.Class == "seeded-defect" && len(testCase.Expected) == 0 {
			return fmt.Errorf("seeded-defect case %q needs an expected finding", testCase.ID)
		}
		if testCase.Class == "clean-control" && len(testCase.Expected) != 0 {
			return fmt.Errorf("clean-control case %q cannot expect a finding", testCase.ID)
		}
		if duplicates(testCase.Expected) || duplicates(testCase.Observed) {
			return fmt.Errorf("benchmark case %q contains duplicate finding ids", testCase.ID)
		}
	}
	return nil
}

func Score(manifest Manifest) Report {
	report := Report{SchemaVersion: "1.0.0", Cases: len(manifest.Cases)}
	for _, testCase := range manifest.Cases {
		expected := set(testCase.Expected)
		observed := set(testCase.Observed)
		if testCase.Class == "clean-control" {
			report.CleanControls++
			if len(observed) == 0 {
				report.CorrectSilence++
			}
		}
		for finding := range observed {
			if expected[finding] {
				report.TruePositives++
			} else {
				report.FalsePositives++
			}
		}
		for finding := range expected {
			if !observed[finding] {
				report.FalseNegatives++
			}
		}
	}
	if total := report.TruePositives + report.FalsePositives; total > 0 {
		report.Precision = float64(report.TruePositives) / float64(total)
	}
	if total := report.TruePositives + report.FalseNegatives; total > 0 {
		report.Recall = float64(report.TruePositives) / float64(total)
	}
	return report
}

func duplicates(values []string) bool {
	copyOfValues := append([]string{}, values...)
	sort.Strings(copyOfValues)
	for index := 1; index < len(copyOfValues); index++ {
		if copyOfValues[index] == copyOfValues[index-1] {
			return true
		}
	}
	return false
}

func set(values []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[value] = true
	}
	return result
}
