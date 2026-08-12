package providers

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

type ParsedArtifact struct {
	Metrics map[string]float64
}

func Parse(format string, content []byte) (ParsedArtifact, error) {
	if format == "" {
		format = "file"
	}
	switch format {
	case "file":
		return ParsedArtifact{Metrics: map[string]float64{"bytes": float64(len(content))}}, nil
	case "text":
		lines := 0
		if len(content) > 0 {
			lines = bytes.Count(content, []byte("\n"))
			if content[len(content)-1] != '\n' {
				lines++
			}
		}
		return ParsedArtifact{Metrics: map[string]float64{"bytes": float64(len(content)), "lines": float64(lines)}}, nil
	case "json":
		return parseJSON(content)
	case "sarif":
		return parseSARIF(content)
	case "xml":
		return parseXML(content)
	case "lcov":
		return parseLCOV(content)
	default:
		return ParsedArtifact{}, fmt.Errorf("unsupported artifact format %q", format)
	}
}

func parseJSON(content []byte) (ParsedArtifact, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return ParsedArtifact{}, fmt.Errorf("parse JSON artifact: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return ParsedArtifact{}, errors.New("parse JSON artifact: unexpected trailing value")
		}
		return ParsedArtifact{}, fmt.Errorf("parse JSON artifact: %w", err)
	}
	metrics := map[string]float64{}
	flattenJSON(metrics, "", value)
	return ParsedArtifact{Metrics: metrics}, nil
}

func flattenJSON(metrics map[string]float64, path string, value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			childPath := key
			if path != "" {
				childPath = path + "." + key
			}
			flattenJSON(metrics, childPath, child)
		}
	case json.Number:
		if number, err := typed.Float64(); err == nil && path != "" {
			metrics[path] = number
		}
	case []any:
		for index, child := range typed {
			flattenJSON(metrics, fmt.Sprintf("%s[%d]", path, index), child)
		}
	}
}

func parseSARIF(content []byte) (ParsedArtifact, error) {
	var document struct {
		Runs []struct {
			Results []struct {
				Level string `json:"level"`
			} `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(content, &document); err != nil {
		return ParsedArtifact{}, fmt.Errorf("parse SARIF artifact: %w", err)
	}
	if document.Runs == nil {
		return ParsedArtifact{}, errors.New("parse SARIF artifact: runs is required")
	}
	metrics := map[string]float64{"results.total": 0, "results.error": 0, "results.warning": 0, "results.note": 0}
	for _, run := range document.Runs {
		for _, result := range run.Results {
			metrics["results.total"]++
			level := strings.ToLower(result.Level)
			if level == "" {
				level = "warning"
			}
			if _, ok := metrics["results."+level]; ok {
				metrics["results."+level]++
			}
		}
	}
	return ParsedArtifact{Metrics: metrics}, nil
}

func parseXML(content []byte) (ParsedArtifact, error) {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	metrics := map[string]float64{}
	var path []string
	for {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return ParsedArtifact{}, fmt.Errorf("parse XML artifact: %w", err)
		}
		switch typed := token.(type) {
		case xml.StartElement:
			path = append(path, typed.Name.Local)
			for _, attribute := range typed.Attr {
				if number, err := strconv.ParseFloat(strings.TrimSpace(attribute.Value), 64); err == nil && finite(number) {
					metrics[strings.Join(path, ".")+".@"+attribute.Name.Local] = number
				} else if err == nil {
					return ParsedArtifact{}, fmt.Errorf("parse XML artifact: attribute %s must be finite", attribute.Name.Local)
				}
			}
		case xml.CharData:
			if number, err := strconv.ParseFloat(strings.TrimSpace(string(typed)), 64); err == nil && finite(number) && len(path) > 0 {
				metrics[strings.Join(path, ".")] = number
			} else if err == nil && !finite(number) {
				return ParsedArtifact{}, fmt.Errorf("parse XML artifact: value at %s must be finite", strings.Join(path, "."))
			}
		case xml.EndElement:
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
		}
	}
	return ParsedArtifact{Metrics: metrics}, nil
}

func parseLCOV(content []byte) (ParsedArtifact, error) {
	totals := map[string]float64{}
	for _, line := range strings.Split(string(content), "\n") {
		key, raw, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		switch key {
		case "LF", "LH", "BRF", "BRH", "FNF", "FNH":
			value, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return ParsedArtifact{}, fmt.Errorf("parse LCOV %s: %w", key, err)
			}
			if !finite(value) {
				return ParsedArtifact{}, fmt.Errorf("parse LCOV %s: value must be finite", key)
			}
			totals[key] += value
		}
	}
	if len(totals) == 0 {
		return ParsedArtifact{}, errors.New("parse LCOV artifact: no summary records found")
	}
	metrics := map[string]float64{
		"lines.found": totals["LF"], "lines.hit": totals["LH"],
		"branches.found": totals["BRF"], "branches.hit": totals["BRH"],
		"functions.found": totals["FNF"], "functions.hit": totals["FNH"],
	}
	percent(metrics, "lines.percent", totals["LH"], totals["LF"])
	percent(metrics, "branches.percent", totals["BRH"], totals["BRF"])
	percent(metrics, "functions.percent", totals["FNH"], totals["FNF"])
	return ParsedArtifact{Metrics: metrics}, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func percent(metrics map[string]float64, key string, hit, found float64) {
	if found > 0 {
		metrics[key] = hit / found * 100
	}
}
