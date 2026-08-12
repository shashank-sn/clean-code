package releasecontract

import (
	"strings"
	"testing"
	"time"
)

func validBinding() Binding {
	now:=time.Date(2026,8,12,0,0,0,0,time.UTC)
	return Binding{
		SchemaVersion:"1.0.0",Repository:"owner/repo",BaseRevision:"base",FinalRevision:"final",RequirementDigest:"requirements",ChangeSetDigest:"change",PolicyRevision:"policy",
		PolicyGates:PolicyGates{PolicyRevision:"policy",RequiredTests:[]string{"acceptance"},RequiredReviews:[]string{"independent"},RequiredDecisions:[]string{"RELEASE_RISK"}},
		ChangedPaths:[]string{"change.go"},
		Tests:[]Test{{ID:"acceptance",RequirementIDs:[]string{"R1"},Revision:"final",Status:"PASS",ArtifactDigest:"artifact",ActorRunID:"test-run",StartedAt:now,FinishedAt:now}},
		Reviews:[]Review{{ID:"independent",ReviewerRunID:"review-run",ReviewerContextID:"review-context",ChangeAuthorRunID:"build-run",ChangeAuthorContextID:"build-context",BaseRevision:"base",FinalRevision:"final",RequirementDigest:"requirements",ChangeSetDigest:"change",PolicyRevision:"policy",ReviewedPaths:[]string{"change.go"},Status:"PASS"}},
		Decisions:[]Decision{{Kind:"RELEASE_RISK",Authority:"human",Subject:"final",Status:"APPROVED"}},
	}
}

func TestValidateAcceptsPolicyRequiredRevisionBoundEvidence(t *testing.T){if err:=validBinding().Validate(time.Time{});err!=nil{t.Fatal(err)}}
func TestValidateRejectsEmptyEvidence(t *testing.T){b:=validBinding();b.Tests=nil;b.Reviews=nil;b.Decisions=nil;if err:=b.Validate(time.Time{});err==nil{t.Fatal("expected empty evidence rejection")}}
func TestValidateDoesNotLetEvidenceDeclareItselfOptional(t *testing.T){b:=validBinding();b.Tests[0].Status="PLANNED";if err:=b.Validate(time.Time{});err==nil||!strings.Contains(err.Error(),"not an executed pass"){t.Fatalf("expected policy gate rejection, got %v",err)}}
func TestValidateRejectsPolicyWithNoRequiredGates(t *testing.T){b:=validBinding();b.PolicyGates.RequiredTests=nil;if err:=b.Validate(time.Time{});err==nil||!strings.Contains(err.Error(),"no required test"){t.Fatalf("expected empty policy rejection, got %v",err)}}
func TestValidateRejectsDuplicateAndUnknownEvidence(t *testing.T){b:=validBinding();b.Tests=append(b.Tests,b.Tests[0]);if err:=b.Validate(time.Time{});err==nil||!strings.Contains(err.Error(),"duplicate"){t.Fatalf("expected duplicate rejection, got %v",err)};b=validBinding();b.Tests[0].Status="SUCCESS";if err:=b.Validate(time.Time{});err==nil||!strings.Contains(err.Error(),"unknown status"){t.Fatalf("expected status rejection, got %v",err)}}
func TestValidateRejectsStalePartialAndCorrelatedReview(t *testing.T){b:=validBinding();b.Reviews[0].FinalRevision="old";if err:=b.Validate(time.Time{});err==nil||!strings.Contains(err.Error(),"stale"){t.Fatalf("expected stale rejection, got %v",err)};b=validBinding();b.Reviews[0].ReviewedPaths=nil;if err:=b.Validate(time.Time{});err==nil||!strings.Contains(err.Error(),"changed scope"){t.Fatalf("expected scope rejection, got %v",err)};b=validBinding();b.Reviews[0].ReviewerContextID=b.Reviews[0].ChangeAuthorContextID;if err:=b.Validate(time.Time{});err==nil||!strings.Contains(err.Error(),"not independent"){t.Fatalf("expected correlation rejection, got %v",err)}}
func TestValidateRejectsUntypedAndNonWaivableExceptions(t *testing.T){b:=validBinding();b.Exceptions=[]Exception{{Kind:"CORRECTNESS",Approver:"human",Subject:"failure",Rationale:"ship",Scope:[]string{"R1"},ExpiresAt:time.Now().Add(time.Hour)}};if err:=b.Validate(time.Now());err==nil||!strings.Contains(err.Error(),"not waivable"){t.Fatalf("expected exception rejection, got %v",err)};b=validBinding();b.Decisions[0].Kind="CODE_APPROVAL";if err:=b.Validate(time.Now());err==nil||!strings.Contains(err.Error(),"human decision"){t.Fatalf("expected decision rejection, got %v",err)}}
