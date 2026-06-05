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

var weeklyCmd = &cobra.Command{
	Use:   "weekly",
	Short: "查看本周 token 用量（按项目）",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		records, err := parser.ParseAll(cfg.ClaudeDir)
		if err != nil {
			return err
		}
		records = stats.Dedup(records)
		records = stats.FilterByWeek(records, time.Now())

		if len(records) == 0 {
			fmt.Println("本周暂无用量数据。")
			return nil
		}

		label := fmt.Sprintf("This Week (%s)", weekRange(time.Now()))
		summary := stats.BuildSummary(records, cfg, label)
		display.RenderToday(summary)
		return nil
	},
}

func weekRange(now time.Time) string {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	return fmt.Sprintf("%s ~ %s", monday.Format("01-02"), now.Format("01-02"))
}

func init() {
	rootCmd.AddCommand(weeklyCmd)
}
