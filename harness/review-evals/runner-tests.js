#!/usr/bin/env node
const fs = require('fs');
const os = require('os');
const path = require('path');
const cp = require('child_process');
const crypto = require('crypto');

const root = __dirname;
function copyPack() { const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'review-eval-selftest-')); fs.cpSync(root, dir, {recursive: true}); return dir; }
function run(dir, args) { return cp.spawnSync(process.execPath, [path.join(dir, 'runner.js'), `--root=${dir}`, ...args], {encoding: 'utf8', timeout: 10000}); }
function assert(condition, message) { if (!condition) throw new Error(message); }
function output(result) { return `${result.stdout || ''}${result.stderr || ''}`; }

function testChangedPacketRejected() {
  const dir = copyPack(); try { fs.appendFileSync(path.join(dir, '01/packet/requirement.md'), '\nchanged'); const result = run(dir, ['--case=01']); assert(result.status !== 0, 'changed packet was accepted'); assert(output(result).includes('frozen reviewer material mismatch'), 'changed packet rejection lacked hash evidence'); } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testChangedPacketAndInventoryRejected() {
  const dir = copyPack(); try {
    const packet = path.join(dir, '01/packet/requirement.md'); fs.appendFileSync(packet, '\nchanged together');
    const digest = crypto.createHash('sha256').update(fs.readFileSync(packet)).digest('hex');
    const inventory = path.join(dir, 'reviewer-material.sha256'); const text = fs.readFileSync(inventory, 'utf8');
    fs.writeFileSync(inventory, text.replace(/^([0-9a-f]{64})(\s+01\/packet\/requirement\.md)$/m, `${digest}$2`));
    const result = run(dir, ['--case=01']); assert(result.status !== 0, 'changed packet plus inventory was accepted'); assert(output(result).includes('pinned reviewer inventory mismatch'), 'pinned inventory rejection lacked evidence');
  } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testOracleManifestMutationRejected() {
  const dir = copyPack(); try { const file = path.join(dir, 'oracle/manifest.json'); const manifest = JSON.parse(fs.readFileSync(file, 'utf8')); manifest.cases = manifest.cases.filter(c => c.id !== '01'); fs.writeFileSync(file, JSON.stringify(manifest)); const result = run(dir, []); assert(result.status !== 0, 'oracle manifest mutation was accepted'); assert(output(result).includes('pinned oracle manifest mismatch'), 'oracle manifest rejection lacked evidence'); } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testOracleBytesMutationRejected() {
  const dir = copyPack(); try { fs.appendFileSync(path.join(dir, 'oracle/tests/01_test.go'), '\n// changed'); const result = run(dir, ['--case=01']); assert(result.status !== 0, 'oracle test mutation was accepted'); assert(output(result).includes('pinned oracle test mismatch'), 'oracle test rejection lacked evidence'); } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testChangedBeforeRejected() {
  const dir = copyPack(); try { fs.appendFileSync(path.join(dir, '01/before/service.go'), '\nchanged'); const result = run(dir, ['--case=01']); assert(result.status !== 0, 'changed before tree was accepted'); assert(output(result).includes('candidate/context binding mismatch'), 'changed before rejection lacked binding evidence'); } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testCompileFailureRejected() {
  const dir = copyPack(); try { fs.appendFileSync(path.join(dir, '03/after/writer.go'), '\nthis is invalid go'); const result = require(path.join(dir, 'runner.js')).runGo('03/after', 'oracle/tests/03_test.go', {command: 'go', timeoutMs: 30000}); assert(result.status === 'compile-fail', `compile failure was classified as ${result.status}`); } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testTimeoutRejected() {
  const dir = copyPack(); try { fs.writeFileSync(path.join(dir, 'oracle/tests/03_test.go'), 'package export\nimport "testing"\nfunc TestPartialWriteIsReported(t *testing.T) { select {} }\n'); const result = require(path.join(dir, 'runner.js')).runGo('03/after', 'oracle/tests/03_test.go', {command: 'go', timeoutMs: 50}); assert(result.status === 'timeout', `timeout was classified as ${result.status}`); } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testMissingCommandRejected() {
  const dir = copyPack(); try { const result = require(path.join(dir, 'runner.js')).runGo('01/after', 'oracle/tests/01_test.go', {command: 'review-eval-command-does-not-exist', timeoutMs: 30000}); assert(result.status === 'command-missing', `missing command was classified as ${result.status}`); } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testPublicTestFailureClassified() {
  const dir = copyPack(); try { fs.writeFileSync(path.join(dir, '09/after/public_test.go'), 'package access\nimport "testing"\nfunc TestPublicFailure(t *testing.T) { t.Fatal("public failure") }\n'); const result = require(path.join(dir, 'runner.js')).runGo('09/after', 'oracle/tests/09_test.go', {command: 'go', timeoutMs: 30000}); assert(result.status === 'public-test-fail', `public test failure was classified as ${result.status}`); } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testOraclePanicIsNotAssertion() {
  const dir = copyPack(); try { fs.writeFileSync(path.join(dir, 'oracle/tests/03_test.go'), 'package export\nimport "testing"\nfunc TestUnexpectedPanic(t *testing.T) { panic("unexpected infrastructure failure") }\n'); const result = require(path.join(dir, 'runner.js')).runGo('03/after', 'oracle/tests/03_test.go', {command: 'go', timeoutMs: 30000}); assert(result.status === 'oracle-runtime-fail', `oracle panic was classified as ${result.status}`); } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testOracleFailWithoutDiagnostic() {
  const dir = copyPack(); try { fs.writeFileSync(path.join(dir, 'oracle/tests/03_test.go'), 'package export\nimport "testing"\nfunc TestDirectFail(t *testing.T) { t.Fail() }\n'); const result = require(path.join(dir, 'runner.js')).runGo('03/after', 'oracle/tests/03_test.go', {command: 'go', timeoutMs: 30000}); assert(result.status === 'oracle-assertion-fail', `t.Fail was classified as ${result.status}`); } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testOracleFailNowWithoutDiagnostic() {
  const dir = copyPack(); try { fs.writeFileSync(path.join(dir, 'oracle/tests/03_test.go'), 'package export\nimport "testing"\nfunc TestDirectFailNow(t *testing.T) { t.FailNow() }\n'); const result = require(path.join(dir, 'runner.js')).runGo('03/after', 'oracle/tests/03_test.go', {command: 'go', timeoutMs: 30000}); assert(result.status === 'oracle-assertion-fail', `t.FailNow was classified as ${result.status}`); } finally { fs.rmSync(dir, {recursive: true, force: true}); }
}
function testLiveIsExplicitlyNotRun() {
  const result = cp.spawnSync(process.execPath, [path.join(root, 'runner.js'), '--live'], {encoding: 'utf8'}); assert(result.status === 2, 'live mode did not return the NOT_RUN status code'); const parsed = JSON.parse(result.stdout); assert(parsed.status === 'NOT_RUN' && parsed.evaluated === false, 'live mode implied evaluation');
}

for (const test of [testChangedPacketRejected, testChangedPacketAndInventoryRejected, testOracleManifestMutationRejected, testOracleBytesMutationRejected, testChangedBeforeRejected, testCompileFailureRejected, testTimeoutRejected, testMissingCommandRejected, testPublicTestFailureClassified, testOraclePanicIsNotAssertion, testOracleFailWithoutDiagnostic, testOracleFailNowWithoutDiagnostic, testLiveIsExplicitlyNotRun]) test();
console.log(JSON.stringify({status: 'PASS', tests: 13}));
