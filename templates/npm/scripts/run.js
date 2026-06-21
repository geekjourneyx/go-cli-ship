#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");

const binaryName = process.platform === "win32" ? "__CLI_NAME__.exe" : "__CLI_NAME__";
const binaryPath = path.join(__dirname, "..", "bin", binaryName);
const skillsPath = path.join(__dirname, "..", "skills");
const SKILLS_ENV_VAR = "__RELEASE_ENV_PREFIX___SKILLS_DIR";

if (!fs.existsSync(binaryPath)) {
  console.error(
    "__CLI_NAME__ binary is missing. Reinstall with: npm install -g __NPM_PACKAGE_NAME__"
  );
  process.exit(1);
}

try {
  const env = { ...process.env };
  if (!env[SKILLS_ENV_VAR] && fs.existsSync(skillsPath)) {
    env[SKILLS_ENV_VAR] = skillsPath;
  }
  execFileSync(binaryPath, process.argv.slice(2), { stdio: "inherit", env });
} catch (error) {
  if (typeof error.status === "number") {
    process.exit(error.status);
  }
  console.error(error.message);
  process.exit(1);
}
