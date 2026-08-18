"use strict";

const { spawnSync } = require("child_process");
const path = require("path");
const { ensureGo, augmentEnv } = require("./runtime");

const packageRoot = path.resolve(__dirname, "..");

async function main() {
  try {
    const ok = await ensureGo(true);
    if (!ok) {
      console.log(
        "clean-code: Go could not be installed automatically. See https://github.com/shashank-stitch/clean-code#install"
      );
      process.exit(0);
    }
  } catch (error) {
    console.log(`clean-code: optional Go bootstrap skipped (${error.message}).`);
    process.exit(0);
  }

  const build = spawnSync(
    "go",
    [
      "build",
      "-o",
      path.join(packageRoot, "bin", "clean-code.bin"),
      "./cmd/clean-code",
    ],
    {
      cwd: packageRoot,
      encoding: "utf8",
      env: augmentEnv(process.env),
    }
  );

  if (build.status !== 0) {
    console.log("clean-code: optional native binary build skipped.");
    if (build.stderr) {
      console.log(build.stderr.trim());
    }
    console.log("You can still run: clean-code <command>");
  }
}

main().catch((error) => {
  console.log(`clean-code: postinstall skipped (${error.message}).`);
  process.exit(0);
});
