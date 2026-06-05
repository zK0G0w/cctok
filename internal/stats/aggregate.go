package stats

import (
	"sort"
	"time"

	"cctok/internal/config"
	"cctok/internal/parser"
)

// GroupStats 聚合统计结果
type GroupStats struct {
	Name         string
	InputTokens  int
	OutputTokens int
	CacheWrite   int
	CacheRead    int
	TotalCost    float64
	RequestCount int
}

// Summary 汇总结果
type Summary struct {
	Projects    []GroupStats
	Models      []GroupStats
	TotalCost   float64
	TotalInput  int
	TotalOutput int
	TimeRange   string
}

// Dedup 按 message.id 去重，保留 output_tokens 最大的记录
func Dedup(records []parser.Record) []parser.Record {
	seen := make(map[string]int)
	result := make([]parser.Record, 0, len(records))
	for _, r := range records {
		if idx, ok := seen[r.MessageID]; ok {
			if r.Usage.OutputTokens > result[idx].Usage.OutputTokens {
				result[idx] = r
			}
		} else {
			seen[r.MessageID] = len(result)
			result = append(result, r)
		}
	}
	return result
}

// FilterByDay 过滤指定日期（本地时区）的记录
func FilterByDay(records []parser.Record, day time.Time) []parser.Record {
	loc := time.Now().Location()
	startOfDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var filtered []parser.Record
	for _, r := range records {
		local := r.Timestamp.In(loc)
		if !local.Before(startOfDay) && local.Before(endOfDay) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// AggregateByProject 按项目聚合统计
func AggregateByProject(records []parser.Record, cfg *config.Config) []GroupStats {
	m := make(map[string]*GroupStats)
	for _, r := range records {
		gs, ok := m[r.Project]
		if !ok {
			gs = &GroupStats{Name: r.Project}
			m[r.Project] = gs
		}
		gs.InputTokens += r.Usage.InputTokens
		gs.OutputTokens += r.Usage.OutputTokens
		gs.CacheWrite += r.Usage.CacheCreationInputTokens
		gs.CacheRead += r.Usage.CacheReadInputTokens
		gs.TotalCost += cfg.CalculateCost(r.Model, r.Usage.InputTokens, r.Usage.OutputTokens, r.Usage.CacheCreationInputTokens, r.Usage.CacheReadInputTokens)
		gs.RequestCount++
	}

	result := make([]GroupStats, 0, len(m))
	for _, gs := range m {
		result = append(result, *gs)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalCost > result[j].TotalCost
	})
	return result
}

// AggregateByModel 按模型聚合统计
func AggregateByModel(records []parser.Record, cfg *config.Config) []GroupStats {
	m := make(map[string]*GroupStats)
	for _, r := range records {
		gs, ok := m[r.Model]
		if !ok {
			gs = &GroupStats{Name: r.Model}
			m[r.Model] = gs
		}
		gs.InputTokens += r.Usage.InputTokens
		gs.OutputTokens += r.Usage.OutputTokens
		gs.CacheWrite += r.Usage.CacheCreationInputTokens
		gs.CacheRead += r.Usage.CacheReadInputTokens
		gs.TotalCost += cfg.CalculateCost(r.Model, r.Usage.InputTokens, r.Usage.OutputTokens, r.Usage.CacheCreationInputTokens, r.Usage.CacheReadInputTokens)
		gs.RequestCount++
	}

	result := make([]GroupStats, 0, len(m))
	for _, gs := range m {
		result = append(result, *gs)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalCost > result[j].TotalCost
	})
	return result
}

// BuildSummary 构建汇总结果
func BuildSummary(records []parser.Record, cfg *config.Config, label string) *Summary {
	projects := AggregateByProject(records, cfg)
	models := AggregateByModel(records, cfg)

	var totalCost float64
	var totalInput, totalOutput int
	for _, p := range projects {
		totalCost += p.TotalCost
		totalInput += p.InputTokens
		totalOutput += p.OutputTokens
	}

	return &Summary{
		Projects:    projects,
		Models:      models,
		TotalCost:   totalCost,
		TotalInput:  totalInput,
		TotalOutput: totalOutput,
		TimeRange:   label,
	}
}
