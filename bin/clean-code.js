#!/usr/bin/env node
"use strict";

const { spawnSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const packageRoot = path.resolve(__dirname, "..");
const args = process.argv.slice(2);

function runGo() {
  const result = spawnSync("go", ["run", "./cmd/clean-code", ...args], {
    cwd: packageRoot,
    stdio: "inherit",
    env: process.env,
  });
  if (result.error) {
    console.error("clean-code: failed to run Go CLI:", result.error.message);
    console.error("Install Go 1.22+ or use a release binary from GitHub.");
    process.exit(1);
  }
  process.exit(result.status ?? 1);
}

function runBinary(binaryPath) {
  const result = spawnSync(binaryPath, args, {
    cwd: process.cwd(),
    stdio: "inherit",
    env: process.env,
  });
  if (result.error) {
    console.error("clean-code: failed to run binary:", result.error.message);
    process.exit(1);
  }
  process.exit(result.status ?? 1);
}

const localBinary = path.join(packageRoot, "bin", "clean-code.bin");
if (fs.existsSync(localBinary)) {
  runBinary(localBinary);
}

runGo();
