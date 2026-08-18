#!/usr/bin/env node
"use strict";

const { spawnSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const { augmentEnv, ensureGo, goBinaryPath } = require("./runtime");

const packageRoot = path.resolve(__dirname, "..");
const args = process.argv.slice(2);

function runGo() {
  const goCmd = goBinaryPath();
  const result = spawnSync(goCmd, ["run", "./cmd/clean-code", ...args], {
    cwd: packageRoot,
    stdio: "inherit",
    env: augmentEnv(process.env),
  });
  if (result.error) {
    console.error("clean-code: failed to run Go CLI:", result.error.message);
    console.error("Re-run npm install or see https://github.com/shashank-stitch/clean-code#install");
    process.exit(1);
  }
  process.exit(result.status ?? 1);
}

function runBinary(binaryPath) {
  const result = spawnSync(binaryPath, args, {
    cwd: process.cwd(),
    stdio: "inherit",
    env: augmentEnv(process.env),
  });
  if (result.error) {
    console.error("clean-code: failed to run binary:", result.error.message);
    process.exit(1);
  }
  process.exit(result.status ?? 1);
}

async function main() {
  try {
    await ensureGo(false);
  } catch (error) {
    console.error(`clean-code: ${error.message}`);
    process.exit(1);
  }

  const localBinary = path.join(packageRoot, "bin", "clean-code.bin");
  if (fs.existsSync(localBinary)) {
    runBinary(localBinary);
  }

  runGo();
}

main().catch((error) => {
  console.error(`clean-code: ${error.message}`);
  process.exit(1);
});
