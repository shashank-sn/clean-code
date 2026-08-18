#!/usr/bin/env node
"use strict";

const { spawnSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const {
  augmentEnv,
  ensureGo,
  goBinaryPath,
  goWorks,
} = require("./runtime");
const { readPackageMeta, goVersionLdflags } = require("./package-meta");

const packageRoot = path.resolve(__dirname, "..");
const args = process.argv.slice(2);
const localBinary = path.join(packageRoot, "bin", "clean-code.bin");

function runGo() {
  const goCmd = goBinaryPath();
  const result = spawnSync(goCmd, ["run", "./cmd/clean-code", ...args], {
    cwd: packageRoot,
    stdio: "inherit",
    env: augmentEnv(process.env),
  });
  if (result.error) {
    console.error("clean-code: failed to run Go CLI:", result.error.message);
    console.error(
      "See https://github.com/shashank-stitch/clean-code#install"
    );
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

function buildNativeBinary(verbose) {
  if (fs.existsSync(localBinary)) {
    return localBinary;
  }

  if (verbose) {
    console.log("clean-code: building native CLI (one-time, first run)...");
  }

  const build = spawnSync(
    "go",
    [
      "build",
      "-ldflags",
      goVersionLdflags(packageRoot),
      "-o",
      localBinary,
      "./cmd/clean-code",
    ],
    {
      cwd: packageRoot,
      encoding: "utf8",
      env: augmentEnv(process.env),
    }
  );

  if (build.status === 0 && fs.existsSync(localBinary)) {
    return localBinary;
  }

  if (verbose && build.stderr) {
    console.log(build.stderr.trim());
  }
  return null;
}

async function main() {
  if (args[0] === "version" && args.length === 1) {
    const { version } = readPackageMeta(packageRoot);
    console.log(version);
    process.exit(0);
  }

  const hadGo = goWorks();

  try {
    if (!hadGo) {
      console.log("clean-code: Go not found — bootstrapping runtime...");
      await ensureGo(true);
    } else {
      await ensureGo(false);
    }
  } catch (error) {
    console.error(`clean-code: ${error.message}`);
    process.exit(1);
  }

  const binary = buildNativeBinary(!hadGo);
  if (binary) {
    runBinary(binary);
  }

  runGo();
}

main().catch((error) => {
  console.error(`clean-code: ${error.message}`);
  process.exit(1);
});
