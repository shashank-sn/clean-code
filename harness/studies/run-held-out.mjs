import { createHash } from "node:crypto";
import { appendFile, mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(fileURLToPath(import.meta.url));
const studyRoot = join(root, "held-out-v1");
const manifestPath = join(root, "held-out-v1.json");
const casePath = join(studyRoot, "cases.json");
const commitmentPath = join(studyRoot, "preregistration.json");
const configPath = join(studyRoot, "model-config.json");
const rawRoot = join(studyRoot, "raw");
const attemptsRoot = join(studyRoot, "attempts");
const journalPath = join(studyRoot, "execution-journal.ndjson");
const validateOnly = process.argv.includes("--validate-only");

const digest = body => createHash("sha256").update(body).digest("hex");
const parse = (body, label) => {
  try {
    return JSON.parse(body);
  } catch (error) {
    throw new Error(`${label} is not valid JSON: ${error.message}`);
  }
};
const requireValue = (condition, message) => {
  if (!condition) throw new Error(message);
};
const outputText = response => (response.output ?? [])
  .flatMap(item => item.content ?? [])
  .filter(item => item.type === "output_text")
  .map(item => item.text)
  .join("\n");
const wordCount = text => text.trim() === "" ? 0 : text.trim().split(/\s+/u).length;
const armsFor = index => index % 2 === 0 ? ["control", "workflow"] : ["workflow", "control"];

const manifestBody = await readFile(manifestPath);
const caseBody = await readFile(casePath);
const commitmentBody = await readFile(commitmentPath);
const configBody = await readFile(configPath);
const manifest = parse(manifestBody, "manifest");
const corpus = parse(caseBody, "case corpus");
const commitment = parse(commitmentBody, "oracle commitment");
const config = parse(configBody, "model config");

requireValue(digest(caseBody) === manifest.case_corpus_digest, "case corpus digest mismatch");
requireValue(commitment.commitments?.oracle_scoring_sha256 === manifest.oracle_corpus_digest, "oracle commitment mismatch");
requireValue(digest(configBody) === manifest.model_config_digest, "model config digest mismatch");
requireValue(config.model === "gpt-5-2025-08-07", "model must be the fixed GPT-5 snapshot");
requireValue(config.execution_order === "alternating", "execution order is not preregistered");
requireValue(config.store === false && Array.isArray(config.tools) && config.tools.length === 0, "study must disable storage and tools");
requireValue(Array.isArray(corpus.cases) && corpus.cases.length === manifest.tasks.length, "case/task count mismatch");

const tasks = new Map(manifest.tasks.map(task => [task.id, task]));
for (const item of corpus.cases) {
  const task = tasks.get(item.task_id);
  requireValue(task, `unknown case task: ${item.task_id}`);
  requireValue(task.model === config.model, `model mismatch for ${item.task_id}`);
  requireValue(task.limit === config.max_output_words, `word limit mismatch for ${item.task_id}`);
  requireValue(JSON.stringify(task.tools) === JSON.stringify(["none"]), `tool mismatch for ${item.task_id}`);
  requireValue(typeof item.requirement === "string" && typeof item.code === "string" && typeof item.response_format === "string", `invalid case: ${item.task_id}`);
}

if (validateOnly) {
  console.log(JSON.stringify({status: "PASS", model: config.model, tasks: corpus.cases.length, case_corpus_digest: digest(caseBody), oracle_corpus_digest: commitment.commitments.oracle_scoring_sha256, model_config_digest: digest(configBody)}));
  process.exit(0);
}

requireValue(typeof process.env.OPENAI_API_KEY === "string" && process.env.OPENAI_API_KEY !== "", "OPENAI_API_KEY is required");
await mkdir(rawRoot, { recursive: true });
await mkdir(attemptsRoot, { recursive: true });
const runs = [];

for (let index = 0; index < corpus.cases.length; index++) {
  const item = corpus.cases[index];
  for (const arm of armsFor(index)) {
    const attemptPath = join(attemptsRoot, `${item.task_id}-${arm}.attempt.json`);
    const startedAt = new Date().toISOString();
    await writeFile(attemptPath, `${JSON.stringify({schema_version:"1.0.0",task_id:item.task_id,arm,started_at:startedAt,state:"reserved"})}\n`, { flag: "wx", mode: 0o600 });
    const armInstruction = arm === "control" ? config.control_instruction : config.workflow_instruction;
    const input = `${config.shared_instruction}\n\n${armInstruction}\n\nRequirement:\n${item.requirement}\n\nCode:\n${item.code}\n\nResponse format:\n${item.response_format}`;
    const request = {
      model: config.model,
      input,
      reasoning: { effort: config.reasoning_effort },
      max_output_tokens: config.max_output_tokens,
      store: config.store,
      tools: config.tools,
      metadata: {
        study_id: manifest.study_id,
        task_id: item.task_id,
        arm,
        case_corpus_digest: manifest.case_corpus_digest,
        model_config_digest: manifest.model_config_digest
      }
    };
    let responseBody;
    let httpStatus;
    try {
      const response = await fetch(config.endpoint, {
        method: "POST",
        headers: { "content-type": "application/json", authorization: `Bearer ${process.env.OPENAI_API_KEY}` },
        body: JSON.stringify(request),
        signal: AbortSignal.timeout(120_000)
      });
      httpStatus = response.status;
      responseBody = Buffer.from(await response.arrayBuffer());
    } catch (error) {
      httpStatus = 0;
      responseBody = Buffer.from(JSON.stringify({error:{type:error.name,message:error.message}}));
    }
    const finishedAt = new Date().toISOString();
    const rawName = `${item.task_id}-${arm}.json`;
    await writeFile(join(rawRoot, rawName), responseBody, { flag: "wx", mode: 0o600 });
    let response = {};
    try { response = JSON.parse(responseBody); } catch {}
    const text = outputText(response);
    const status = httpStatus === 200 && response.status === "completed" && response.model === config.model && wordCount(text) <= config.max_output_words ? "completed" : httpStatus === 0 ? "timeout" : "failed";
    const run = {
      task_id: item.task_id,
      arm,
      started_at: startedAt,
      finished_at: finishedAt,
      http_status: httpStatus,
      response_id: response.id ?? "",
      response_model: response.model ?? "",
      response_status: response.status ?? "",
      execution_status: status,
      output_words: wordCount(text),
      artifact_path: `raw/${rawName}`,
      artifact_digest: digest(responseBody)
    };
    runs.push(run);
    await appendFile(journalPath, `${JSON.stringify(run)}\n`, { flag: "a", mode: 0o600 });
    await writeFile(attemptPath, `${JSON.stringify({schema_version:"1.0.0",task_id:item.task_id,arm,started_at:startedAt,finished_at:finishedAt,state:"terminal",execution_status:status,artifact_digest:run.artifact_digest})}\n`, { flag: "w", mode: 0o600 });
    if (status !== "completed") throw new Error(`${item.task_id}/${arm} failed; partial raw evidence is preserved`);
  }
}

await writeFile(join(studyRoot, "execution-index.json"), `${JSON.stringify({schema_version:"1.0.0",study_id:manifest.study_id,manifest_digest:digest(manifestBody),runs}, null, 2)}\n`, { flag: "wx", mode: 0o600 });
console.log(JSON.stringify({status:"PASS", model:config.model, runs:runs.length, manifest_digest:digest(manifestBody)}));
