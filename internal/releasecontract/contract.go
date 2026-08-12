package releasecontract

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type PolicyGates struct {
	PolicyRevision    string   `json:"policy_revision"`
	RequiredTests     []string `json:"required_tests"`
	RequiredReviews   []string `json:"required_reviews"`
	RequiredDecisions []string `json:"required_decisions"`
}

type Binding struct {
	SchemaVersion string `json:"schema_version"`
	Repository string `json:"repository"`
	BaseRevision string `json:"base_revision"`
	FinalRevision string `json:"final_revision"`
	RequirementDigest string `json:"requirement_digest"`
	ChangeSetDigest string `json:"change_set_digest"`
	PolicyRevision string `json:"policy_revision"`
	PolicyGates PolicyGates `json:"policy_gates"`
	ChangedPaths []string `json:"changed_paths"`
	Tests []Test `json:"tests"`
	Reviews []Review `json:"reviews"`
	Decisions []Decision `json:"decisions"`
	Exceptions []Exception `json:"exceptions,omitempty"`
}

type Test struct {
	ID string `json:"id"`
	RequirementIDs []string `json:"requirement_ids"`
	Revision string `json:"revision"`
	Status string `json:"status"`
	ArtifactDigest string `json:"artifact_digest,omitempty"`
	ActorRunID string `json:"actor_run_id,omitempty"`
	StartedAt time.Time `json:"started_at,omitempty"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
}

type Review struct {
	ID string `json:"id"`
	ReviewerRunID string `json:"reviewer_run_id"`
	ReviewerContextID string `json:"reviewer_context_id"`
	ChangeAuthorRunID string `json:"change_author_run_id"`
	ChangeAuthorContextID string `json:"change_author_context_id"`
	BaseRevision string `json:"base_revision"`
	FinalRevision string `json:"final_revision"`
	RequirementDigest string `json:"requirement_digest"`
	ChangeSetDigest string `json:"change_set_digest"`
	PolicyRevision string `json:"policy_revision"`
	ReviewedPaths []string `json:"reviewed_paths"`
	Status string `json:"status"`
}

type Decision struct {
	Kind string `json:"kind"`
	Authority string `json:"authority"`
	Subject string `json:"subject"`
	Status string `json:"status"`
}

type Exception struct {
	Kind string `json:"kind"`
	Approver string `json:"approver"`
	Subject string `json:"subject"`
	Rationale string `json:"rationale"`
	Scope []string `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
}

func Load(path string) (Binding, error) {
	file, err := os.Open(path)
	if err != nil { return Binding{}, err }
	defer file.Close()
	var binding Binding
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&binding); err != nil { return Binding{}, err }
	var trailing any
	if err := decoder.Decode(&trailing); err == nil { return Binding{}, errors.New("unexpected trailing JSON value") }
	return binding, nil
}

func (b Binding) Validate(now time.Time) error {
	required := []string{b.SchemaVersion,b.Repository,b.BaseRevision,b.FinalRevision,b.RequirementDigest,b.ChangeSetDigest,b.PolicyRevision,b.PolicyGates.PolicyRevision}
	for _, value := range required { if strings.TrimSpace(value)=="" { return errors.New("release binding metadata is incomplete") } }
	if b.SchemaVersion!="1.0.0" { return errors.New("unsupported release binding schema") }
	if b.PolicyGates.PolicyRevision!=b.PolicyRevision { return errors.New("policy gates belong to another policy revision") }
	if len(b.ChangedPaths)==0 { return errors.New("changed path scope is empty") }
	if err:=uniqueRequired("test",b.PolicyGates.RequiredTests); err!=nil{return err}
	if err:=uniqueRequired("review",b.PolicyGates.RequiredReviews); err!=nil{return err}
	if err:=uniqueRequired("decision",b.PolicyGates.RequiredDecisions); err!=nil{return err}

	tests:=map[string]Test{}
	for _, test:=range b.Tests {
		if test.ID=="" || tests[test.ID].ID!="" { return errors.New("test evidence has an empty or duplicate id") }
		tests[test.ID]=test
		if !oneOf(test.Status,"PLANNED","NOT_RUN","NOT_AVAILABLE","INAPPLICABLE","FAIL","PASS"){return fmt.Errorf("test %q has unknown status",test.ID)}
		if test.Revision!=b.FinalRevision{return fmt.Errorf("test %q belongs to another revision",test.ID)}
		if test.Status=="PASS"&&(len(test.RequirementIDs)==0||test.ArtifactDigest==""||test.ActorRunID==""||test.StartedAt.IsZero()||test.FinishedAt.Before(test.StartedAt)){return fmt.Errorf("test %q lacks executed evidence",test.ID)}
	}
	for _, id:=range b.PolicyGates.RequiredTests { if tests[id].Status!="PASS"{return fmt.Errorf("required test %q is not an executed pass",id)} }

	reviews:=map[string]Review{}
	for _, review:=range b.Reviews {
		if review.ID==""||reviews[review.ID].ID!=""{return errors.New("review evidence has an empty or duplicate id")}
		reviews[review.ID]=review
		if !oneOf(review.Status,"PASS","FAIL","INCOMPLETE"){return fmt.Errorf("review %q has unknown status",review.ID)}
		if review.BaseRevision!=b.BaseRevision||review.FinalRevision!=b.FinalRevision||review.RequirementDigest!=b.RequirementDigest||review.ChangeSetDigest!=b.ChangeSetDigest||review.PolicyRevision!=b.PolicyRevision{return fmt.Errorf("review %q is stale or covers another contract",review.ID)}
		if review.ReviewerRunID==""||review.ReviewerContextID==""||review.ChangeAuthorRunID==""||review.ChangeAuthorContextID==""||review.ReviewerRunID==review.ChangeAuthorRunID||review.ReviewerContextID==review.ChangeAuthorContextID{return fmt.Errorf("review %q is not independent",review.ID)}
		if !sameSet(review.ReviewedPaths,b.ChangedPaths){return fmt.Errorf("review %q does not cover the changed scope",review.ID)}
	}
	for _, id:=range b.PolicyGates.RequiredReviews { if reviews[id].Status!="PASS"{return fmt.Errorf("required review %q is incomplete",id)} }

	decisions:=map[string]Decision{}
	for _, decision:=range b.Decisions {
		if !oneOf(decision.Kind,"INTENT","POLICY","RELEASE_RISK")||decisions[decision.Kind].Kind!=""||decision.Authority==""||decision.Subject==""||!oneOf(decision.Status,"APPROVED","REJECTED"){return errors.New("human decision is invalid or duplicated")}
		decisions[decision.Kind]=decision
	}
	for _, kind:=range b.PolicyGates.RequiredDecisions { if decisions[kind].Status!="APPROVED"{return fmt.Errorf("required %s decision is not approved",kind)} }
	for _, exception:=range b.Exceptions {
		if !oneOf(exception.Kind,"PROCESS","AVAILABILITY","TEMPORARY_POLICY"){return errors.New("exception kind is not waivable")}
		if exception.Approver==""||exception.Subject==""||exception.Rationale==""||len(exception.Scope)==0||!exception.ExpiresAt.After(now){return errors.New("exception is invalid or expired")}
	}
	return nil
}

func uniqueRequired(kind string, values []string) error {
	if len(values)==0{return fmt.Errorf("policy defines no required %s gates",kind)}
	seen:=map[string]bool{}
	for _, value:=range values {if value==""||seen[value]{return fmt.Errorf("policy has an empty or duplicate required %s gate",kind)};seen[value]=true}
	return nil
}
func sameSet(a,b []string)bool{if len(a)!=len(b){return false};seen:=map[string]bool{};for _,v:=range a{if v==""||seen[v]{return false};seen[v]=true};for _,v:=range b{if !seen[v]{return false}};return true}
func oneOf(value string,allowed ...string)bool{for _,candidate:=range allowed{if value==candidate{return true}};return false}
