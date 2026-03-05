---
name: go-cli-ship
description: "Use this skill whenever users need to build, upgrade, or stabilize a Go CLI to production quality. It enforces deterministic CLI contracts, strict engineering gates, CI/release hardening, rollback safety, and auditable delivery standards from architecture to installation."
---

# Go CLI Ship

将“需求 -> 设计 -> 开发 -> 测试 -> 发版 -> 安装 -> 文档/技能沉淀”变成可复用、可审计、可回滚的生产级流程。

## 触发意图

当用户出现以下诉求时使用本技能：

- 从 0 到 1 开发 Go CLI。
- 将现有 CLI 提升到可发布、可维护、可扩展。
- CI/lint/test/release 不稳定，需要系统性闭环。
- 需要统一团队与 Agent 的工程标准。
- 希望把流程技能化，复用到后续项目。

## 第一性原理

1. CLI 本质是自动化接口，不只是命令行工具。
2. 稳定契约优先于局部体验：命令、JSON、错误码都要可演进。
3. 发版是系统工程：版本一致性 + workflow 可重复 + 产物可验证。
4. 文档是产品一部分：README 直接影响采用率与维护成本。
5. 技能文档必须可执行，不是背景科普文。

## 输入契约（开始前必须确认）

- 产品目标：CLI 解决什么问题，目标用户是谁。
- 命令边界：人类可读输出与机器 JSON 输出是否分层。
- 兼容策略：是否允许破坏性变更，迁移窗口如何定义。
- 运行环境：Go 版本、CGO 要求、目标平台矩阵。
- 外部依赖：数据库/对象存储/API 认证来源与最小权限策略。
- 质量门禁：fmt/vet/lint/test/e2e/release-check 的通过标准。

若输入不完整，先补输入再实现。

## 输出契约（交付物必须具备）

- 可运行 CLI：核心命令、参数、帮助、退出码、错误模型。
- 稳定契约：关键命令支持 `--json`，字段遵循增量兼容。
- 自动化链路：Makefile、CI workflow、release workflow、安装脚本。
- 可验证发布：多平台产物、`SHA256SUMS`、版本一致性检查。
- 文档体系：README、CHANGELOG、AGENTS（或等价执行规范）。
- Skill 封装：`SKILL.md` 可稳定指导 Agent 全流程执行。

## 逆向约束（先防失败）

优先检查高频失败点：

1. `go.mod` Go 版本与 CI/lint 构建器是否一致。
2. `golangci-lint` 是否因 Go 版本不匹配导致 panic。
3. release 产物下载后是否为平铺文件（避免 checksum 扫到目录）。
4. tag 是否指向正确提交（release 通常仅对 tag push 触发）。
5. `Makefile/install.sh/CHANGELOG.md` 版本是否一致。
6. README 是否缺少价值主张、安装路径、输出契约、云端教程。
7. SKILL frontmatter 是否仅保留 `name` 和 `description`。

## 非功能可靠性基线

- 可重复：本地与 CI 在同样输入下得到一致结果。
- 可诊断：失败日志包含命令、上下文、退出码。
- 可恢复：关键阶段具备最小回滚策略。
- 可审计：版本源、门禁结果、发布信息可追溯。
- 可扩展：新增命令和字段默认增量兼容。

## 标准执行流程（架构师视角，强制顺序）

### 0) 预检

1. 对齐 Go、lint、workflow 版本。
2. 检查目录结构、模块路径、依赖边界。
3. 检查仓库脏状态并确认是否可继续。

### 1) 领域建模

- 定义核心实体、状态机、错误码字典。
- 先写状态流转表与失败恢复策略。
- 错误对象建议：`code/message/retryable/details?`。

### 2) CLI 接口与合同设计

- 人类接口可读、低学习成本。
- 机器接口稳定 JSON 契约，字段新增优先。
- 关键自动化命令支持 `--json`。
- 参数演进采用增量兼容，避免破坏旧脚本。

### 3) 目录与模块

- `cmd/` 命令层。
- `pkg/` 业务层与依赖适配层。
- `e2e/` 端到端测试。
- `scripts/` 安装、冒烟、一致性检查。
- `.github/workflows/` CI 与 release。
- `skills/<skill-name>/SKILL.md` 技能封装。

### 4) 开发规范

- Go 版本固定（示例：`1.26.0`）。
- 需要 CGO 时统一 `CGO_ENABLED=1`。
- 错误返回用稳定错误码，不依赖字符串匹配。
- 不提交密钥、Token、私有地址与敏感配置。

### 5) 测试策略

- 单元测试覆盖状态机、边界输入、错误路径。
- 集成测试覆盖仓储与依赖适配层。
- e2e 覆盖核心闭环与失败回退。
- 涉及外部认证时至少一次真实 smoke 验证。

### 6) 代码检查门禁（必须全绿）

```bash
gofmt -l .
go vet ./...
golangci-lint run
CGO_ENABLED=1 go test -count=1 ./...
make release-check
```

任一失败即停止推进，先修复再继续。

### 7) Makefile 规范

- 至少包含：`build`、`fmt`、`lint`、`test`、`release-check`。
- `build` 默认产出 `bin/<cli-name>`。
- `release-check` 至少检查版本一致性与关键文件完整性。

### 8) GitHub Workflow 规范

- `ci.yml`：fmt/vet/test/lint/build 可分 stage 执行。
- `release.yml`：tag 触发、多平台构建、checksum、自动发布。
- `actions/setup-go` 与 `go.mod` 主版本一致。
- `golangci-lint-action@v7` 与 lint 主版本对齐。

### 9) 发版规范

- 采用 SemVer（`MAJOR.MINOR.PATCH`）。
- 发布前统一对齐版本源：
  - `scripts/install.sh` 的 `VERSION`
  - `Makefile` 的版本输出
  - `CHANGELOG.md` 最新版本头
- 强制执行 `make release-check`。

### 10) 安装与分发

- 提供 `scripts/install.sh`：平台检测、下载 release 资产、PATH 提示。
- 安装脚本必须输出最小诊断信息（URL、平台、文件大小或校验信息）。
- 发布产物必须包含 `SHA256SUMS` 并可验证。

### 11) 文档与 Skill 封装

- README 提供价值主张、快速开始、生产卡点、JSON 契约示例。
- CHANGELOG 遵循 Keep a Changelog。
- SKILL.md 聚焦可执行流程，复杂循环放到 `scripts/`。

## 失败兜底与回滚

- lint/toolchain 冲突：先对齐 Go 与 linter 构建版本。
- release 产物校验失败：阻断发布并保留失败产物用于诊断。
- tag 打错提交：删除错误 tag 并在正确提交重新打 tag。
- 云端 smoke 失败：先降级为只读路径，保留最小可复现命令。

## 交付报告模板（每次交付必须输出）

1. 范围：本次实现内容与明确非范围。
2. 变更：关键文件清单与目的。
3. 验证：门禁命令与结果。
4. 风险：已知风险、影响面、回滚方案。
5. 后续：下一步建议与优先级。

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

## CHANGELOG 规范

- 文件名固定：`CHANGELOG.md`。
- 采用 Keep a Changelog：`Added`/`Changed`/`Fixed`/`Docs`/`Tests`。
- 最新版本置顶：`## [x.y.z] - YYYY-MM-DD`。
- 只记录用户可感知变化，不写过程噪音。

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

# 版本与文档一致性
make release-check
```

## 与当前 agent-nexus 基线对齐建议

- Go 基线：`1.26.0`
- lint：`golangci-lint v2.x` + `golangci-lint-action@v7`
- 默认门禁：`go vet` + `golangci-lint run` + `CGO_ENABLED=1 go test -count=1 ./...`
- 发版前：`make release-check`
- 真实联调：`scripts/real_smoke.sh --env-file <path>`

## 常见故障速解

- lint panic：`application built with go1.xx` 低于项目 Go 版本。
  - 处理：升级/固定 linter 安装方式，确保与目标 Go 版本兼容。
- checksum 报目录错误。
  - 处理：artifact 下载合并到单目录后再计算。
- 改了 workflow 但 release 不触发。
  - 处理：检查是否 push tag；必要时重打 tag 到正确提交。
- 本地通过 CI 失败。
  - 处理：按 CI 命令本地复现并统一构建参数与缓存策略。

## Agent 执行纪律

- 先确认输入契约，再改代码。
- 每个阶段结束必须输出可验证结果。
- 不以“猜测通过”替代命令验证。
- 无法执行时明确阻塞原因与替代验证。
- 不提交密钥、令牌、私有地址与敏感配置。

## 完成定义（DoD）

- 功能满足需求，命令与 JSON 契约稳定。
- 所有质量门禁通过，本地与 CI 一致。
- release 稳定产出多平台二进制与 `SHA256SUMS`。
- 安装脚本可用，失败时有最小诊断信息。
- `CHANGELOG.md`、README、AGENTS、SKILL 文档同步更新。
- 每次交付均附带交付报告，可复核可追溯。
