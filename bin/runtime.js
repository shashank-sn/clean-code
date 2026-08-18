"use strict";

const fs = require("fs");
const os = require("os");
const path = require("path");
const { spawnSync } = require("child_process");
const https = require("https");
const http = require("http");

const GO_VERSION = "1.22.10";
const NODE_VERSION = "20.18.0";
const MIN_NODE_MAJOR = 18;

function log(verbose, message) {
  if (verbose) {
    console.log(`clean-code: ${message}`);
  }
}

function runtimeHome() {
  const base =
    process.env.CLEAN_CODE_HOME ||
    path.join(os.homedir(), ".clean-code-cli");
  return path.join(base, "runtime");
}

function pathSeparator() {
  return process.platform === "win32" ? ";" : ":";
}

function pathWithRuntimes() {
  const parts = [];
  const home = runtimeHome();
  const goBin = path.join(home, "go", "bin");
  const nodeBin = path.join(home, "node", "bin");
  if (fs.existsSync(goBin)) {
    parts.push(goBin);
  }
  if (fs.existsSync(nodeBin)) {
    parts.push(nodeBin);
  }
  return parts.join(pathSeparator());
}

function augmentEnv(env = process.env) {
  const prefix = pathWithRuntimes();
  if (!prefix) {
    return env;
  }
  const key = process.platform === "win32" ? "Path" : "PATH";
  const existing = env[key] || "";
  return { ...env, [key]: `${prefix}${pathSeparator()}${existing}` };
}

function run(cmd, args, options = {}) {
  return spawnSync(cmd, args, {
    encoding: "utf8",
    ...options,
    env: augmentEnv(options.env || process.env),
  });
}

function parseGoVersion(output) {
  const match = (output || "").match(/go(\d+)\.(\d+)/);
  if (!match) {
    return null;
  }
  return { major: parseInt(match[1], 10), minor: parseInt(match[2], 10) };
}

function goWorks() {
  const result = run("go", ["version"]);
  if (result.status !== 0) {
    return false;
  }
  const version = parseGoVersion(result.stdout || result.stderr);
  if (!version) {
    return false;
  }
  return version.major > 1 || (version.major === 1 && version.minor >= 22);
}

function managedGoBinary() {
  const ext = process.platform === "win32" ? ".exe" : "";
  return path.join(runtimeHome(), "go", "bin", `go${ext}`);
}

function managedNodeBinary() {
  const ext = process.platform === "win32" ? ".exe" : "";
  return path.join(runtimeHome(), "node", "bin", `node${ext}`);
}

function nodeMajorFromOutput(output) {
  const major = parseInt((output || "").trim().split(".")[0], 10);
  return Number.isFinite(major) ? major : 0;
}

function systemNodeMajor() {
  const result = spawnSync("node", ["-p", "process.versions.node"], {
    encoding: "utf8",
  });
  if (result.status !== 0) {
    return 0;
  }
  return nodeMajorFromOutput(result.stdout);
}

function currentNodeMajor() {
  return nodeMajorFromOutput(process.versions.node);
}

function nodeWorks() {
  if (currentNodeMajor() >= MIN_NODE_MAJOR) {
    return true;
  }
  if (systemNodeMajor() >= MIN_NODE_MAJOR) {
    return true;
  }
  return fs.existsSync(managedNodeBinary());
}

function goPlatform() {
  const platform =
    process.platform === "win32"
      ? "windows"
      : process.platform === "darwin"
        ? "darwin"
        : "linux";
  const arch =
    process.arch === "x64"
      ? "amd64"
      : process.arch === "arm64"
        ? "arm64"
        : null;
  if (!arch) {
    throw new Error(`unsupported CPU architecture: ${process.arch}`);
  }
  return { platform, arch };
}

function nodePlatform() {
  const platform =
    process.platform === "win32"
      ? "win"
      : process.platform === "darwin"
        ? "darwin"
        : "linux";
  const arch =
    process.arch === "x64"
      ? "x64"
      : process.arch === "arm64"
        ? "arm64"
        : null;
  if (!arch) {
    throw new Error(`unsupported CPU architecture: ${process.arch}`);
  }
  return { platform, arch };
}

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

function downloadSync(url, dest) {
  ensureDir(path.dirname(dest));
  const curl = spawnSync("curl", ["-fsSL", url, "-o", dest], {
    encoding: "utf8",
  });
  if (curl.status === 0) {
    return;
  }

  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    const request = (targetUrl) => {
      const lib = targetUrl.startsWith("https") ? https : http;
      lib
        .get(targetUrl, (response) => {
          if (
            response.statusCode &&
            response.statusCode >= 300 &&
            response.statusCode < 400 &&
            response.headers.location
          ) {
            request(response.headers.location);
            return;
          }
          if (response.statusCode !== 200) {
            reject(
              new Error(`download failed (${response.statusCode}): ${targetUrl}`)
            );
            return;
          }
          response.pipe(file);
          file.on("finish", () => {
            file.close(resolve);
          });
        })
        .on("error", reject);
    };
    request(url);
  });
}

async function download(url, dest) {
  const syncResult = spawnSync("curl", ["-fsSL", url, "-o", dest], {
    encoding: "utf8",
  });
  if (syncResult.status === 0) {
    return;
  }
  await downloadSync(url, dest);
}

function runTar(args) {
  const result = spawnSync("tar", args, { encoding: "utf8" });
  if (result.status !== 0) {
    throw new Error(
      `tar ${args.join(" ")} failed: ${(result.stderr || "").trim()}`
    );
  }
}

function extractZip(zipPath, destDir) {
  if (process.platform === "win32") {
    const result = spawnSync(
      "powershell",
      [
        "-NoProfile",
        "-Command",
        `Expand-Archive -Path '${zipPath.replace(/'/g, "''")}' -DestinationPath '${destDir.replace(/'/g, "''")}' -Force`,
      ],
      { encoding: "utf8" }
    );
    if (result.status !== 0) {
      throw new Error(`zip extract failed: ${(result.stderr || "").trim()}`);
    }
    return;
  }
  runTar(["-xf", zipPath, "-C", destDir]);
}

function moveIntoPlace(source, target) {
  fs.rmSync(target, { recursive: true, force: true });
  ensureDir(path.dirname(target));
  fs.renameSync(source, target);
}

async function installGo(verbose = true) {
  const goBinary = managedGoBinary();
  if (fs.existsSync(goBinary) && goWorks()) {
    return path.dirname(path.dirname(goBinary));
  }

  const { platform, arch } = goPlatform();
  const home = runtimeHome();
  const cacheDir = path.join(home, "cache");
  ensureDir(cacheDir);

  let archiveName;
  let url;
  let isZip = false;
  if (platform === "windows") {
    archiveName = `go${GO_VERSION}.${platform}-${arch}.zip`;
    url = `https://go.dev/dl/${archiveName}`;
    isZip = true;
  } else {
    archiveName = `go${GO_VERSION}.${platform}-${arch}.tar.gz`;
    url = `https://go.dev/dl/${archiveName}`;
  }

  const archivePath = path.join(cacheDir, archiveName);
  log(verbose, `Downloading Go ${GO_VERSION} for ${platform}-${arch}...`);
  await download(url, archivePath);

  const extractRoot = path.join(cacheDir, "extract-go");
  fs.rmSync(extractRoot, { recursive: true, force: true });
  ensureDir(extractRoot);

  if (isZip) {
    extractZip(archivePath, extractRoot);
    moveIntoPlace(path.join(extractRoot, "go"), path.join(home, "go"));
  } else {
    runTar(["-xzf", archivePath, "-C", extractRoot]);
    moveIntoPlace(path.join(extractRoot, "go"), path.join(home, "go"));
  }

  if (!fs.existsSync(goBinary)) {
    throw new Error(`Go install finished but binary is missing: ${goBinary}`);
  }

  log(verbose, `Go ${GO_VERSION} installed under ${path.join(home, "go")}`);
  return path.join(home, "go");
}

async function installNode(verbose = true) {
  if (nodeWorks()) {
    return path.dirname(path.dirname(managedNodeBinary()));
  }

  const { platform, arch } = nodePlatform();
  const home = runtimeHome();
  const cacheDir = path.join(home, "cache");
  ensureDir(cacheDir);

  const folder = `node-v${NODE_VERSION}-${platform}-${arch}`;
  const archiveName =
    platform === "win" ? `${folder}.zip` : `${folder}.tar.xz`;
  const url = `https://nodejs.org/dist/v${NODE_VERSION}/${archiveName}`;
  const archivePath = path.join(cacheDir, archiveName);

  log(
    verbose,
    `Downloading Node.js ${NODE_VERSION} for ${platform}-${arch}...`
  );
  await download(url, archivePath);

  const extractRoot = path.join(cacheDir, "extract-node");
  fs.rmSync(extractRoot, { recursive: true, force: true });
  ensureDir(extractRoot);

  if (platform === "win") {
    extractZip(archivePath, extractRoot);
    moveIntoPlace(path.join(extractRoot, folder), path.join(home, "node"));
  } else {
    runTar(["-xJf", archivePath, "-C", extractRoot]);
    moveIntoPlace(path.join(extractRoot, folder), path.join(home, "node"));
  }

  const nodeBinary = managedNodeBinary();
  if (!fs.existsSync(nodeBinary)) {
    throw new Error(`Node install finished but binary is missing: ${nodeBinary}`);
  }

  log(
    verbose,
    `Node.js ${NODE_VERSION} installed under ${path.join(home, "node")}`
  );
  return path.join(home, "node");
}

async function ensureGo(verbose = true) {
  if (goWorks()) {
    return true;
  }
  await installGo(verbose);
  return goWorks();
}

async function ensureNode(verbose = true) {
  if (nodeWorks()) {
    return true;
  }
  await installNode(verbose);
  return nodeWorks();
}

async function ensureRuntimes(verbose = true) {
  await ensureNode(verbose);
  await ensureGo(verbose);
}

function goBinaryPath() {
  if (goWorks()) {
    const result = run("go", ["env", "GOROOT"]);
    if (result.status === 0) {
      const ext = process.platform === "win32" ? ".exe" : "";
      const candidate = path.join(
        (result.stdout || "").trim(),
        "bin",
        `go${ext}`
      );
      if (fs.existsSync(candidate)) {
        return candidate;
      }
    }
    return "go";
  }
  return managedGoBinary();
}

if (require.main === module) {
  const command = process.argv[2] || "ensure-all";
  const verbose = !process.argv.includes("--quiet");

  (async () => {
    try {
      if (command === "ensure-go") {
        await ensureGo(verbose);
      } else if (command === "ensure-node") {
        await ensureNode(verbose);
      } else if (command === "ensure-all") {
        await ensureRuntimes(verbose);
      } else if (command === "path") {
        process.stdout.write(pathWithRuntimes());
      } else {
        throw new Error(`unknown runtime command: ${command}`);
      }
    } catch (error) {
      console.error(`clean-code: ${error.message}`);
      process.exit(1);
    }
  })();
}

module.exports = {
  GO_VERSION,
  NODE_VERSION,
  MIN_NODE_MAJOR,
  runtimeHome,
  pathWithRuntimes,
  augmentEnv,
  goWorks,
  nodeWorks,
  ensureGo,
  ensureNode,
  ensureRuntimes,
  goBinaryPath,
  managedGoBinary,
  managedNodeBinary,
};
