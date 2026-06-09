package cmd

import (
	"fmt"
	"os"
	"time"

	"cctok/internal/config"
	"cctok/internal/display"
	"cctok/internal/parser"
	"cctok/internal/stats"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Version: version,
	Use:     "cctok",
	Short:   "Claude Code & Codex token 用量统计工具",
	Long: `cctok - Claude Code & Codex token 用量与费用统计

读取本地 JSONL 会话文件，统计 token 用量并计算费用。
支持 Claude Code 和 OpenAI Codex 双工具，按项目、模型、会话等维度聚合。

数据来源:
  - Claude Code: ~/.claude/projects/
  - Codex:       ~/.codex/sessions/

配置文件: ~/.cctok/config.toml (通过 cctok init 或 cctok config 管理)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		records, err := parser.ParseAll(cfg.ClaudeDir)
		if err != nil {
			return err
		}
		records = stats.Dedup(records)
		records = stats.FilterByDay(records, time.Now())

		if len(records) == 0 {
			fmt.Println("今日暂无用量数据。")
			return nil
		}

		label := fmt.Sprintf("Today (%s)", time.Now().Format("2006-01-02"))
		summaries := stats.BuildSourceSummaries(records, cfg, label)
		display.RenderTodayBySource(summaries, label)
		return nil
	},
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
