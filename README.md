# cctok

Claude Code & Codex 本地 token 用量统计与费用计算 CLI 工具。

读取本地 JSONL 会话文件，按项目、模型、时间等维度统计 token 消耗并估算费用。支持 Claude Code 和 OpenAI Codex 双工具，按来源分区展示。无需联网，纯本地离线分析。

## 数据来源

- **Claude Code**: `~/.claude/projects/`（含子代理）
- **Codex**: `~/.codex/sessions/`

## 安装

### Homebrew（推荐）

```bash
brew install zK0G0w/tap/cctok
```

### Go install

```bash
go install github.com/zK0G0w/cctok@latest
```

### 从源码编译

```bash
git clone https://github.com/zK0G0w/cctok.git
cd cctok
go build -o cctok .
```

## 快速开始

```bash
# 直接运行，查看今日用量
cctok

# 查看本周/本月用量
cctok weekly
cctok monthly

# 按模型维度查看
cctok models

# 指定日期范围
cctok range --from 2026-06-01 --to 2026-06-05

# 查看会话级详情
cctok sessions
cctok sessions --project branch_payment
```

## 输出示例

```
  This Week (06-02 ~ 06-08)  Total: $361.28

  ── Claude Code ──

  Project                         Input       Output      Cache W     Cache R     Cost
  ──────────────────────────────────────────────────────────────────────────────────────────
  Back-end/branch_payment_api     6.2M        505.8K      3.2M        95.1M       $173.81
  Back-end/non-degree-server      584.5K      69.7K       558.8K      12.1M       $40.47
  personal/cctok                  81.2K       85.6K       415.1K      18.0M       $27.36
  ──────────────────────────────────────────────────────────────────────────────────────────
  Total                           6.9M        661.1K                              $241.64

  ── Codex ──

  Project                         Input       Output      Cache W     Cache R     Cost
  ──────────────────────────────────────────────────────────────────────────────────────────
  Back-end/non-degree-server      46.0M       474.2K      0           37.9M       $114.66
  Back-end/branch_payment_api     1.8M        24.5K       0           1.5M        $4.62
  ahead/cs_project                136.4K      3.7K        0           92.8K       $0.35
  ──────────────────────────────────────────────────────────────────────────────────────────
  Total                           47.9M       502.4K                              $119.64
```

## 配置

首次运行即可使用（内置默认定价），如需自定义：

```bash
# 生成配置文件
cctok init

# 用编辑器打开配置
cctok config

# 交互式修改定价和倍率
cctok config -i
```

配置文件位于 `~/.cctok/config.toml`：

```toml
# 全局倍率（默认 1.0，费用 = token × 单价 × 倍率）
multiplier = 1.0

# 模型定价（$/1M tokens），支持前缀匹配，大小写不敏感

# --- Anthropic ---
[pricing.claude-opus-4]
input = 15.0
output = 75.0
cache_write = 18.75
cache_read = 1.50

[pricing.claude-sonnet-4]
input = 3.0
output = 15.0
cache_write = 3.75
cache_read = 0.30

# --- OpenAI (Codex) ---
[pricing.gpt-5.5]
input = 2.0
output = 8.0
cache_write = 0
cache_read = 0.50

[pricing.gpt-5.4]
input = 2.0
output = 8.0
cache_write = 0
cache_read = 0.50
```

支持精确模型名匹配（最长前缀优先），例如为 `claude-opus-4-8` 单独设定不同价格。

## 命令一览

| 命令 | 说明 |
|------|------|
| `cctok` | 查看今日用量（默认） |
| `cctok today` | 查看今日用量（按项目） |
| `cctok models` | 查看今日用量（按模型） |
| `cctok weekly` | 查看本周用量 |
| `cctok monthly` | 查看本月用量 |
| `cctok range` | 查看指定日期范围（`--from` / `--to`） |
| `cctok sessions` | 查看会话级详情（支持 `--project` 过滤） |
| `cctok init` | 生成默认配置文件 |
| `cctok config` | 编辑配置文件 |
| `cctok config -i` | 交互式配置向导 |

## 工作原理

1. 扫描 `~/.claude/projects/` 和 `~/.codex/sessions/` 下所有 `.jsonl` 文件
2. Claude Code：流式解析 `type: "assistant"` 记录，按 `message.id` 去重
3. Codex：解析 `token_count` 事件，从 `turn_context` 获取模型名
4. 按项目目录归属分组，从 `cwd` 字段提取可读的项目名
5. 根据配置的模型定价和倍率计算费用，按来源分区展示

## 技术栈

- Go 1.22+
- [cobra](https://github.com/spf13/cobra) — CLI 框架
- [lipgloss](https://github.com/charmbracelet/lipgloss) — 终端样式
- [huh](https://github.com/charmbracelet/huh) — 交互式表单
- [BurntSushi/toml](https://github.com/BurntSushi/toml) — TOML 配置解析

## License

MIT
