package stats

import (
	"sort"
	"strings"
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

// SourceSummary 按工具来源分组的汇总
type SourceSummary struct {
	Source  string
	Summary *Summary
}

// SplitBySource 按 Source 字段分组记录
func SplitBySource(records []parser.Record) map[string][]parser.Record {
	m := make(map[string][]parser.Record)
	for _, r := range records {
		source := r.Source
		if source == "" {
			source = "Claude Code"
		}
		m[source] = append(m[source], r)
	}
	return m
}

// BuildSourceSummaries 构建按工具分组的汇总列表
func BuildSourceSummaries(records []parser.Record, cfg *config.Config, label string) []SourceSummary {
	groups := SplitBySource(records)
	// 固定顺序：Claude Code 在前，Codex 在后
	order := []string{"Claude Code", "Codex"}
	var result []SourceSummary
	for _, source := range order {
		recs, ok := groups[source]
		if !ok || len(recs) == 0 {
			continue
		}
		summary := BuildSummary(recs, cfg, label)
		result = append(result, SourceSummary{Source: source, Summary: summary})
	}
	// 处理其他可能的来源
	for source, recs := range groups {
		if source == "Claude Code" || source == "Codex" {
			continue
		}
		summary := BuildSummary(recs, cfg, label)
		result = append(result, SourceSummary{Source: source, Summary: summary})
	}
	return result
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

// FilterByWeek 过滤本周（周一至今天）的记录
func FilterByWeek(records []parser.Record, now time.Time) []parser.Record {
	loc := time.Now().Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	weekday := int(today.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := today.AddDate(0, 0, -(weekday - 1))
	end := today.Add(24 * time.Hour)
	return filterByRange(records, monday, end)
}

// FilterByMonth 过滤本月的记录
func FilterByMonth(records []parser.Record, now time.Time) []parser.Record {
	loc := time.Now().Location()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	end := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, loc)
	return filterByRange(records, firstDay, end)
}

// FilterByRange 过滤指定日期范围 [from, to] 的记录（含两端）
func FilterByRange(records []parser.Record, from, to time.Time) []parser.Record {
	loc := time.Now().Location()
	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, loc)
	end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, loc).Add(24 * time.Hour)
	return filterByRange(records, start, end)
}

func filterByRange(records []parser.Record, start, end time.Time) []parser.Record {
	loc := time.Now().Location()
	var filtered []parser.Record
	for _, r := range records {
		local := r.Timestamp.In(loc)
		if !local.Before(start) && local.Before(end) {
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

// SessionStats 会话级统计
type SessionStats struct {
	SessionID    string
	Project      string
	Model        string
	InputTokens  int
	OutputTokens int
	CacheWrite   int
	CacheRead    int
	TotalCost    float64
	RequestCount int
	FirstTime    time.Time
	LastTime     time.Time
}

// AggregateBySession 按会话聚合统计
func AggregateBySession(records []parser.Record, cfg *config.Config) []SessionStats {
	m := make(map[string]*SessionStats)
	for _, r := range records {
		ss, ok := m[r.SessionID]
		if !ok {
			ss = &SessionStats{
				SessionID: r.SessionID,
				Project:   r.Project,
				Model:     r.Model,
				FirstTime: r.Timestamp,
				LastTime:  r.Timestamp,
			}
			m[r.SessionID] = ss
		}
		ss.InputTokens += r.Usage.InputTokens
		ss.OutputTokens += r.Usage.OutputTokens
		ss.CacheWrite += r.Usage.CacheCreationInputTokens
		ss.CacheRead += r.Usage.CacheReadInputTokens
		ss.TotalCost += cfg.CalculateCost(r.Model, r.Usage.InputTokens, r.Usage.OutputTokens, r.Usage.CacheCreationInputTokens, r.Usage.CacheReadInputTokens)
		ss.RequestCount++
		if r.Timestamp.Before(ss.FirstTime) {
			ss.FirstTime = r.Timestamp
		}
		if r.Timestamp.After(ss.LastTime) {
			ss.LastTime = r.Timestamp
		}
	}

	result := make([]SessionStats, 0, len(m))
	for _, ss := range m {
		result = append(result, *ss)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastTime.After(result[j].LastTime)
	})
	return result
}

// FilterByProject 过滤指定项目名（模糊匹配）的记录
func FilterByProject(records []parser.Record, project string) []parser.Record {
	var filtered []parser.Record
	for _, r := range records {
		if strings.Contains(strings.ToLower(r.Project), strings.ToLower(project)) {
			filtered = append(filtered, r)
		}
	}
	return filtered
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
