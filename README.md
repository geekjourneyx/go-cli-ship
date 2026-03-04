# go-cli-ship

一个面向 Agent 的 Go CLI 生产级交付 Skill，帮助你把 `需求 -> 设计 -> 开发 -> 测试 -> 发布 -> 安装` 变成可复用、可审计、可自动化的标准流程。

## 推荐安装

```bash
npx skills add https://github.com/geekjourneyx/go-cli-ship --skill go-cli-ship
```

## 这个 Skill 解决什么问题

- 从 0 到 1 搭建 Go CLI 项目工程骨架
- 把现有 CLI 提升到可发布、可维护、可扩展
- 统一团队和 Agent 的工程规范，减少“每次重来”

## 核心能力

- 需求到交付的端到端执行流程（建模、实现、测试、发布）
- 生产级质量门禁（`fmt`/`vet`/`lint`/`test`/`release-check`）
- CI/CD 与发版规范（多平台构建、checksum、SemVer）
- 安装脚本与文档规范（`install.sh`、`CHANGELOG.md`、`AGENTS.md`）

## 快速使用

安装后，在你的 Agent 对话中明确提出目标即可，例如：

```text
请使用 go-cli-ship，帮我从 0 到 1 搭建一个 Go CLI：
- 目标：批量上传文件到对象存储
- 约束：支持 --json 输出，要求可自动化部署
- 门禁：go vet / golangci-lint / go test / release-check 必须通过
```

Skill 会引导并落地以下产物：

- CLI 命令与参数设计（含人类可读输出与机器可读 JSON 输出）
- 标准工程目录（`cmd/`、`pkg/`、`e2e/`、`scripts/`）
- Makefile、GitHub Workflows、发布与安装脚本
- CHANGELOG 与文档同步更新规范

## 最低质量门禁（建议）

```bash
gofmt -l .
go vet ./...
golangci-lint run
CGO_ENABLED=1 go test -count=1 ./...
make release-check
```

## 仓库结构

```text
.
├── README.md
└── skills/
    └── SKILL.md
```

## 仓库初始化（Git）

```bash
git branch -M main
git remote add origin git@github.com:geekjourneyx/go-cli-ship.git
```

## 最佳实践建议

- 始终为自动化场景提供稳定的 `--json` 输出契约
- 错误处理使用稳定错误码，避免依赖错误文案匹配
- 发布前统一检查版本一致性（`install.sh` / Makefile / CHANGELOG）
- 所有门禁命令保持无交互，确保 CI 可稳定执行

