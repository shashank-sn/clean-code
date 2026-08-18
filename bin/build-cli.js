"use strict";

const { spawnSync } = require("child_process");
const path = require("path");
const { augmentEnv } = require("./runtime");
const { goVersionLdflags } = require("./package-meta");

const packageRoot = path.resolve(__dirname, "..");
const output = path.join(packageRoot, "bin", "clean-code.bin");

const build = spawnSync(
  "go",
  [
    "build",
    "-ldflags",
    goVersionLdflags(packageRoot),
    "-o",
    output,
    "./cmd/clean-code",
  ],
  {
    cwd: packageRoot,
    encoding: "utf8",
    env: augmentEnv(process.env),
  }
);

if (build.status !== 0) {
  console.error("clean-code: native CLI build failed.");
  if (build.stderr) {
    console.error(build.stderr.trim());
  }
  process.exit(build.status ?? 1);
}

console.log(`clean-code: built ${output}`);
