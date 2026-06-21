#!/usr/bin/env bash
set -euo pipefail

version="${1:-}"
if [[ -z "$version" ]]; then
  version="$(node -p "require('./package.json').version")"
fi

if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo "invalid semver version: $version" >&2
  exit 1
fi

required_files=(
  "README.md"
  "CHANGELOG.md"
  "LICENSE"
  "Makefile"
  "package.json"
  "scripts/install.js"
  "scripts/run.js"
  ".github/workflows/ci.yml"
  ".github/workflows/release.yml"
)

for file in "${required_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "missing required file: $file" >&2
    exit 1
  fi
done

package_name="$(node -p "require('./package.json').name")"
package_version="$(node -p "require('./package.json').version")"
if [[ "$package_version" != "$version" ]]; then
  echo "package.json version $package_version does not match $version" >&2
  exit 1
fi

if [[ -n "${GITHUB_REF_NAME:-}" && "$GITHUB_REF_NAME" != "v$version" ]]; then
  echo "git ref $GITHUB_REF_NAME does not match v$version" >&2
  exit 1
fi

grep -q "npm install -g ${package_name}" README.md || {
  echo "README.md must document npm install -g ${package_name}" >&2
  exit 1
}

grep -q "## \[$version\]" CHANGELOG.md || {
  echo "CHANGELOG.md must contain ## [$version]" >&2
  exit 1
}

if grep -q "^VERSION[[:space:]]*?=" Makefile; then
  makefile_version="$(awk '/^VERSION[[:space:]]*\?=/{print $3; exit}' Makefile)"
  if [[ "$makefile_version" != "$version" ]]; then
    echo "Makefile VERSION $makefile_version does not match $version" >&2
    exit 1
  fi
fi

grep -q "npm publish --access public" .github/workflows/release.yml || {
  echo "release workflow must publish npm package" >&2
  exit 1
}

grep -q "NPM_TOKEN" .github/workflows/release.yml || {
  echo "release workflow must use NPM_TOKEN" >&2
  exit 1
}

grep -q "npm run pack:check" .github/workflows/ci.yml || {
  echo "ci workflow must run npm pack check" >&2
  exit 1
}

if [[ -d skills ]]; then
  skill_go_files="$(find skills -name '*.go' -print)"
  if [[ -n "$skill_go_files" ]]; then
    echo "skills/ must contain content only, not Go files:" >&2
    echo "$skill_go_files" >&2
    exit 1
  fi
  grep -q "npx skills add" README.md || {
    echo "README.md must document Agent Skill install with npx skills add" >&2
    exit 1
  }
  node -e 'const p=require("./package.json"); if (!Array.isArray(p.files) || !p.files.includes("skills/**")) process.exit(1)' || {
    echo "package.json files must include skills/** when skills/ exists" >&2
    exit 1
  }
  if [[ -d internal ]]; then
    duplicated_skill_files="$(find internal -name 'SKILL.md' -print 2>/dev/null || true)"
    if [[ -n "$duplicated_skill_files" ]]; then
      echo "skill content must not be duplicated under internal/:" >&2
      echo "$duplicated_skill_files" >&2
      exit 1
    fi
  fi
  grep -q "cp -R skills" .github/workflows/release.yml || {
    echo "release workflow must include skills/ in release archives" >&2
    exit 1
  }
  grep -q "skills list --json" .github/workflows/release.yml || {
    echo "release workflow must smoke-test skills list --json when skills/ exists" >&2
    exit 1
  }
  grep -q "skills read" .github/workflows/release.yml || {
    echo "release workflow must smoke-test skills read when skills/ exists" >&2
    exit 1
  }
fi

node -c scripts/install.js >/dev/null
node -c scripts/run.js >/dev/null
npm run pack:check >/dev/null

unformatted="$(gofmt -l .)"
if [[ -n "$unformatted" ]]; then
  echo "gofmt required:" >&2
  echo "$unformatted" >&2
  exit 1
fi

go vet ./...
golangci-lint run
CGO_ENABLED=1 go test -count=1 ./...
