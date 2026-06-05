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

var sessionsProject string

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "查看会话级详细用量",
	Long:  "展示今天每个会话的详细 token 用量，支持按项目名模糊过滤。",
	Example: `  cctok sessions
  cctok sessions --project branch_payment`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		records, err := parser.ParseAll(cfg.ClaudeDir)
		if err != nil {
			return err
		}
		records = stats.Dedup(records)
		records = stats.FilterByDay(records, time.Now())

		if sessionsProject != "" {
			records = stats.FilterByProject(records, sessionsProject)
		}

		if len(records) == 0 {
			fmt.Println("暂无匹配的会话数据。")
			return nil
		}

		sessions := stats.AggregateBySession(records, cfg)
		label := fmt.Sprintf("Today (%s)", time.Now().Format("2006-01-02"))
		display.RenderSessions(sessions, label)
		return nil
	},
}

func init() {
	sessionsCmd.Flags().StringVar(&sessionsProject, "project", "", "按项目名过滤（模糊匹配）")
	rootCmd.AddCommand(sessionsCmd)
}
