---
name: go-cli-ship
description: Use this skill when building or upgrading a Go CLI tool to production quality, including architecture design, coding standards, lint/test gates, Makefile rules, GitHub workflows, release automation, installer script, CHANGELOG discipline, and skill packaging.
---

# Go CLI 生产级交付技能

用于把“拿到需求 -> 设计 -> 开发 -> 测试 -> 发版 -> 可安装”变成一套可复用、可审计、可自动化的高标准流程。

## 触发场景

- 你要从 0 到 1 开发一个 Go CLI。
- 你要把一个现有 CLI 提升到可发版、可维护、可扩展。
- 你要让大模型按统一工程规范稳定产出。

## 输入契约（开始前必须确认）

- 产品目标：CLI 解决什么问题，目标用户是谁。
- 命令边界：人类命令与机器命令是否分层（例如 `--json`）。
- 运行环境：Go 版本、是否需要 CGO、目标平台矩阵。
- 依赖系统：数据库/对象存储/API 凭据来源。
- 质量门禁：lint、test、e2e、release-check 的通过标准。

## 输出契约（交付物必须具备）

- 可运行 CLI：核心命令、参数、错误码、帮助文本。
- 自动化脚本：构建、测试、发布、一键安装。
- 工程门禁：Makefile、lint 配置、GitHub workflow。
- 文档体系：README（多语言可选）、CHANGELOG、AGENTS 规范。
- Skill 封装：`SKILL.md` 可指导 Agent 稳定执行全流程。

## 标准执行流程（架构师视角）

1. 领域建模
- 定义核心实体、状态机、错误码字典。
- 先写“状态转换表”和“失败恢复策略”。

2. CLI 接口设计
- 人类接口：可读输出、低学习成本。
- 机器接口：稳定 JSON 契约，字段兼容优先。
- 对自动化场景，所有关键命令支持 `--json`。

3. 目录与模块
- `cmd/` 命令层。
- `pkg/` 业务与存储层。
- `e2e/` 端到端测试。
- `scripts/` 安装、冒烟、一致性检查。
- `skills/<skill-name>/` Agent 技能封装。

4. 开发规范
- Go 版本固定（示例：`1.26.0`）。
- 需要 CGO 时统一 `CGO_ENABLED=1`。
- 错误返回统一编码，不让上层靠字符串匹配。
- 不提交密钥/Token/私有地址。

5. 测试规范
- 单元测试覆盖状态机与边界输入。
- 集成测试覆盖仓储与依赖适配层。
- e2e 验证核心业务闭环（创建、认领、提交、审核、回退）。
- 所有测试命令可在 CI 中无交互执行。

6. 代码检查门禁
- `gofmt -l .` 无输出。
- `go vet ./...` 通过。
- `golangci-lint run` 0 issues。
- `CGO_ENABLED=1 go test -count=1 ./...` 全绿。

7. Makefile 规范
- 至少包含：`build`、`test`、`lint`、`fmt`、`release-check`。
- `build` 默认产生 `bin/<cli-name>`。
- 所有门禁命令支持本地一键执行。

8. GitHub Workflow 规范
- `ci.yml`：fmt/goimports/vet/test/lint/build 分 job 或分 stage。
- `release.yml`：tag 触发、多平台构建、checksum、自动发布。
- lint action 与 lint 主版本匹配（如 v2 对应 action v7）。

9. 发版规范
- 采用 SemVer（`MAJOR.MINOR.PATCH`）。
- 发布前统一对齐版本源：
  - `scripts/install.sh` 的 `VERSION`
  - `Makefile` 的版本输出
  - `CHANGELOG.md` 最新版本头
- 强制执行 `make release-check`。

10. 安装与分发
- 提供 `scripts/install.sh`：平台检测、下载 release 资产、PATH 提示、可选配置向导。
- 安装脚本必须有失败提示与最小诊断信息（URL、平台、文件大小）。

11. Skill 封装
- 产出 `SKILL.md`（必须含 frontmatter `name`/`description`）。
- 复杂循环放 `scripts/`，`SKILL.md` 只定义策略和调用方式。
- 避免在 `SKILL.md` 堆砌大段背景解释，聚焦可执行流程。

## 超级元提示词（可直接给大模型）

```text
你是资深 Go CLI 架构师与交付负责人。请基于以下需求实现“生产级 CLI 工程”，并严格遵守质量门禁。

【项目目标】
{{PROJECT_GOAL}}

【功能范围】
{{FEATURE_SCOPE}}

【技术约束】
- Go: {{GO_VERSION|默认1.26.0}}
- 是否启用 CGO: {{CGO_REQUIRED|默认true}}
- 目标平台: {{TARGET_PLATFORMS|linux/amd64, linux/arm64, darwin/amd64, darwin/arm64}}

【强制工程规范】
1) CLI 分层：人类可读输出 + 机器 JSON 输出（关键命令支持 --json）
2) 错误码稳定：不可依赖错误文案匹配
3) 状态机先行：实现前先给出状态流转表
4) 安全合规：不得提交密钥/Token/私有凭据
5) 提交前必须通过：
   - gofmt -l .
   - go vet ./...
   - golangci-lint run
   - CGO_ENABLED=1 go test -count=1 ./...
   - make release-check

【必须产物】
- cmd/, pkg/, e2e/, scripts/, .github/workflows/ 的完整实现
- Makefile（build/test/lint/fmt/release-check）
- scripts/install.sh（可一键安装 release 资产）
- CHANGELOG.md（Keep a Changelog + SemVer）
- AGENTS.md（执行规范与检查清单）
- SKILL.md（可指导 Agent 自动执行）

【发版要求】
- Tag 触发 release workflow
- 多平台二进制 + SHA256SUMS
- 版本一致性检查：install.sh / Makefile / CHANGELOG 必须一致

【输出方式】
- 先给架构与任务分解，再直接落地代码。
- 每次改动后执行门禁并报告结果。
- 如果某项无法执行，明确说明阻塞原因与替代验证。
```

## 安装脚本模板（可复用）

```bash
#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
REPO="<org>/<repo>"
BASE="https://github.com/${REPO}/releases/download/v${VERSION}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "unsupported arch: $arch"; exit 1 ;;
esac
case "$os" in
  linux|darwin) ;; 
  *) echo "unsupported os: $os"; exit 1 ;;
esac

bin_name="<cli-name>"
asset="${bin_name}-${os}-${arch}"
url="${BASE}/${asset}"
install_dir="${HOME}/.local/bin"
mkdir -p "$install_dir"

tmp="/tmp/${bin_name}-$$"
curl -fsSL -o "$tmp" "$url"
chmod +x "$tmp"
mv "$tmp" "${install_dir}/${bin_name}"

if [[ ":$PATH:" != *":${install_dir}:"* ]]; then
  echo "add PATH: export PATH=\"\$PATH:${install_dir}\""
fi

echo "installed: ${install_dir}/${bin_name}"
```

## CHANGELOG 规范（你提到的 CHNAGELOG，按 CHANGELOG.md 执行）

- 文件名固定：`CHANGELOG.md`。
- 采用 Keep a Changelog 结构：`Added`/`Changed`/`Fixed`/`Docs`/`Tests`。
- 最新版本放在最上方，格式：`## [x.y.z] - YYYY-MM-DD`。
- 只写可感知变化，不写纯过程噪音。

## Go 测试与质量基线命令

```bash
# 编译
CGO_ENABLED=1 go build -o ./bin/<cli-name> ./main.go

# 格式
gofmt -l .

# 静态检查
go vet ./...
golangci-lint run

# 全量测试
CGO_ENABLED=1 go test -count=1 ./...

# 版本与文档一致性（建议做成脚本）
make release-check
```

## 与当前 agent-nexus 基线对齐建议

- Go 基线：`1.26.0`
- lint：`golangci-lint v2.x` + `golangci-lint-action@v7`
- 默认门禁：`go vet` + `golangci-lint run` + `CGO_ENABLED=1 go test -count=1 ./...`
- 发版前：`make release-check`
- 真实联调：`scripts/real_smoke.sh --env-file <path>`

## 完成定义（DoD）

- 新需求实现后，以上门禁全通过。
- CLI 人机接口一致，`--json` 契约稳定。
- 安装脚本可用，release 资产命名与 workflow 一致。
- `CHANGELOG.md`、README、AGENTS、SKILL 文档同步更新。
