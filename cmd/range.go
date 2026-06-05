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

var (
	rangeFrom string
	rangeTo   string
)

var rangeCmd = &cobra.Command{
	Use:   "range",
	Short: "查看指定日期范围的 token 用量",
	Long:  "展示指定日期范围内的 token 消耗和费用，按项目分组汇总。",
	Example: `  cctok range --from 2026-06-01 --to 2026-06-05
  cctok range --from 2026-05-01 --to 2026-05-31`,
	RunE: func(cmd *cobra.Command, args []string) error {
		from, err := time.ParseInLocation("2006-01-02", rangeFrom, time.Now().Location())
		if err != nil {
			return fmt.Errorf("无效的起始日期 --from: %w", err)
		}
		to, err := time.ParseInLocation("2006-01-02", rangeTo, time.Now().Location())
		if err != nil {
			return fmt.Errorf("无效的结束日期 --to: %w", err)
		}

		cfg := config.Load()
		records, err := parser.ParseAll(cfg.ClaudeDir)
		if err != nil {
			return err
		}
		records = stats.Dedup(records)
		records = stats.FilterByRange(records, from, to)

		if len(records) == 0 {
			fmt.Println("该时间范围内暂无用量数据。")
			return nil
		}

		label := fmt.Sprintf("%s ~ %s", rangeFrom, rangeTo)
		summary := stats.BuildSummary(records, cfg, label)
		display.RenderToday(summary)
		return nil
	},
}

func init() {
	today := time.Now().Format("2006-01-02")
	rangeCmd.Flags().StringVar(&rangeFrom, "from", today, "起始日期 (YYYY-MM-DD)")
	rangeCmd.Flags().StringVar(&rangeTo, "to", today, "结束日期 (YYYY-MM-DD)")
	rootCmd.AddCommand(rangeCmd)
}
