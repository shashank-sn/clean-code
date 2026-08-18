"use strict";

const { spawnSync } = require("child_process");
const path = require("path");

const packageRoot = path.resolve(__dirname, "..");
const goCheck = spawnSync("go", ["version"], { encoding: "utf8" });

if (goCheck.status !== 0) {
  console.log(
    "clean-code: Go is not installed. The clean-code CLI needs Go 1.22+ or a release binary."
  );
  console.log("See https://github.com/shashank-stitch/clean-code#install");
  process.exit(0);
}

const build = spawnSync(
  "go",
  ["build", "-o", path.join(packageRoot, "bin", "clean-code.bin"), "./cmd/clean-code"],
  { cwd: packageRoot, encoding: "utf8" }
);

if (build.status !== 0) {
  console.log("clean-code: optional native binary build skipped.");
  if (build.stderr) {
    console.log(build.stderr.trim());
  }
  console.log("You can still run: clean-code <command> (uses go run)");
}
