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

var monthlyCmd = &cobra.Command{
	Use:   "monthly",
	Short: "查看本月 token 用量（按项目）",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		records, err := parser.ParseAll(cfg.ClaudeDir)
		if err != nil {
			return err
		}
		records = stats.Dedup(records)
		records = stats.FilterByMonth(records, time.Now())

		if len(records) == 0 {
			fmt.Println("本月暂无用量数据。")
			return nil
		}

		label := fmt.Sprintf("This Month (%s)", time.Now().Format("2006-01"))
		summary := stats.BuildSummary(records, cfg, label)
		display.RenderToday(summary)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(monthlyCmd)
}
