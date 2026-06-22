---
name: go-cli-ship
description: "Use this skill to build, upgrade, or stabilize Go CLI projects for production release. It covers CLI contracts, Go engineering gates, npm-based binary distribution, GitHub Release automation, README quality, Agent Skill packaging, skills list/read support, version checks, release checks, and CI smoke-test closure. Use when a user asks to create a Go CLI from scratch, harden an existing CLI, fix CI/release drift, publish a version, or standardize a reusable CLI scaffold."
---

# Go CLI Ship

Use this skill to turn a Go CLI into a reproducible product: stable command contract, testable implementation, npm installer, GitHub Release artifacts, concise README, optional Agent Skill package, and release gates that prevent drift.

## First Principles

- Treat the CLI as an automation API, not only a terminal UI.
- Keep stdout for results, stderr for diagnostics, and exit codes stable.
- Make JSON output additive and backward compatible.
- Make release repeatable: one version, one tag, verified artifacts, tested installer.
- Keep `skills/` as content only. Put Go code in `internal/`, not under Agent Skill folders.
- Keep README and SKILL.md operational. Remove internal paths, release process noise, and maintenance notes unless the target user needs them.

## Reject These Anti-Patterns

- Do not use labels like "Unix-friendly" or "Agent Native" as decoration. Convert them into concrete contracts: stable stdout/stderr, exit codes, JSON, noninteractive commands, and repeatable install flows.
- Do not put source code, generated files, changelog notes, or repository maintenance instructions inside `skills/`.
- Do not copy `skills/<name>/SKILL.md` into `internal/` or any other second source of truth. Runtime code should read the packaged or repository `skills/` directory.
- Do not publish a release after changing install behavior unless npm install, GitHub Release artifacts, checksum verification, and optional `skills list/read` all pass from a fresh environment.
- Do not make README or SKILL.md explain the whole architecture when the user only needs to install, configure, run, and verify the CLI.
- Do not rely on local success alone. CI must run the same gates, and release workflows must verify the packaged artifact.

## Start Here

Before changing a project:

```bash
git status --short
go version
npm --version
```

Identify:

- CLI binary name.
- Go module and main package.
- npm package name.
- GitHub repo.
- target OS/ARCH matrix.
- whether the CLI ships an Agent Skill.
- whether the CLI should expose `skills list/read`.

## Template Replacement Protocol

When copying files from `templates/`, replace every placeholder before running tests:

| Placeholder | Replace with |
|---|---|
| `__CLI_NAME__` | final binary name, for example `acme` |
| `__NPM_PACKAGE_NAME__` | npm package, for example `@org/acme` |
| `__GITHUB_REPO__` | `owner/repo` |
| `__RELEASE_ENV_PREFIX__` | uppercase env prefix without trailing underscore, for example `ACME` |
| `__GO_VERSION_FILE__` | Go version file, usually `go.mod` |
| `__GO_MAIN_PACKAGE__` | build package, for example `./cmd/acme` |
| `__SKILL_NAME__` | bundled skill directory name under `skills/` |

After replacing placeholders, run:

```bash
rg "__[A-Z0-9_]+__" .
node -c scripts/install.js
node -c scripts/run.js
bash -n scripts/release-check.sh
```

If `rg "__[A-Z0-9_]+__" .` returns any template placeholder, stop and replace it before continuing.

## Required Project Shape

Use this baseline unless the existing repository has stronger conventions:

```text
cmd/<cli>/          CLI entrypoint
internal/          private implementation
e2e/               end-to-end tests
scripts/           install, smoke, release checks
.github/workflows/ CI and release automation
skills/<name>/     Agent Skill content only
```

Rules:

- Do not put `.go` files under `skills/`.
- Do not duplicate `skills/<name>/SKILL.md` under `internal/`.
- If the CLI reads skills, keep reader and locator code under `internal/skillcontent`.
- Let npm `scripts/run.js` set `<PREFIX>_SKILLS_DIR` to the package `skills/` directory when present.
- Let local development read the repository `skills/` directory.

## CLI Contract

Define these before implementation:

- commands and subcommands.
- human output versus `--json` output.
- stable error and exit codes.
- config precedence: flags, environment, config file, defaults.
- compatibility policy for new fields and flags.

Prefer explicit commands and flags over hidden global state, background daemons, caches, or browser scraping.

## Required Agent Skill Commands

When a CLI ships an Agent Skill under `skills/`, add these commands:

```bash
<cli> skills list --json
<cli> skills read <name>
<cli> skills read <name> --json
<cli> skills read <name>/<path> --json
```

Contract:

- `skills list` returns JSON with `version`, `skills`, and `count`.
- `skills read` returns raw Markdown by default.
- `skills read --json` returns `version`, `skill`, `path`, `content`, and optional `guidance`.
- reject absolute paths, `..`, backslashes, directories, and unknown skills.
- tests cover list, read, JSON read, unknown skill, and path traversal.

Use `templates/go/internal/skillcontent/` as the starting point, then wire it into the CLI parser.

If the CLI intentionally ships `skills/` without `skills list/read`, remove the corresponding release-check assertions and document why. Do not leave the default checks failing.

## Quality Gates

Every production CLI should expose:

```bash
make release-check
```

The gate should cover:

- version consistency.
- required files.
- README install command.
- npm package metadata and `npm run pack:check`.
- `gofmt -l .`.
- `go vet ./...`.
- `go test -count=1 ./...`.
- locked `golangci-lint`.
- no Go files under `skills/`.
- if `skills/` exists: README documents `npx skills add`, `package.json` includes `skills/**`, and the release smoke test exercises `skills list/read` when available.

If local sandboxing blocks Go cache or `httptest`, rerun with a writable `GOCACHE` or with the same permissions CI has. Do not mark the gate green without a real command result.

## Failure Handling

Use this table instead of guessing through release failures:

| Trigger | First action | If still failing |
|---|---|---|
| Placeholder scan finds `__...__` | Replace all placeholders in copied templates | Remove unused template files or document intentionally literal placeholders |
| `npm run pack:check` omits `skills/**` | Add `skills/**` to `package.json.files` | Inspect `npm pack --json --dry-run` output and fix path roots |
| `skills list/read` fails from npm wrapper | Check `<PREFIX>_SKILLS_DIR` in `scripts/run.js` and package contents | Run the installed wrapper from a fresh prefix and inspect its environment |
| Checksum mismatch | Regenerate `SHA256SUMS` from flattened release artifacts | Stop release; rebuild artifacts from the exact tag |
| `NPM_TOKEN` missing | Add repository secret before tagging | Publish GitHub Release only after deciding whether npm publish is in scope |
| npm publish permission denied | Confirm package name, scope ownership, and `publishConfig.access` | Rename package or request scope access before retrying |
| Git tag and package version differ | Move the tag or update version sources before release | Delete the bad tag only after user confirmation |
| CI passes locally but fails in workflow | Reproduce the exact workflow command locally | Align Go, Node, CGO, cache, and lint versions |

## Version Checks

Before tagging a release, align all version sources:

- `package.json` version.
- `Makefile` version when present.
- CLI `version` variable or ldflags value.
- top `CHANGELOG.md` entry.
- Git tag `vX.Y.Z`.
- GitHub Actions `GITHUB_REF_NAME` when running in CI.

Release checks should fail fast on drift. A patch release is required after changing published install behavior, package contents, or release workflow outputs.

## npm Distribution

Recommend npm for developer installation:

```bash
npm install -g <package-name>
<cli-name> version
```

The npm package is a thin wrapper:

- `package.json` declares `bin`, `files`, `postinstall`, `pack:check`, and `publishConfig.access`.
- `scripts/install.js` downloads the matching GitHub Release archive and verifies `SHA256SUMS`.
- `scripts/run.js` forwards args, stdio, and exit code to the real binary.
- no binary is committed to git or packed from the working tree.
- first publish does not require manually creating the npm package; `npm publish --access public` creates it when the account or scope has permission.

If the project ships an Agent Skill, include `skills/**` in the npm package and let `scripts/run.js` set `<PREFIX>_SKILLS_DIR`.

## Agent Skill Packaging

Use Agent Skills for agent instructions, not project maintenance documentation.

Recommended README command:

```bash
npx skills add <owner>/<repo>
```

Skill rules:

- `SKILL.md` frontmatter contains only `name` and `description`.
- Body is concise, procedural, and English unless the project explicitly requires another language.
- Body describes how to use the CLI, not how to maintain the repository.
- Do not include changelogs, internal source paths, release process, or test commands unless the skill is specifically for development work.
- Keep `skills/<name>/` free of Go code and generated binaries.

## Release Automation

Use tag-triggered release workflow:

CHECKPOINT - STOP before pushing a release tag. Confirm with the user:

- final version and tag.
- npm package name and access.
- GitHub repository target.
- `NPM_TOKEN` is configured when npm publish is in scope.
- `make release-check` passed from the current commit.
- fresh npm install smoke test passed when install behavior changed.

- trigger: `push.tags: v*`.
- build linux, darwin, and windows for amd64 and arm64 when supported.
- inject version from the tag.
- archive `README.md`, `LICENSE`, `CHANGELOG.md`, and `skills/` when present.
- generate `SHA256SUMS` from flattened artifacts.
- smoke-test npm install on Linux and Windows before publishing.
- if the CLI supports `skills list/read`, smoke-test those commands from the npm-installed wrapper.
- publish GitHub Release before npm.
- publish npm with `NPM_TOKEN` after version-tag validation.

## README Standard

First screen should answer:

- what is this CLI?
- who is it for?
- how do I install it?
- what is the fastest useful command?

Keep README practical. Avoid internal architecture unless users need it to operate the CLI.

## Delivery Report

When finishing work, report:

- scope changed.
- important files changed.
- commands run and results.
- known risks or skipped checks.
- whether the work was committed, tagged, or only left in the worktree.
