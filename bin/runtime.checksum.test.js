"use strict";

const fs = require("fs");
const os = require("os");
const path = require("path");

const {
  computeSha256,
  verifyChecksum,
  nodeChecksumFrom,
} = require("./runtime.js");

let failures = 0;

function check(ok, message) {
  if (ok) {
    console.log(`PASS: ${message}`);
  } else {
    failures += 1;
    console.error(`FAIL: ${message}`);
  }
}

async function expectReject(promise, message) {
  try {
    await promise;
    failures += 1;
    console.error(`FAIL: ${message} (expected an error, got success)`);
  } catch (err) {
    console.log(`PASS: ${message}`);
  }
}

(async () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "runtime-checksum-"));
  const filePath = path.join(tempDir, "sample.bin");
  fs.writeFileSync(filePath, "clean-code checksum test payload\n");

  const digest = await computeSha256(filePath);
  check(
    /^[0-9a-f]{64}$/.test(digest),
    `computeSha256 returns a 64-character hex digest (${digest})`
  );

  try {
    await verifyChecksum(filePath, digest.toUpperCase(), "test archive");
    check(true, "verifyChecksum accepts the correct checksum (case-insensitive)");
  } catch (err) {
    check(false, `verifyChecksum accepted the correct checksum: ${err.message}`);
  }

  const tampered = digest === "0".repeat(64) ? "1".repeat(64) : "0".repeat(64);
  await expectReject(
    verifyChecksum(filePath, tampered, "test archive"),
    "verifyChecksum rejects a tampered checksum"
  );

  await expectReject(
    verifyChecksum(filePath, "not-a-hex-digest", "test archive"),
    "verifyChecksum rejects a malformed checksum string"
  );

  await expectReject(
    Promise.resolve().then(() =>
      nodeChecksumFrom(
        `${"a".repeat(64)}  node-v20.18.0-darwin-arm64.tar.xz\n`,
        "node-v20.18.0-darwin-x64.tar.xz"
      )
    ),
    "nodeChecksumFrom rejects an archive missing from the checksum list"
  );

  fs.rmSync(tempDir, { recursive: true, force: true });

  if (failures > 0) {
    console.error(`\n${failures} assertion(s) failed`);
    process.exit(1);
  }
  console.log("\nAll checksum tests passed");
  process.exit(0);
})().catch((err) => {
  console.error(`runtime.checksum.test.js crashed: ${err.stack || err.message}`);
  process.exit(1);
});
