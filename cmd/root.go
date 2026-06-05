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

var rootCmd = &cobra.Command{
	Use:   "cctok",
	Short: "Claude Code token 用量统计工具",
	Long: `cctok - Claude Code token 用量与费用统计

读取 Claude Code 本地 JSONL 会话文件，统计 token 用量并计算费用。
支持按项目、模型、会话等维度聚合，支持自定义模型定价和倍率。

数据来源: ~/.claude/projects/ 下的会话文件
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
		summary := stats.BuildSummary(records, cfg, label)
		display.RenderToday(summary)
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
