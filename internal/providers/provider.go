package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"clean-code/internal/contracts"
)

const providerSchemaVersion = "1.0.0"
const maxProviderManifestBytes int64 = 1 << 20
const maxProviderResultBytes = 10 << 20

var providerID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

var providerCategories = map[string]bool{
	"complexity":          true,
	"mutation":            true,
	"acceptance":          true,
	"browser-qa":          true,
	"dependency-graph":    true,
	"architecture-render": true,
}

var artifactFormats = map[string]bool{
	"file": true, "json": true, "xml": true, "sarif": true, "lcov": true, "text": true,
}

// Spec is a portable, declarative provider contract. It names only a command
// and its evidence contract; execution remains subject to trusted policy.
type Spec struct {
	SchemaVersion string                `json:"schema_version"`
	ID            string                `json:"id"`
	Category      string                `json:"category"`
	Version       string                `json:"version"`
	Languages     []string              `json:"languages"`
	Command       contracts.CommandSpec `json:"command"`
	Input         ArtifactContract      `json:"input"`
	Output        ArtifactContract      `json:"output"`
	Availability  Availability          `json:"availability"`
	Policy        PolicyRequirements    `json:"policy"`
}

type ArtifactContract struct {
	Format    string                  `json:"format"`
	Schema    string                  `json:"schema"`
	Artifacts []NamedArtifactContract `json:"artifacts,omitempty"`
}

// NamedArtifactContract declares the individual evidence files expected inside
// a provider report. The parent contract remains the normalized report format.
type NamedArtifactContract struct {
	ID       string `json:"id"`
	Format   string `json:"format"`
	Schema   string `json:"schema"`
	Required bool   `json:"required"`
}

type Availability struct {
	Mode string `json:"mode"`
	Cost string `json:"cost"`
}

// PolicyRequirements are restrictive declarations. The portable core only
// accepts offline, non-installing providers and always requires trusted policy.
type PolicyRequirements struct {
	RequiresTrustedPolicy  bool `json:"requires_trusted_policy"`
	NetworkAccess          bool `json:"network_access"`
	MayInstallDependencies bool `json:"may_install_dependencies"`
}

// Result is a revision-bound projection of one provider execution.
type Result struct {
	SchemaVersion   string             `json:"schema_version"`
	ProviderID      string             `json:"provider_id"`
	ProviderVersion string             `json:"provider_version"`
	Category        string             `json:"category"`
	Revision        string             `json:"revision"`
	Status          contracts.Status   `json:"status"`
	StartedAt       time.Time          `json:"started_at"`
	DurationMS      int64              `json:"duration_ms"`
	Evidence        contracts.Evidence `json:"evidence,omitempty"`
}

func Load(path string) (Spec, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Spec{}, fmt.Errorf("inspect provider manifest: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Spec{}, errors.New("inspect provider manifest: input must be a regular file")
	}
	if info.Size() > maxProviderManifestBytes {
		return Spec{}, fmt.Errorf("inspect provider manifest: input exceeds %d bytes", maxProviderManifestBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return Spec{}, fmt.Errorf("open provider manifest: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxProviderManifestBytes+1))
	decoder.DisallowUnknownFields()
	var spec Spec
	if err := decoder.Decode(&spec); err != nil {
		return Spec{}, fmt.Errorf("parse provider manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Spec{}, errors.New("parse provider manifest: unexpected trailing JSON value")
		}
		return Spec{}, fmt.Errorf("parse provider manifest: %w", err)
	}
	if err := spec.Validate(); err != nil {
		return Spec{}, fmt.Errorf("validate provider manifest: %w", err)
	}
	return spec, nil
}

func (spec Spec) Validate() error {
	if spec.SchemaVersion != providerSchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", spec.SchemaVersion)
	}
	if !providerID.MatchString(spec.ID) {
		return errors.New("id must use lowercase letters, digits, and hyphens")
	}
	if !providerCategories[spec.Category] {
		return fmt.Errorf("unsupported category %q", spec.Category)
	}
	if strings.TrimSpace(spec.Version) == "" {
		return errors.New("version is required")
	}
	if len(spec.Languages) == 0 {
		return errors.New("at least one supported language is required")
	}
	seenLanguages := map[string]bool{}
	for _, language := range spec.Languages {
		language = strings.TrimSpace(language)
		if language == "" || seenLanguages[language] {
			return errors.New("languages must be non-empty and unique")
		}
		seenLanguages[language] = true
	}
	if err := spec.Command.Validate(); err != nil {
		return fmt.Errorf("command: %w", err)
	}
	if spec.Command.Category != "" && spec.Command.Category != spec.Category {
		return errors.New("command category must match provider category")
	}
	if err := spec.Input.Validate("input"); err != nil {
		return err
	}
	if err := spec.Output.Validate("output"); err != nil {
		return err
	}
	if spec.Availability.Mode != "local" && spec.Availability.Mode != "optional" {
		return errors.New("availability mode must be local or optional")
	}
	if spec.Availability.Cost != "none" && spec.Availability.Cost != "external" {
		return errors.New("availability cost must be none or external")
	}
	if !spec.Policy.RequiresTrustedPolicy {
		return errors.New("provider must require trusted policy")
	}
	if spec.Policy.NetworkAccess {
		return errors.New("network access is not permitted for providers")
	}
	if spec.Policy.MayInstallDependencies {
		return errors.New("providers may not install dependencies")
	}
	return nil
}

func (contract ArtifactContract) Validate(name string) error {
	if !artifactFormats[contract.Format] {
		return fmt.Errorf("%s artifact format %q is unsupported", name, contract.Format)
	}
	if strings.TrimSpace(contract.Schema) == "" {
		return fmt.Errorf("%s artifact schema is required", name)
	}
	seen := map[string]bool{}
	for index, artifact := range contract.Artifacts {
		if !providerID.MatchString(artifact.ID) || seen[artifact.ID] {
			return fmt.Errorf("%s artifact %d id must be unique and portable", name, index)
		}
		seen[artifact.ID] = true
		if !artifactFormats[artifact.Format] || strings.TrimSpace(artifact.Schema) == "" {
			return fmt.Errorf("%s artifact %q has an invalid format or schema", name, artifact.ID)
		}
	}
	return nil
}

func (result Result) Validate() error {
	if result.SchemaVersion != providerSchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", result.SchemaVersion)
	}
	if !providerID.MatchString(result.ProviderID) {
		return errors.New("provider_id is invalid")
	}
	if strings.TrimSpace(result.ProviderVersion) == "" || !providerCategories[result.Category] {
		return errors.New("provider version and category are required")
	}
	if strings.TrimSpace(result.Revision) == "" {
		return errors.New("revision is required")
	}
	if result.StartedAt.IsZero() || result.DurationMS < 0 {
		return errors.New("started_at and a non-negative duration_ms are required")
	}
	check := contracts.CheckResult{
		SchemaVersion: result.SchemaVersion, CheckID: result.ProviderID, Category: result.Category,
		Provider: result.ProviderID, Status: result.Status, Revision: result.Revision,
		StartedAt: result.StartedAt, DurationMS: result.DurationMS, Evidence: result.Evidence,
	}
	if err := check.Validate(); err != nil {
		return fmt.Errorf("result: %w", err)
	}
	if (result.Status == contracts.StatusPass || result.Status == contracts.StatusFail) && len(result.Evidence.ArtifactDetails) == 0 {
		return errors.New("pass and fail results require revision-bound artifact evidence")
	}
	for index, artifact := range result.Evidence.ArtifactDetails {
		if artifact.Revision != result.Revision || ((result.Status == contracts.StatusPass || result.Status == contracts.StatusFail) && !artifact.Fresh) {
			return fmt.Errorf("artifact %d revision does not match provider result", index)
		}
	}
	return nil
}

// ValidateResult binds a normalized provider result to its declared output contract.
func ValidateResult(spec Spec, result Result) error {
	if err := spec.Validate(); err != nil {
		return err
	}
	if err := result.Validate(); err != nil {
		return err
	}
	if result.ProviderID != spec.ID || result.ProviderVersion != spec.Version || result.Category != spec.Category {
		return errors.New("provider result does not match manifest id, version, and category")
	}
	if result.Status != contracts.StatusPass && result.Status != contracts.StatusFail {
		return nil
	}
	provided := map[string]bool{}
	for _, artifact := range result.Evidence.ArtifactDetails {
		provided[artifact.Schema] = true
	}
	if !provided[spec.Output.Schema] {
		return fmt.Errorf("provider result is missing output schema %q", spec.Output.Schema)
	}
	for _, expected := range spec.Output.Artifacts {
		if expected.Required && !provided[expected.Schema] {
			return fmt.Errorf("provider result is missing required artifact %q", expected.ID)
		}
	}
	return nil
}

// ParseResult decodes a bounded provider output only after its revision-bound
// evidence contract has validated. Callers must preserve unavailable and stale
// states rather than translating them into PASS.
func ParseResult(content []byte) (Result, error) {
	if len(content) > maxProviderResultBytes {
		return Result{}, fmt.Errorf("parse provider result: input exceeds %d bytes", maxProviderResultBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var result Result
	if err := decoder.Decode(&result); err != nil {
		return Result{}, fmt.Errorf("parse provider result: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Result{}, errors.New("parse provider result: unexpected trailing JSON value")
		}
		return Result{}, fmt.Errorf("parse provider result: %w", err)
	}
	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf("validate provider result: %w", err)
	}
	return result, nil
}

func (result Result) CheckResult(required bool) contracts.CheckResult {
	return contracts.CheckResult{
		SchemaVersion: result.SchemaVersion, CheckID: result.ProviderID, Category: result.Category,
		Provider: result.ProviderID, Status: result.Status, Required: required, Revision: result.Revision,
		StartedAt: result.StartedAt, DurationMS: result.DurationMS, Evidence: result.Evidence,
	}
}

func FixturePath(category string) string {
	return filepath.ToSlash(filepath.Join("harness", "providers", category, "provider.json"))
}
