"use strict";

const fs = require("fs");
const path = require("path");

function readPackageMeta(packageRoot) {
  const raw = fs.readFileSync(path.join(packageRoot, "package.json"), "utf8");
  const pkg = JSON.parse(raw);
  return {
    name: pkg.name || "@shashanksn/clean-code",
    version: pkg.version || "0.0.0",
  };
}

function goVersionLdflags(packageRoot) {
  const { version } = readPackageMeta(packageRoot);
  return `-X main.version=${version}`;
}

module.exports = {
  readPackageMeta,
  goVersionLdflags,
};
