const crypto = require("crypto");
const fs = require("fs");
const http = require("http");
const https = require("https");
const os = require("os");
const path = require("path");
const { execFileSync } = require("child_process");
const { fileURLToPath } = require("url");
const zlib = require("zlib");

const pkg = require("../package.json");

const CLI_NAME = "__CLI_NAME__";
const VERSION = pkg.version;
const REPO = "__GITHUB_REPO__";
const RELEASE_ENV_PREFIX = "__RELEASE_ENV_PREFIX__";

const TARGETS = {
  darwin: {
    x64: { goos: "darwin", goarch: "amd64", archive: "tar.gz" },
    arm64: { goos: "darwin", goarch: "arm64", archive: "tar.gz" },
  },
  linux: {
    x64: { goos: "linux", goarch: "amd64", archive: "tar.gz" },
    arm64: { goos: "linux", goarch: "arm64", archive: "tar.gz" },
  },
  win32: {
    x64: { goos: "windows", goarch: "amd64", archive: "zip" },
    arm64: { goos: "windows", goarch: "arm64", archive: "zip" },
  },
};

const target = TARGETS[process.platform]?.[process.arch];
const releaseBaseUrl =
  process.env[`${RELEASE_ENV_PREFIX}_RELEASE_BASE_URL`] ||
  `https://github.com/${REPO}/releases/download/v${VERSION}`;
const binaryName = process.platform === "win32" ? `${CLI_NAME}.exe` : CLI_NAME;
const binDir = path.join(__dirname, "..", "bin");
const destination = path.join(binDir, binaryName);

if (!target) {
  console.error(
    [
      `Unsupported platform for ${pkg.name}: ${process.platform}-${process.arch}`,
      "Supported npm install targets are:",
      "  - darwin-x64",
      "  - darwin-arm64",
      "  - linux-x64",
      "  - linux-arm64",
      "  - win32-x64",
      "  - win32-arm64",
    ].join("\n")
  );
  process.exit(1);
}

const archiveName = `${CLI_NAME}_${VERSION}_${target.goos}_${target.goarch}.${target.archive}`;

function hasScheme(value) {
  return /^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(value) && !path.win32.isAbsolute(value);
}

function resolveAssetLocation(base, name) {
  if (!hasScheme(base)) {
    return path.join(base, name);
  }
  return base.endsWith("/") ? `${base}${name}` : `${base}/${name}`;
}

function downloadToFile(source, destinationPath) {
  if (!hasScheme(source)) {
    fs.copyFileSync(source, destinationPath);
    return Promise.resolve();
  }

  if (source.startsWith("file://")) {
    fs.copyFileSync(fileURLToPath(source), destinationPath);
    return Promise.resolve();
  }

  return new Promise((resolve, reject) => {
    const client = source.startsWith("https:") ? https : http;

    client
      .get(source, (response) => {
        if ([301, 302, 307, 308].includes(response.statusCode) && response.headers.location) {
          response.resume();
          downloadToFile(response.headers.location, destinationPath).then(resolve, reject);
          return;
        }

        if (response.statusCode !== 200) {
          response.resume();
          reject(new Error(`download failed with status ${response.statusCode}: ${source}`));
          return;
        }

        const file = fs.createWriteStream(destinationPath);
        response.pipe(file);
        file.on("finish", () => file.close(resolve));
        file.on("error", reject);
      })
      .on("error", reject);
  });
}

function sha256(filePath) {
  const hash = crypto.createHash("sha256");
  hash.update(fs.readFileSync(filePath));
  return hash.digest("hex");
}

function expectedChecksum(checksumsPath, filename) {
  const line = fs
    .readFileSync(checksumsPath, "utf8")
    .split(/\r?\n/)
    .find((entry) => entry.trim().endsWith(` ${filename}`));

  if (!line) {
    throw new Error(`SHA256SUMS does not contain an entry for ${filename}`);
  }

  return line.trim().split(/\s+/)[0].toLowerCase();
}

function extractArchive(archivePath, extractDir) {
  if (archiveName.endsWith(".tar.gz")) {
    extractTarGz(archivePath, extractDir);
    return;
  }

  if (archiveName.endsWith(".zip") && process.platform === "win32") {
    const archiveLiteral = powershellSingleQuoted(archivePath);
    const extractLiteral = powershellSingleQuoted(extractDir);
    execFileSync("powershell.exe", [
      "-NoProfile",
      "-Command",
      `Expand-Archive -LiteralPath ${archiveLiteral} -DestinationPath ${extractLiteral} -Force`,
    ]);
    return;
  }

  throw new Error(`unsupported archive format for this platform: ${archiveName}`);
}

function powershellSingleQuoted(value) {
  return `'${String(value).replace(/'/g, "''")}'`;
}

function readTarString(buffer, start, length) {
  return buffer
    .subarray(start, start + length)
    .toString("utf8")
    .replace(/\0.*$/, "");
}

function extractTarGz(archivePath, extractDir) {
  const data = zlib.gunzipSync(fs.readFileSync(archivePath));
  let offset = 0;

  while (offset + 512 <= data.length) {
    const header = data.subarray(offset, offset + 512);
    if (header.every((byte) => byte === 0)) {
      break;
    }

    const name = readTarString(header, 0, 100);
    const prefix = readTarString(header, 345, 155);
    const fullName = prefix ? `${prefix}/${name}` : name;
    const sizeText = readTarString(header, 124, 12).trim();
    const size = sizeText ? parseInt(sizeText, 8) : 0;
    const typeflag = readTarString(header, 156, 1) || "0";
    const bodyOffset = offset + 512;
    const nextOffset = bodyOffset + Math.ceil(size / 512) * 512;

    const destinationPath = path.resolve(extractDir, fullName);
    const extractRoot = path.resolve(extractDir);
    if (destinationPath !== extractRoot && !destinationPath.startsWith(`${extractRoot}${path.sep}`)) {
      throw new Error(`refusing to extract path outside target directory: ${fullName}`);
    }

    if (typeflag === "5") {
      fs.mkdirSync(destinationPath, { recursive: true });
    } else if (typeflag === "0") {
      fs.mkdirSync(path.dirname(destinationPath), { recursive: true });
      fs.writeFileSync(destinationPath, data.subarray(bodyOffset, bodyOffset + size));
    }

    offset = nextOffset;
  }
}

function findBinary(root) {
  const entries = fs.readdirSync(root, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = path.join(root, entry.name);
    if (entry.isDirectory()) {
      const found = findBinary(fullPath);
      if (found) {
        return found;
      }
    } else if (entry.isFile() && entry.name === binaryName) {
      return fullPath;
    }
  }
  return "";
}

async function install() {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), `${CLI_NAME}-npm-`));
  const archivePath = path.join(tmpDir, archiveName);
  const checksumsPath = path.join(tmpDir, "SHA256SUMS");
  const extractDir = path.join(tmpDir, "extract");

  try {
    fs.mkdirSync(binDir, { recursive: true });
    fs.mkdirSync(extractDir, { recursive: true });

    await downloadToFile(resolveAssetLocation(releaseBaseUrl, archiveName), archivePath);
    await downloadToFile(resolveAssetLocation(releaseBaseUrl, "SHA256SUMS"), checksumsPath);

    const expected = expectedChecksum(checksumsPath, archiveName);
    const actual = sha256(archivePath);
    if (actual !== expected) {
      throw new Error(`checksum mismatch for ${archiveName}: expected ${expected}, got ${actual}`);
    }

    extractArchive(archivePath, extractDir);
    const sourceBinary = findBinary(extractDir);
    if (!sourceBinary) {
      throw new Error(`archive did not contain ${binaryName}`);
    }

    fs.copyFileSync(sourceBinary, destination);
    if (process.platform !== "win32") {
      fs.chmodSync(destination, 0o755);
    }

    console.log(`Installed ${CLI_NAME} ${VERSION} from ${resolveAssetLocation(releaseBaseUrl, archiveName)}`);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

install().catch((error) => {
  console.error(`Failed to install ${pkg.name}: ${error.message}`);
  process.exit(1);
});
