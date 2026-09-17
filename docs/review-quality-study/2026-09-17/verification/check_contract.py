"""Black-box acceptance probes derived from the preregistered review requirements."""
import copy
import json
import pathlib
import shutil
import subprocess
import sys
import tempfile

from jsonschema import Draft202012Validator
from referencing import Registry, Resource

repo = pathlib.Path(sys.argv[1]).resolve()
binary = pathlib.Path(sys.argv[2]).resolve()
output = pathlib.Path(sys.argv[3]).resolve()
revision = "b" * 40
base = {
    "schema_version": "2.0.0", "base_revision": "a" * 40,
    "revision": revision, "change_author": "implementer", "reviewer": "independent",
    "scope": ["src/service.go"], "requirements": ["R1: preserve tenant isolation"],
    "coverage": [{"path": "src/service.go", "status": "REVIEWED", "evidence": "src/service.go:10-30 and its caller src/handler.go:18"}],
    "dimensions": [
        {"id": name, "status": "PASS", "evidence": ["check:inspection"]}
        for name in ["correctness", "integration", "tests", "failure_modes", "security", "maintainability"]
    ],
    "checks": [{"id": "inspection", "kind": "inspection", "required": True,
                "status": "PASS", "revision": revision, "source": "static causal trace of changed service and caller",
                "artifact": "inspection.txt", "sha256": "a" * 64}],
    "limitations": [], "completion": "COMPLETE", "findings": [],
}

cases = []
def case(name, mutate, expected_pass, schema_expected=None):
    value = copy.deepcopy(base)
    mutate(value)
    cases.append((name, value, expected_pass, schema_expected))

case("complete-static-review", lambda x: None, True, True)
case("missing-findings-is-not-correct-silence", lambda x: x.pop("findings"), False, False)
case("null-findings-is-not-correct-silence", lambda x: x.update(findings=None), False, False)
case("empty-coverage", lambda x: x.update(coverage=[]), False, False)
case("missing-one-dimension", lambda x: x["dimensions"].pop(), False)
case("duplicate-dimension", lambda x: x["dimensions"].append(copy.deepcopy(x["dimensions"][0])), False)
case("duplicate-scope", lambda x: x["scope"].append(x["scope"][0]), False, False)
case("coverage-outside-scope", lambda x: x["coverage"][0].update(path="other.go"), False)
case("unreviewed-scope-with-no-findings", lambda x: x["coverage"][0].update(status="NOT_REVIEWED", evidence="", reason="caller unavailable"), False)
case("stale-green-check", lambda x: x["checks"][0].update(revision="c" * 40), False)
case("invalid-hash", lambda x: x["checks"][0].update(sha256="looks-good"), False, False)
case("required-field-is-not-optional", lambda x: x["checks"][0].pop("required"), False, False)
case("null-required-field", lambda x: x["checks"][0].update(required=None), False, False)
case("unknown-dimension-status", lambda x: x["dimensions"][0].update(status="PROBABLY_FINE"), False, False)
case("reasonless-not-applicable", lambda x: x["dimensions"][0].update(status="NOT_APPLICABLE", evidence=[]), False, False)
case("not-applicable-with-reason", lambda x: x["dimensions"][4].update(status="NOT_APPLICABLE", evidence=[], reason="read-only formatting change, no trust boundary affected"), True, True)
case("optional-unavailable-check", lambda x: x["checks"].append({"id":"browser", "kind":"browser", "required":False, "status":"NOT_AVAILABLE", "source":"browser interaction check", "unavailable_reason":"no UI or browser required for this library review"}), True, True)
case("required-unavailable-check", lambda x: x["checks"].append({"id":"integration", "kind":"test", "required":True, "status":"NOT_AVAILABLE", "source":"integration service check", "unavailable_reason":"required service unavailable"}), False)
case("declared-incomplete", lambda x: x.update(completion="INCOMPLETE", limitations=["source boundary unavailable"]), False)
case("unknown-top-level-field", lambda x: x.update(automatic_approval=True), False, False)
case("unknown-nested-field", lambda x: x["coverage"][0].update(automatic_approval=True), False, False)
case("same-author", lambda x: x.update(reviewer=x["change_author"]), False)
case("same-author-whitespace", lambda x: x.update(reviewer=x["change_author"] + " "), False)

blocking = {
    "id":"F1", "severity":"BLOCKING", "file":"src/service.go", "line":12,
    "behavior":"cross-tenant cache reuse", "evidence":"cache key omits tenant at line 12",
    "consequence":"tenant B receives tenant A data", "confidence":"HIGH",
    "fix":"scope the cache key by tenant", "disposition":"DISMISSED",
    "resolution_reason":"the caller already incorporates tenant into the key",
}
case("dismissal-without-recheck", lambda x: x.update(findings=[copy.deepcopy(blocking)]), False)
def open_blocker(value):
    finding = copy.deepcopy(blocking)
    finding["disposition"] = "OPEN"
    finding.pop("resolution_reason")
    value["findings"] = [finding]
case("open-blocker-never-passes", open_blocker, False, True)
def negative_line(value):
    open_blocker(value)
    value["findings"][0].update(severity="ADVISORY", line=-1)
case("negative-finding-location", negative_line, False, False)
def zero_line(value):
    negative_line(value)
    value["findings"][0]["line"] = 0
case("explicit-zero-finding-location", zero_line, False, False)
def omitted_line(value):
    negative_line(value)
    value["findings"][0].pop("line")
case("optional-finding-location-omitted", omitted_line, True, True)
def duplicate_cause(value):
    open_blocker(value)
    value["findings"][0]["severity"] = "IMPROVEMENT"
    other = copy.deepcopy(value["findings"][0])
    other["id"] = "F2"
    value["findings"].append(other)
case("duplicate-causal-finding", duplicate_cause, False)
def resolved(value):
    finding = copy.deepcopy(blocking)
    finding["resolution_check"] = "inspection"
    value["findings"] = [finding]
case("disproven-blocker-with-final-recheck", resolved, True, True)
def accepted(value):
    resolved(value)
    value["findings"][0]["disposition"] = "ACCEPTED_RISK"
case("accepted-risk-does-not-resolve-blocker", accepted, False)

schema_root = repo / "harness/schemas"
registry = Registry()
for path in schema_root.glob("*.schema.json"):
    schema = json.loads(path.read_text())
    Draft202012Validator.check_schema(schema)
    registry = registry.with_resource(schema.get("$id", path.as_uri()), Resource.from_contents(schema))
validator = Draft202012Validator(json.loads((schema_root / "review-input.schema.json").read_text()), registry=registry)
results = []
with tempfile.TemporaryDirectory(prefix="review-contract-") as directory:
    for name, value, expected, schema_expected in cases:
        path = pathlib.Path(directory) / f"{name}.json"
        path.write_text(json.dumps(value))
        execution = subprocess.run([str(binary), "review", "--input", str(path)], cwd=repo, capture_output=True, text=True)
        schema_errors = list(validator.iter_errors(value))
        schema_pass = not schema_errors
        runtime_pass = execution.returncode == 0
        passed = runtime_pass == expected and (schema_expected is None or schema_pass == schema_expected)
        results.append({"case":name,"pass":passed,"runtime_pass":runtime_pass,"expected_runtime_pass":expected,
                        "schema_pass":schema_pass,"expected_schema_pass":schema_expected,
                        "runtime":execution.stdout.strip() or execution.stderr.strip(),
                        "schema_errors":[e.message for e in schema_errors[:3]]})

report = {"scope":"independent synthetic CLI and Draft 2020-12 schema acceptance probes; no semantic review certification",
          "passed":sum(x["pass"] for x in results),"total":len(results),"results":results}
audit_results = []
for name, mutate, should_pass in [
    ("audit-valid-v2", lambda x: None, True),
    ("audit-null-required", lambda x: x["checks"][0].update(required=None), False),
    ("audit-missing-required", lambda x: x["checks"][0].pop("required"), False),
    ("audit-unknown-nested-field", lambda x: x["checks"][0].update(approve=True), False),
]:
    with tempfile.TemporaryDirectory(prefix="review-audit-") as directory:
        root = pathlib.Path(directory)
        shutil.copytree(repo / "tests/fixtures/audit", root, dirs_exist_ok=True)
        value = copy.deepcopy(base)
        value.update(revision="abc", change_author="author")
        value["checks"][0]["revision"] = "abc"
        mutate(value)
        (root / "review.json").write_text(json.dumps(value))
        execution = subprocess.run([str(binary), "audit", "--input", str(root / "audit-input.json"),
                                    "--output", str(root / "receipt.json")], cwd=repo, capture_output=True, text=True)
        passed = (execution.returncode == 0) == should_pass
        audit_results.append({"case":name,"pass":passed,"expected_pass":should_pass,"exit_code":execution.returncode,
                              "output":execution.stdout.strip() or execution.stderr.strip()})
report["audit_results"] = audit_results
report["passed"] += sum(x["pass"] for x in audit_results)
report["total"] += len(audit_results)
output.write_text(json.dumps(report,indent=2)+"\n")
print(json.dumps({"passed":report["passed"],"total":report["total"],"failures":[x["case"] for x in results+audit_results if not x["pass"]]}))
sys.exit(0 if report["passed"] == report["total"] else 1)
