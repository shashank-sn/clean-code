#!/usr/bin/env node
const fs = require('fs');

const CASES = Array.from({length: 12}, (_, i) => String(i + 1).padStart(2, '0'));
const ORACLES = new Map([
  ['01', {id: 'case-01-primary', highImpact: true}], ['02', {id: 'case-02-primary', highImpact: true}],
  ['03', {id: 'case-03-primary', highImpact: true}], ['04', {id: 'case-04-primary', highImpact: false}],
  ['05', {id: 'case-05-primary', highImpact: false}], ['06', {id: 'case-06-primary', highImpact: false}],
  ['07', {id: 'case-07-primary', highImpact: false}], ['08', {id: 'case-08-primary', highImpact: true}]
]);
const CONTROLS = new Set(['09', '10', '11', '12']);
const SEVERITIES = new Set(['blocking', 'high', 'medium', 'low']);
function hasConcreteEvidence(finding) {
  return typeof finding.title === 'string' && finding.title.trim() && SEVERITIES.has(finding.severity) && typeof finding.file === 'string' && finding.file.trim() && Number.isInteger(finding.line) && finding.line > 0 && typeof finding.explanation === 'string' && finding.explanation.trim() && typeof finding.evidence === 'string' && finding.evidence.trim();
}

function score(input) {
  const errors = [];
  if (!input || !Array.isArray(input.cases)) throw new Error('input must contain a cases array');
  const seenCases = new Set(); const seenFindings = new Set(); const found = new Set();
  let unsupported = 0; let unsupportedBlocking = 0; let incomplete = 0;
  for (const record of input.cases) {
    if (!record || typeof record.case_id !== 'string' || !CASES.includes(record.case_id)) { errors.push(`unknown case record: ${record && record.case_id}`); continue; }
    if (seenCases.has(record.case_id)) { errors.push(`duplicate case record: ${record.case_id}`); continue; }
    seenCases.add(record.case_id);
    if (typeof record.review_complete !== 'boolean') errors.push(`review_complete must be boolean: ${record.case_id}`);
    else if (record.review_complete !== true) incomplete++;
    if (record.model_completion !== undefined && (!record.model_completion || typeof record.model_completion !== 'object' || typeof record.model_completion.completed !== 'boolean' || typeof record.model_completion.rationale !== 'string')) errors.push(`model_completion must contain boolean completed and string rationale: ${record.case_id}`);
    if (!Array.isArray(record.findings)) { errors.push(`findings must be an array: ${record.case_id}`); continue; }
    const mapped = new Set();
    for (const finding of record.findings) {
      if (!finding || typeof finding.finding_id !== 'string' || !finding.finding_id) { errors.push(`finding_id required: ${record.case_id}`); continue; }
      if (seenFindings.has(finding.finding_id)) { errors.push(`duplicate finding_id: ${finding.finding_id}`); continue; }
      seenFindings.add(finding.finding_id);
      if (typeof finding.supported !== 'boolean') { errors.push(`supported must be boolean: ${finding.finding_id}`); continue; }
      if (typeof finding.blocking !== 'boolean') { errors.push(`blocking must be boolean: ${finding.finding_id}`); continue; }
      const oracle = finding.oracle_id;
      if (oracle !== null && oracle !== undefined) {
        const expected = ORACLES.get(record.case_id);
        if (!expected || oracle !== expected.id) { errors.push(`unknown oracle mapping: ${record.case_id}/${oracle}`); continue; }
        if (mapped.has(oracle)) { errors.push(`duplicate oracle mapping: ${record.case_id}/${oracle}`); continue; }
        if (finding.supported === true && !hasConcreteEvidence(finding)) { errors.push(`supported primary finding lacks concrete evidence: ${finding.finding_id}`); continue; }
        mapped.add(oracle);
        if (record.review_complete === true && finding.supported === true) found.add(oracle);
        else if (finding.supported === false) { unsupported++; if (finding.blocking === true) unsupportedBlocking++; }
      } else if (finding.supported === false) {
        unsupported++; if (finding.blocking === true) unsupportedBlocking++;
      }
    }
  }
  for (const id of CASES) if (!seenCases.has(id)) errors.push(`missing case record: ${id}`);
  if (errors.length) return {status: 'REJECTED', errors};
  const primaryFound = [...ORACLES.values()].filter(o => found.has(o.id)).length;
  const highImpactMisses = [...ORACLES.values()].filter(o => o.highImpact && !found.has(o.id)).map(o => o.id);
  const controlSilence = [...CONTROLS].every(id => input.cases.find(c => c.case_id === id).findings.length === 0);
  const acceptance = incomplete === 0 && highImpactMisses.length === 0 && primaryFound >= 7 && unsupportedBlocking === 0 && controlSilence;
  return {status: 'SCORED', acceptance, primary_recall: {found: primaryFound, total: 8, ratio: primaryFound / 8}, high_impact_misses: highImpactMisses, unsupported_findings: unsupported, unsupported_blocking_findings: unsupportedBlocking, control_silence: {silent: [...CONTROLS].filter(id => input.cases.find(c => c.case_id === id).findings.length === 0).length, total: 4}, incomplete_cases: incomplete, model_completion: input.cases.map(c => ({case_id: c.case_id, value: c.model_completion === undefined ? null : c.model_completion}))};
}

if (require.main === module) {
  try {
    const inputPath = process.argv[2]; if (!inputPath) throw new Error('usage: node scorer.js <adjudication.json>');
    const result = score(JSON.parse(fs.readFileSync(inputPath, 'utf8')));
    console.log(JSON.stringify(result, null, 2));
    if (result.status === 'REJECTED' || result.acceptance !== true) process.exitCode = 1;
  } catch (error) { console.error(JSON.stringify({status: 'REJECTED', errors: [error.message]}, null, 2)); process.exitCode = 2; }
}
module.exports = {score, ORACLES, CASES};
