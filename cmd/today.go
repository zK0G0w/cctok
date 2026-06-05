package cmd

import (
	"fmt"
	"time"

	"cctok/internal/config"
	"cctok/internal/display"
	"cctok/internal/parser"
	"cctok/internal/stats"

	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "查看今日 token 用量（按项目）",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Default()
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

func init() {
	rootCmd.AddCommand(todayCmd)
}
