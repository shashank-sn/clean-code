import { createHash } from "node:crypto";
import { readFile, readdir, realpath, stat, writeFile } from "node:fs/promises";
import { dirname, isAbsolute, join, relative, resolve, sep } from "node:path";

const [studyRootArgument, oraclePathArgument] = process.argv.slice(2);
if (!studyRootArgument || !oraclePathArgument || process.argv.length !== 4) {
  throw new Error("usage: node clean-code-score-held-out-v1.mjs STUDY_ROOT ORACLE_PATH");
}

const studyRoot = await realpath(studyRootArgument);
const repositoryRoot = resolve(studyRoot, "..", "..", "..");
const oraclePath = await realpath(oraclePathArgument);
const manifestPath = join(repositoryRoot, "harness", "studies", "held-out-v1.json");
const casesPath = join(studyRoot, "cases.json");
const preregistrationPath = join(studyRoot, "preregistration.json");
const configPath = join(studyRoot, "model-config.json");
const attemptManifestPath = join(studyRoot, "execution-attempt.json");
const journalPath = join(studyRoot, "execution-journal.ndjson");
const rawRoot = join(studyRoot, "raw");
const resultsPath = join(studyRoot, "results.json");
const reportPath = join(studyRoot, "scoring-report.json");

const sha256 = body => createHash("sha256").update(body).digest("hex");
const fail = message => { throw new Error(message); };
const requireValue = (condition, message) => { if (!condition) fail(message); };
const parseJSON = (body, label) => {
  try {
    requireValue(!(body[0] === 0xef && body[1] === 0xbb && body[2] === 0xbf), `${label} must not contain a UTF-8 BOM`);
    const text = new TextDecoder("utf-8", { fatal: true }).decode(body);
    return JSON.parse(text);
  } catch (error) {
    fail(`${label} is not valid JSON: ${error.message}`);
  }
};
const exactKeys = (value, keys, label) => {
  requireValue(value && typeof value === "object" && !Array.isArray(value), `${label} must be an object`);
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  requireValue(JSON.stringify(actual) === JSON.stringify(expected), `${label} has unexpected or missing fields`);
};
const validDigest = value => typeof value === "string" && /^[0-9a-f]{64}$/.test(value);
const validDate = value => typeof value === "string" && Number.isFinite(Date.parse(value));
const normalizedTokens = description => new Set(
  description.toLowerCase().replace(/[^a-z0-9]+/g, " ").trim().split(/ +/).filter(Boolean),
);
const sumBy = (items, key) => items.reduce((total, item) => total + item[key], 0);
const pairsEqual = (a, b) => JSON.stringify(a) === JSON.stringify(b);
const armsFor = index => index % 2 === 0 ? ["control", "workflow"] : ["workflow", "control"];
const safeArtifactPath = async value => {
  requireValue(typeof value === "string" && value !== "" && !isAbsolute(value), "artifact_path must be a nonempty relative path");
  const candidate = resolve(studyRoot, value);
  const rel = relative(studyRoot, candidate);
  requireValue(rel !== "" && rel !== ".." && !rel.startsWith(`..${sep}`), `artifact_path escapes study root: ${value}`);
  const resolved = await realpath(candidate);
  const resolvedRel = relative(studyRoot, resolved);
  requireValue(resolvedRel !== "" && resolvedRel !== ".." && !resolvedRel.startsWith(`..${sep}`), `artifact_path resolves outside study root: ${value}`);
  const info = await stat(resolved);
  requireValue(info.isFile(), `artifact_path is not a regular file: ${value}`);
  return resolved;
};

const [manifestBody, casesBody, preregistrationBody, configBody, oracleBody, attemptManifestBody, journalBody] = await Promise.all([
  readFile(manifestPath),
  readFile(casesPath),
  readFile(preregistrationPath),
  readFile(configPath),
  readFile(oraclePath),
  readFile(attemptManifestPath),
  readFile(journalPath),
]);
const manifest = parseJSON(manifestBody, "study manifest");
const corpus = parseJSON(casesBody, "case corpus");
const preregistration = parseJSON(preregistrationBody, "preregistration");
const config = parseJSON(configBody, "model config");
const oracle = parseJSON(oracleBody, "oracle");
const attempt = parseJSON(attemptManifestBody, "execution attempt");

requireValue(manifest.schema_version === "1.0.0" && manifest.study_id === "held-out-v1", "unexpected study manifest identity");
requireValue(manifest.repository === preregistration.repository, "repository mismatch between manifest and preregistration");
requireValue(manifest.revision === preregistration.target_revision, "revision mismatch between manifest and preregistration");
requireValue(sha256(casesBody) === manifest.case_corpus_digest, "case corpus digest does not match manifest");
requireValue(manifest.case_corpus_digest === preregistration.commitments?.case_corpus_sha256, "case corpus commitment mismatch");
requireValue(sha256(configBody) === manifest.model_config_digest, "model config digest does not match manifest");
requireValue(validDigest(manifest.oracle_corpus_digest), "manifest oracle digest is invalid");
requireValue(sha256(oracleBody) === manifest.oracle_corpus_digest, "oracle digest does not match manifest");
requireValue(manifest.oracle_corpus_digest === preregistration.commitments?.oracle_scoring_sha256, "oracle digest does not match preregistration");
requireValue(oracle.case_corpus_sha256 === manifest.case_corpus_digest, "oracle is bound to a different case corpus");
requireValue(oracle.study_id === manifest.study_id && oracle.oracle_id === "review-oracle-v1", "unexpected oracle identity");
requireValue(oracle.algorithm_id === "structured-semantic-token-v1" && oracle.algorithm_version === "1.0.0", "unsupported oracle scoring algorithm");
requireValue(preregistration.scoring?.algorithm_id === oracle.algorithm_id && preregistration.scoring?.algorithm_version === oracle.algorithm_version, "oracle algorithm was not preregistered");
requireValue(config.model === "gpt-5-2025-08-07" && config.store === false && Array.isArray(config.tools) && config.tools.length === 0, "model config is outside the preregistered execution envelope");

requireValue(Array.isArray(corpus.cases) && corpus.cases.length === 10, "case corpus must contain exactly ten cases");
requireValue(Array.isArray(manifest.tasks) && manifest.tasks.length === 10, "manifest must contain exactly ten tasks");
requireValue(Array.isArray(oracle.cases) && oracle.cases.length === 10, "oracle must contain exactly ten cases");
const cases = new Map(corpus.cases.map(item => [item.task_id, item]));
const tasks = new Map(manifest.tasks.map(item => [item.id, item]));
const oracleCases = new Map(oracle.cases.map(item => [item.task_id, item]));
requireValue(cases.size === 10 && tasks.size === 10 && oracleCases.size === 10, "duplicate case, task, or oracle ID");
for (let number = 1; number <= 10; number += 1) {
  const taskID = `held-out-review-${String(number).padStart(2, "0")}`;
  const task = tasks.get(taskID);
  requireValue(cases.has(taskID) && task && oracleCases.has(taskID), `missing binding for ${taskID}`);
  requireValue(task.model === config.model && task.limit === config.max_output_words, `task config mismatch for ${taskID}`);
  requireValue(pairsEqual(task.tools, ["none"]) && task.oracle === oracle.oracle_id, `task tools/oracle mismatch for ${taskID}`);
}

const manifestDigest = sha256(manifestBody);
exactKeys(attempt, ["schema_version", "study_id", "attempt_id", "manifest_digest", "created_at", "schedule"], "execution attempt");
requireValue(attempt.schema_version === "1.0.0" && attempt.study_id === manifest.study_id, "execution attempt identity mismatch");
requireValue(typeof attempt.attempt_id === "string" && /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(attempt.attempt_id), "execution attempt ID is invalid");
requireValue(attempt.manifest_digest === manifestDigest && validDate(attempt.created_at), "execution attempt commitment is invalid");
const expectedSchedule = corpus.cases.flatMap((item, index) => armsFor(index).map(arm => ({ task_id: item.task_id, arm })));
requireValue(Array.isArray(attempt.schedule) && attempt.schedule.length === expectedSchedule.length, "execution attempt must reserve all 20 ordered slots");
for (const [index, slot] of attempt.schedule.entries()) {
  exactKeys(slot, ["ordinal", "task_id", "arm"], `execution attempt slot ${index + 1}`);
  const expected = expectedSchedule[index];
  requireValue(slot.ordinal === index + 1 && slot.task_id === expected.task_id && slot.arm === expected.arm, `execution attempt schedule mismatch at ordinal ${index + 1}`);
}

const journalText = journalBody.toString("utf8");
requireValue(journalText.endsWith("\n"), "execution journal must end with LF");
const journalLines = journalText.slice(0, -1).split("\n");
requireValue(journalLines.length === 20 && journalLines.every(line => line !== ""), "execution journal must contain exactly 20 nonempty records");
const journalRuns = journalLines.map((line, index) => parseJSON(Buffer.from(line), `journal line ${index + 1}`));
const seenRunKeys = new Set();
const seenRunIDs = new Set();
const seenArtifacts = new Set();
const referencedArtifactPaths = new Set();
const outcomes = [];
const privateScores = [];

for (const [index, run] of journalRuns.entries()) {
  exactKeys(run, ["attempt_id", "ordinal", "task_id", "arm", "started_at", "finished_at", "http_status", "response_id", "response_model", "response_status", "execution_status", "output_words", "artifact_path", "artifact_digest"], `journal line ${index + 1}`);
  const slot = attempt.schedule[index];
  requireValue(run.attempt_id === attempt.attempt_id && run.ordinal === slot.ordinal && run.task_id === slot.task_id && run.arm === slot.arm, `journal order or attempt mismatch on line ${index + 1}`);
  requireValue(tasks.has(run.task_id) && (run.arm === "control" || run.arm === "workflow"), `invalid run identity on journal line ${index + 1}`);
  const runKey = `${run.task_id}|${run.arm}`;
  requireValue(!seenRunKeys.has(runKey), `duplicate terminal run for ${runKey}`);
  seenRunKeys.add(runKey);
  requireValue(validDate(run.started_at) && validDate(run.finished_at) && Date.parse(run.finished_at) >= Date.parse(run.started_at), `invalid timestamps for ${runKey}`);
  requireValue(run.execution_status === "completed" && run.http_status === 200 && run.response_status === "completed", `non-completed terminal run for ${runKey}`);
  requireValue(run.response_model === config.model && typeof run.response_id === "string" && run.response_id !== "", `response identity mismatch for ${runKey}`);
  requireValue(!seenRunIDs.has(run.response_id), `duplicate OpenAI response ID: ${run.response_id}`);
  seenRunIDs.add(run.response_id);
  requireValue(validDigest(run.artifact_digest) && !seenArtifacts.has(run.artifact_digest), `invalid or duplicate artifact digest for ${runKey}`);
  seenArtifacts.add(run.artifact_digest);
  requireValue(Number.isInteger(run.output_words) && run.output_words >= 0 && run.output_words <= config.max_output_words, `output word limit mismatch for ${runKey}`);

  const artifactPath = await safeArtifactPath(run.artifact_path);
  requireValue(!referencedArtifactPaths.has(artifactPath), `duplicate artifact path for ${runKey}`);
  referencedArtifactPaths.add(artifactPath);
  const artifactBody = await readFile(artifactPath);
  requireValue(sha256(artifactBody) === run.artifact_digest, `raw artifact digest mismatch for ${runKey}`);
  const response = parseJSON(artifactBody, `raw artifact for ${runKey}`);
  requireValue(response.id === run.response_id && response.model === run.response_model && response.status === "completed", `raw response identity mismatch for ${runKey}`);
  requireValue(response.error == null && response.incomplete_details == null, `raw response is incomplete or errored for ${runKey}`);
  requireValue(Array.isArray(response.tools) && response.tools.length === 0, `raw response used tools for ${runKey}`);
  const metadata = response.metadata;
  exactKeys(metadata, ["study_id", "attempt_id", "ordinal", "task_id", "arm", "case_corpus_digest", "model_config_digest"], `raw response metadata for ${runKey}`);
  requireValue(metadata.study_id === manifest.study_id && metadata.attempt_id === attempt.attempt_id && metadata.ordinal === String(run.ordinal), `raw response metadata attempt mismatch for ${runKey}`);
  requireValue(metadata.task_id === run.task_id && metadata.arm === run.arm, `raw response metadata identity mismatch for ${runKey}`);
  requireValue(metadata.case_corpus_digest === manifest.case_corpus_digest && metadata.model_config_digest === manifest.model_config_digest, `raw response metadata digest mismatch for ${runKey}`);

  const textBlocks = (Array.isArray(response.output) ? response.output : [])
    .filter(item => item?.role === "assistant")
    .flatMap(item => Array.isArray(item.content) ? item.content : [])
    .filter(item => item?.type === "output_text");
  requireValue(textBlocks.length === 1 && typeof textBlocks[0].text === "string", `raw response must contain exactly one assistant output_text for ${runKey}`);
  const rawText = textBlocks[0].text;
  const computedWords = rawText.trim() === "" ? 0 : rawText.trim().split(/\s+/u).length;
  requireValue(computedWords === run.output_words && computedWords <= config.max_output_words, `raw/journal word count mismatch for ${runKey}`);

  let text = rawText;
  let syntaxValid = true;
  if (text.startsWith("\uFEFF")) syntaxValid = false;
  if (text.endsWith("\n")) text = text.slice(0, -1);
  if (text.includes("\r") || text.includes("\n")) syntaxValid = false;
  let parsed = { kind: "invalid" };
  if (syntaxValid && text === "NO_FINDING") {
    parsed = { kind: "silence" };
  } else if (syntaxValid) {
    const fields = text.split("|");
    if (fields.length === 4 && fields[0] === "FINDING" && ["HIGH", "MEDIUM", "LOW"].includes(fields[1]) && /^[1-9][0-9]*$/.test(fields[2]) && fields[3] !== "") {
      parsed = { kind: "finding", severity: fields[1], line: Number.parseInt(fields[2], 10), description: fields[3] };
    }
  }

  const expected = oracleCases.get(run.task_id);
  const actionable = expected.correct_silence === false;
  let pass = false;
  let truePositive = 0;
  let falsePositive = 0;
  let falseNegative = 0;
  let correctSilence = 0;
  let invalidResponse = 0;
  if (parsed.kind === "invalid") {
    invalidResponse = 1;
    if (actionable) falseNegative = 1;
  } else if (!actionable) {
    pass = parsed.kind === "silence";
    correctSilence = pass ? 1 : 0;
    falsePositive = parsed.kind === "finding" ? 1 : 0;
  } else if (parsed.kind === "finding") {
    const [minimumLine, maximumLine] = expected.accepted_line_range;
    const tokens = normalizedTokens(parsed.description);
    const semanticMatch = expected.required_semantic_tokens.every(group => group.some(token => tokens.has(token)));
    pass = parsed.line >= minimumLine && parsed.line <= maximumLine && semanticMatch;
    truePositive = pass ? 1 : 0;
    falseNegative = pass ? 0 : 1;
    falsePositive = pass ? 0 : 1;
  } else {
    falseNegative = 1;
  }

  const task = tasks.get(run.task_id);
  outcomes.push({
    repository: manifest.repository,
    revision: manifest.revision,
    run_id: run.response_id,
    artifact_digest: run.artifact_digest,
    started_at: run.started_at,
    finished_at: run.finished_at,
    task_id: run.task_id,
    arm: run.arm,
    model: task.model,
    tools: task.tools,
    limit: task.limit,
    oracle: task.oracle,
    case_corpus_digest: manifest.case_corpus_digest,
    oracle_corpus_digest: manifest.oracle_corpus_digest,
    model_config_digest: manifest.model_config_digest,
    status: pass ? "PASS" : "FAIL",
    false_positives: falsePositive,
    correct_silence: correctSilence === 1,
  });
  privateScores.push({ run_key: runKey, pass: pass ? 1 : 0, fail: pass ? 0 : 1, true_positive: truePositive, false_positive: falsePositive, false_negative: falseNegative, correct_silence: correctSilence, invalid_response: invalidResponse });
}

requireValue(seenRunKeys.size === 20 && seenRunIDs.size === 20 && seenArtifacts.size === 20, "study must contain exactly 20 unique terminal runs, response IDs, and artifacts");
const rawEntries = await readdir(rawRoot, { withFileTypes: true });
requireValue(rawEntries.length === 20 && rawEntries.every(entry => entry.isFile()), "raw directory must contain exactly 20 regular files");
for (const entry of rawEntries) {
  const rawPath = await realpath(join(rawRoot, entry.name));
  requireValue(referencedArtifactPaths.has(rawPath), `unreferenced raw artifact: ${entry.name}`);
}
for (const taskID of tasks.keys()) {
  requireValue(seenRunKeys.has(`${taskID}|control`) && seenRunKeys.has(`${taskID}|workflow`), `missing paired arm for ${taskID}`);
}
outcomes.sort((a, b) => a.task_id.localeCompare(b.task_id) || a.arm.localeCompare(b.arm));

const results = { schema_version: "1.0.0", study_id: manifest.study_id, manifest_digest: manifestDigest, outcomes };
const aggregateForArm = arm => {
  const selected = privateScores.filter(score => score.run_key.endsWith(`|${arm}`));
  return {
    runs: selected.length,
    passed: sumBy(selected, "pass"),
    failed: sumBy(selected, "fail"),
    true_positives: sumBy(selected, "true_positive"),
    false_positives: sumBy(selected, "false_positive"),
    false_negatives: sumBy(selected, "false_negative"),
    correct_silences: sumBy(selected, "correct_silence"),
    invalid_responses: sumBy(selected, "invalid_response"),
  };
};
const control = aggregateForArm("control");
const workflow = aggregateForArm("workflow");
const claimAllowed = control.failed === 0 && workflow.failed === 0;
const report = {
  schema_version: "1.0.0",
  study_id: manifest.study_id,
  scoring_algorithm: { id: oracle.algorithm_id, version: oracle.algorithm_version },
  commitments: {
    manifest_digest: manifestDigest,
    case_corpus_digest: manifest.case_corpus_digest,
    oracle_corpus_digest: manifest.oracle_corpus_digest,
    model_config_digest: manifest.model_config_digest,
  },
  execution: { attempt_id: attempt.attempt_id, terminal_runs: outcomes.length, unique_response_ids: seenRunIDs.size, unique_artifacts: seenArtifacts.size, complete_pairs: tasks.size },
  arms: { control, workflow },
  claim_allowed: claimAllowed,
  limitations: claimAllowed ? [] : ["One or more raw-artifact-derived outcomes failed; performance claims remain blocked."],
};

await writeFile(resultsPath, `${JSON.stringify(results, null, 2)}\n`, { flag: "wx", mode: 0o600 });
await writeFile(reportPath, `${JSON.stringify(report, null, 2)}\n`, { flag: "wx", mode: 0o600 });
process.stdout.write(`${JSON.stringify({ status: "PASS", results_path: resultsPath, report_path: reportPath, terminal_runs: outcomes.length, claim_allowed: claimAllowed })}\n`);
