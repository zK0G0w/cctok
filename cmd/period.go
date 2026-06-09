package cmd

import (
	"fmt"
	"time"

	"cctok/internal/parser"
	"cctok/internal/stats"
)

// filterByPeriod 根据 period 字符串过滤记录并生成对应 label
func filterByPeriod(records []parser.Record, period string, now time.Time) ([]parser.Record, string) {
	switch period {
	case "weekly", "week":
		return stats.FilterByWeek(records, now), fmt.Sprintf("This Week (%s)", weekRange(now))
	case "monthly", "month":
		return stats.FilterByMonth(records, now), fmt.Sprintf("This Month (%s)", now.Format("2006-01"))
	default:
		return stats.FilterByDay(records, now), fmt.Sprintf("Today (%s)", now.Format("2006-01-02"))
	}
}
