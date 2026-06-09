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

var modelsPeriod string

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "查看 token 用量（按模型）",
	Long:  "按模型分组展示 token 消耗和费用，支持 --period 指定时间范围。",
	Example: `  cctok models
  cctok models --period weekly
  cctok models --period monthly`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		records, err := parser.ParseAll(cfg.ClaudeDir)
		if err != nil {
			return err
		}
		records = stats.Dedup(records)
		records, label := filterByPeriod(records, modelsPeriod, time.Now())

		if len(records) == 0 {
			fmt.Println("暂无用量数据。")
			return nil
		}

		summaries := stats.BuildSourceSummaries(records, cfg, label)
		display.RenderModelsBySource(summaries, label)
		return nil
	},
}

func init() {
	modelsCmd.Flags().StringVarP(&modelsPeriod, "period", "p", "today", "时间范围: today, weekly, monthly")
	rootCmd.AddCommand(modelsCmd)
}
