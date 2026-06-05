# cctok

Claude Code 本地 token 用量统计与费用计算 CLI 工具。

读取 `~/.claude/projects/` 下的 JSONL 会话文件，按项目、模型、时间等维度统计 token 消耗并估算费用。无需联网，纯本地离线分析。

## 安装

```bash
go install github.com/zK0G0w/cctok@latest
```

或克隆后本地编译：

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
  This Week (06-01 ~ 06-05)  Total: $228.42

  Project                         Input       Output      Cache W     Cache R     Cost
  ──────────────────────────────────────────────────────────────────────────────────────────
  Back-end/branch_payment_api     2.3M        465.2K      3.2M        92.0M       $172.95
  Back-end/non-degree-server      567.8K      57.3K       518.1K      10.3M       $37.42
  personal/cctok                  69.8K       63.1K       335.7K      11.0M       $18.05
  ──────────────────────────────────────────────────────────────────────────────────────────
  Total                           2.9M        585.5K                              $228.42
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
# 全局倍率（如 Max plan 5x，默认 1.0）
multiplier = 1.0

# 模型定价（$/1M tokens），支持前缀匹配
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

[pricing.claude-haiku-4]
input = 0.80
output = 4.0
cache_write = 1.0
cache_read = 0.08
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

1. 扫描 `~/.claude/projects/` 下所有 `.jsonl` 文件（含子代理）
2. 流式解析，只提取 `type: "assistant"` 且包含 `usage` 的记录
3. 按 `message.id` 去重（保留 `output_tokens` 最大的记录）
4. 按项目目录归属分组，从 `cwd` 字段提取可读的项目名
5. 根据配置的模型定价和倍率计算费用

## 技术栈

- Go 1.22+
- [cobra](https://github.com/spf13/cobra) — CLI 框架
- [lipgloss](https://github.com/charmbracelet/lipgloss) — 终端样式
- [huh](https://github.com/charmbracelet/huh) — 交互式表单
- [BurntSushi/toml](https://github.com/BurntSushi/toml) — TOML 配置解析

## License

MIT
