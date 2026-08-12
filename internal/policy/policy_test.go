package policy

import (
	"reflect"
	"testing"

	"clean-code/internal/contracts"
)

func TestCompareReportsStableSortedDeltasWithoutValues(t *testing.T) {
	trusted := []contracts.CommandSpec{{ID: "test", Executable: "go", Args: []string{"test", "./..."}, Required: true}}
	proposed := []contracts.CommandSpec{
		{ID: "lint", Executable: "lint", Env: map[string]string{"API_TOKEN": "do-not-print"}},
		{ID: "test", Executable: "go", Args: []string{"test", "./pkg"}, Required: false},
	}
	want := []string{
		`command "lint" was added`,
		`command "test" changed arguments`,
		`command "test" changed required`,
	}
	if got := Compare(trusted, proposed); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected deltas:\nwant: %#v\n got: %#v", want, got)
	}
}

func TestCompareReportsRemovedCommand(t *testing.T) {
	trusted := []contracts.CommandSpec{{ID: "test", Executable: "go"}}
	if got := Compare(trusted, nil); !reflect.DeepEqual(got, []string{`command "test" was removed`}) {
		t.Fatalf("unexpected deltas: %#v", got)
	}
}

func TestCompareTreatsNilAndEmptyCollectionsAsEquivalent(t *testing.T) {
	trusted := []contracts.CommandSpec{{ID: "test", Executable: "go"}}
	proposed := []contracts.CommandSpec{{
		ID: "test", Executable: "go", Args: []string{}, AllowedExitCodes: []int{},
		Env: map[string]string{}, Artifacts: []contracts.ArtifactSpec{},
	}}
	if got := Compare(trusted, proposed); len(got) != 0 {
		t.Fatalf("expected no semantic delta, got %#v", got)
	}
}
