#!/usr/bin/env node
const fs = require('fs');
const os = require('os');
const path = require('path');
const crypto = require('crypto');
const cp = require('child_process');

const argv = process.argv.slice(2);
const option = name => { const hit = argv.find(a => a.startsWith(`--${name}=`)); return hit ? hit.slice(name.length + 3) : undefined; };
const root = path.resolve(option('root') || __dirname);
const manifest = JSON.parse(fs.readFileSync(path.join(root, 'manifest.json'), 'utf8'));
const oracleManifest = JSON.parse(fs.readFileSync(path.join(root, 'oracle/manifest.json'), 'utf8'));
const PINNED_PRIVATE = {
  inventory: '945f512ddcfe74e1e9688ca021654ca7b98ac0b4368d726c8cb0dd4a4981e647',
  oracleManifest: '33bb8fa96f5dbbaaa2ae4d5872fa46f639aab78efadf88bc80cf90eeb2e398e4',
  oracleTests: {
    '01': 'b0eb91f7b6c0b2a55ac5e517424c4ef0a26a6d1aa872a4c13deb1cf71df15f02', '02': 'c72fc96922601b5c358576689c52bcdcf079f919380dbfcd19fecfb082404e13',
    '03': '1cce41d4097b73866e14e9380350d7fa7cd25d6fd18ac8d93a3f5c40d467d00a', '04': 'e3401025f0497a60fe02b26a68b7246cb7fc7bc1934763aa9d676c7c9fad033d',
    '05': '4fdb11b8b7c7aff7ca3cf4b2ac80d3a8b0bf7fb9b461f1ae7d341ad1138094fd', '06': 'a1611561fe6f92a3f615dfa9ed7cf59adbad6377ff37e926950a8e86f76c072d',
    '07': 'a4b1a48cc6194e974cd3a9950e1b6a99397f864ad4e49103d71c55ad110c8ab8', '08': '7783523c778030b1620d122dcdd23b4d4837f0acfd7b2864721b9b372f65dd91',
    '09': '4e6fbe5d7b0151965a374e69bc97e9067a65903d2ac23d8cba28036c82b3ad71', '10': '84c1d3dc037b88c54e203d24905be4aa6874ad4f1c22974b6cbb8977c2f2c7a0',
    '11': '96bcb899ad83f829dc8350b0b559f6885b857ba1fded1d9ff4205e52bc5b74c1', '12': '8533d89d473c6a498c24747a864e321d9258c224ea416c3ad6ee41460b0ed482'
  }
};
const EXPECTED_CASES = new Map([...Array(8)].map((_, i) => [String(i + 1).padStart(2, '0'), 'defect']).concat([...Array(4)].map((_, i) => [String(i + 9).padStart(2, '0'), 'clean'])));
const expectedBindings = {
  '01': ['ea55f429873c9ee741def8ecabf1285e60cd3b0ec35a8e5855b5b7eb42457dcc', 'ccffed729f12592a46d4424b27181e77d5baa3c0e05d2df47fb4d2128aed988d', '9337791248186b0b2f26cb72102aed7aa6e1ec9a6cd03e2c6cedb356b7d58afb'],
  '02': ['f829ae84ee5585fab36d6b180ba77a957c4d6374530e20625cf30d963a511f52', '89b9f3a32ed0446887b7a9d6a00bc57b8428798fbbea3795b1e8b6fad42f8ae8', 'e15938071c457c11cd60111f072b954f648642168f86b0be15cbe222f3bf4bb9'],
  '03': ['e29a7e10918a4583c225edad14fbad0a3ed590dcf03f51807d248fed454e069d', 'd9b82294e6e9b904615371e7593c5bb9bfde7c7ad9d3bff92ac9c758ceb32b35', '74eb6d0b16e3583f324ca23c3b3a874f87fd3b5032a4793a5a59548193d21c1f'],
  '04': ['e591c6c434de95dd4e09c346a3413847a75a81664d19d65816cd0591253c6d69', 'b4ff8e6ed1f5c3e5cad876b0ab90e640182092eadbb7ae00fb14e4f0cb7dcce6', '59c52ace4c2eeb1142a31bdf2168853961a57879f87d6dd59a03d5aaeadb587b'],
  '05': ['8d22440e429bb0600486ef79b97a5e480868b6cfedf40342b92df1780325d360', 'f6fe38c721ef185939547c631d4eb723b299ac57fbd5ecbe7434f1e4e1f6c3a7', 'aac87ed77d3a18c90a6f9245a3638dafb01d05274968f5fc5f0f10cbf2f6c538'],
  '06': ['c7807bc63137498a9faa255c5543d119a3c44cfebf683763f2746844e2a18bb4', 'b5f982908f717f6b0787b30ebd7672d564e5f1a360f15357ee8a02208303af66', '320d2ec16b8752defbfd6e079635d6cb2d298d8f13a09782578024578464adda'],
  '07': ['e41916d6b85b5a2dcce1237d81fb2baffc8f93ea57be242cbb62c6a90e6ee6e2', 'fc4ddad85032e248d727c266cd35a6afd10c69d6b06512e3646300056cf4aecb', 'd75a5a4a28b55b80cd3d88f49eccc0232f3e510266a1c554d622087431a1c01d'],
  '08': ['4dfb106240e5ca7ff4e96e10e217d8352adb22e8b4b022ae541a531cf7fceff3', '867857e971b5edaddf587f3f8bce694ff96411d8ffa7423c65b44bb1c7cccfe3', '6c7835b81e29743cd7c52f99adfd15e2fd0720be95421eb5ec291b68b6e2beb4'],
  '09': ['0d255ad4ec94db92dff15b68e145c2b2bb89184cc86bfc2c89a808faa520c0f6', '0d255ad4ec94db92dff15b68e145c2b2bb89184cc86bfc2c89a808faa520c0f6', '0752187c270b46d90185307830220425ff986bd9af8aca089569a06fe1315784'],
  '10': ['e15543cce248c81e04719245e9c13e215213d429097699bbef99c7485c495fb0', '5e48abf14dd30b7be346ff1288b6e0002ca4987eec42ee058ac4088ddaeeb7f1', 'dfcb2151be9f760fcc3772f3528614fff7dd3561f9f9d69d5ba811ca80a58e10'],
  '11': ['3dd35b64cf40ac52f8521465d735eeaed55649bc9a4b5c4bdf296e269ca51983', '2fa81d5b787297b832ec672ceceaa9fa5b1296441d9320e9000a4a5e490661b1', '0e866d58a678faac58d6dc260c074ffcdfe0a26c1c4ee5e743f295917f3c31b9'],
  '12': ['cfd7ddd12d0058de55083305c527cfdb1909057747f09639f13b6aff59f4d8bb', 'cfd7ddd12d0058de55083305c527cfdb1909057747f09639f13b6aff59f4d8bb', 'd2bbdcbd61d0f0f25fc779fd298a241ae2227008adb059bfb63e4897f8acd9a5']
};
const FIXTURE_GOMAXPROCS = '1';

function digestFile(file) { return crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex'); }
function digestTree(dir) {
  const files = [];
  function walk(current) { for (const name of fs.readdirSync(current).sort()) { const file = path.join(current, name); const stat = fs.statSync(file); if (stat.isDirectory()) walk(file); else files.push(path.relative(dir, file)); } }
  walk(dir);
  const hash = crypto.createHash('sha256');
  for (const file of files) { hash.update(file); hash.update('\0'); hash.update(fs.readFileSync(path.join(dir, file))); hash.update('\0'); }
  return hash.digest('hex');
}
function validatePinnedPack() {
  if (digestFile(path.join(root, 'reviewer-material.sha256')) !== PINNED_PRIVATE.inventory) throw new Error('pinned reviewer inventory mismatch');
  if (digestFile(path.join(root, 'oracle/manifest.json')) !== PINNED_PRIVATE.oracleManifest) throw new Error('pinned oracle manifest mismatch');
  const seen = new Set();
  if (!Array.isArray(oracleManifest.cases) || oracleManifest.cases.length !== EXPECTED_CASES.size) throw new Error('oracle manifest must contain exactly 12 cases');
  for (const c of oracleManifest.cases) {
    if (!EXPECTED_CASES.has(c.id) || EXPECTED_CASES.get(c.id) !== c.kind || seen.has(c.id) || c.oracle !== `oracle/tests/${c.id}_test.go`) throw new Error(`unexpected oracle case set: ${c.id}`);
    seen.add(c.id);
    const testPath = path.join(root, c.oracle);
    if (!fs.existsSync(testPath) || digestFile(testPath) !== PINNED_PRIVATE.oracleTests[c.id]) throw new Error(`pinned oracle test mismatch: ${c.id}`);
  }
  if (seen.size !== EXPECTED_CASES.size) throw new Error('oracle case IDs are incomplete');
  if (manifest.cases.map(c => c.id).sort().join(',') !== [...EXPECTED_CASES.keys()].sort().join(',')) throw new Error('public manifest case set is incomplete');
}
function frozenInventory() {
  const file = path.join(root, 'reviewer-material.sha256');
  const entries = fs.readFileSync(file, 'utf8').split(/\r?\n/).filter(line => line && !line.startsWith('#')).map(line => {
    const match = line.match(/^([0-9a-f]{64})\s+(.+)$/); if (!match) throw new Error(`invalid frozen inventory line: ${line}`); return {hash: match[1], rel: match[2]};
  });
  const seen = new Set();
  for (const entry of entries) {
    if (seen.has(entry.rel)) throw new Error(`duplicate frozen inventory path: ${entry.rel}`); seen.add(entry.rel);
    if (path.isAbsolute(entry.rel) || entry.rel.split(path.sep).includes('..')) throw new Error(`unsafe frozen inventory path: ${entry.rel}`);
    const full = path.join(root, entry.rel); if (!fs.existsSync(full) || digestFile(full) !== entry.hash) throw new Error(`frozen reviewer material mismatch: ${entry.rel}`);
  }
  const listed = new Set(['manifest.json']);
  for (const item of manifest.cases) { const dir = path.join(root, item.packet); if (!fs.existsSync(dir)) throw new Error(`missing reviewer packet: ${item.packet}`); for (const name of fs.readdirSync(dir)) { const full = path.join(dir, name); if (fs.statSync(full).isDirectory()) { const walk = d => { for (const child of fs.readdirSync(d)) { const p = path.join(d, child); if (fs.statSync(p).isDirectory()) walk(p); else listed.add(path.relative(root, p)); } }; walk(full); } else listed.add(path.relative(root, full)); } }
  for (const probe of manifest.modes.live_model_evaluation.protocol_probes || []) { const dir = path.join(root, probe); if (!fs.existsSync(dir)) throw new Error(`missing protocol probe: ${probe}`); const walk = d => { for (const child of fs.readdirSync(d)) { const p = path.join(d, child); if (fs.statSync(p).isDirectory()) walk(p); else listed.add(path.relative(root, p)); } }; walk(dir); }
  for (const entry of entries) if (!listed.has(entry.rel)) throw new Error(`unexpected frozen reviewer material: ${entry.rel}`);
  for (const rel of listed) if (!seen.has(rel)) throw new Error(`unhashed reviewer material: ${rel}`);
}
function verifyBindings(cases) {
  for (const c of cases) {
    const expected = expectedBindings[c.id]; if (!expected) throw new Error(`missing candidate binding: ${c.id}`);
    const afterDigest = digestTree(path.join(root, c.after));
    const candidateDigest = digestTree(path.join(root, c.id, 'packet/candidate'));
    const diffDigest = digestFile(path.join(root, c.id, 'packet/change.diff'));
    const beforeDigest = digestTree(path.join(root, c.before));
    if (beforeDigest !== expected[0] || afterDigest !== expected[1] || candidateDigest !== afterDigest || diffDigest !== expected[2]) throw new Error(`candidate/context binding mismatch: ${c.id}`);
  }
}
function parseEvents(output) { return output.split(/\r?\n/).filter(Boolean).flatMap(line => { try { return [JSON.parse(line)]; } catch (_) { return []; } }); }
function runGo(sourceRel, oracleRel, opts) {
  const temp = fs.mkdtempSync(path.join(os.tmpdir(), 'review-eval-'));
  try {
    fs.cpSync(path.join(root, sourceRel), temp, {recursive: true}); fs.copyFileSync(path.join(root, oracleRel), path.join(temp, 'oracle_test.go'));
    const result = cp.spawnSync(opts.command, ['test', '-json', './...'], {
      cwd: temp,
      encoding: 'utf8',
      timeout: opts.timeoutMs,
      env: {...process.env, GOMAXPROCS: FIXTURE_GOMAXPROCS}
    });
    const output = `${result.stdout || ''}${result.stderr || ''}`; const events = parseEvents(output);
    if (result.error && result.error.code === 'ENOENT') return {status: 'command-missing', output};
    if (result.error && result.error.code === 'ETIMEDOUT') return {status: 'timeout', output};
    if (result.signal) return {status: `signal-${result.signal.toLowerCase()}`, output};
    if (result.error) return {status: 'process-error', output};
    if (result.status === 0) return {status: 'pass', output};
    const oracleText = fs.readFileSync(path.join(root, oracleRel), 'utf8');
    const oracleTests = new Set([...oracleText.matchAll(/func\s+(Test[A-Za-z0-9_]*)\s*\(/g)].map(match => match[1]));
    const oracleEvents = events.filter(e => typeof e.Test === 'string' && oracleTests.has(e.Test));
    if (oracleEvents.some(e => e.Action === 'output' && /fatal error:|panic:/.test(e.Output || ''))) return {status: 'oracle-runtime-fail', output};
    if (oracleEvents.some(e => e.Action === 'fail')) return {status: 'oracle-assertion-fail', output};
    if (events.some(e => e.Action === 'fail' && typeof e.Test === 'string')) return {status: 'public-test-fail', output};
    return {status: 'compile-fail', output};
  } finally { fs.rmSync(temp, {recursive: true, force: true}); }
}
function main() {
  if (argv.includes('--live') || argv.includes('--mode=live-model-evaluation')) { console.log(JSON.stringify({mode: 'live_model_evaluation', status: 'NOT_RUN', evaluated: false, network: false, reason: 'model/provider networking is outside this runner'}, null, 2)); return 2; }
  validatePinnedPack();
  frozenInventory();
  const selected = option('case'); const cases = selected ? oracleManifest.cases.filter(c => c.id === selected) : oracleManifest.cases;
  if (!cases.length) throw new Error('unknown case');
  verifyBindings(cases);
  const opts = {command: option('go-command') || 'go', timeoutMs: Number(option('timeout-ms') || 30000)};
  const results = []; let failed = 0;
  for (const c of cases) {
    const before = runGo(c.before, c.oracle, opts); const after = runGo(c.after, c.oracle, opts);
    const oracleFailure = after.status === 'oracle-assertion-fail' || after.status === 'oracle-runtime-fail';
    const expected = c.kind === 'defect' ? before.status === 'pass' && oracleFailure : before.status === 'pass' && after.status === 'pass';
    if (!expected) failed++;
    results.push({id: c.id, kind: c.kind, before: before.status, after: after.status, expected, diagnostics: expected ? undefined : {before: before.output.slice(-4000), after: after.output.slice(-4000)}});
  }
  console.log(JSON.stringify({mode: 'fixture_validation', total: cases.length, failed, results}, null, 2)); return failed ? 1 : 0;
}
if (require.main === module) { try { process.exitCode = main(); } catch (error) { console.error(JSON.stringify({mode: 'fixture_validation', status: 'REJECTED', error: error.message}, null, 2)); process.exitCode = 1; } }
module.exports = {main, frozenInventory, verifyBindings, runGo};
