package display

import (
	"fmt"
	"strings"
	"time"

	"cctok/internal/stats"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true)
	projectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	costStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	totalStyle   = lipgloss.NewStyle().Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

// RenderToday 渲染按项目汇总的表格
func RenderToday(summary *stats.Summary) {
	renderProjectTable("", summary)
}

func renderSourceHeader(summaries []stats.SourceSummary, timeRange string) {
	var grandTotal float64
	for _, s := range summaries {
		grandTotal += s.Summary.TotalCost
	}
	fmt.Printf("\n  %s  Total: %s\n",
		headerStyle.Render(timeRange),
		costStyle.Render(formatCost(grandTotal)))
}

// RenderTodayBySource 渲染按工具分区的项目汇总表格
func RenderTodayBySource(summaries []stats.SourceSummary, timeRange string) {
	renderSourceHeader(summaries, timeRange)
	for _, s := range summaries {
		renderProjectTable(s.Source, s.Summary)
	}
}

func renderProjectTable(source string, summary *stats.Summary) {
	if source != "" {
		fmt.Printf("\n  %s\n\n", dimStyle.Render("── "+source+" ──"))
	} else {
		fmt.Printf("\n")
	}

	headers := []string{"Project", "Input", "Output", "Cache W", "Cache R", "Cost"}
	widths := []int{30, 10, 10, 10, 10, 10}

	printRow(headers, widths, headerStyle)
	printSep(widths)

	for _, p := range summary.Projects {
		row := []string{
			projectStyle.Render(truncate(p.Name, widths[0])),
			formatTokens(p.InputTokens),
			formatTokens(p.OutputTokens),
			formatTokens(p.CacheWrite),
			formatTokens(p.CacheRead),
			costStyle.Render(formatCost(p.TotalCost)),
		}
		printRawRow(row, widths)
	}

	printSep(widths)
	totalRow := []string{
		totalStyle.Render("Total"),
		formatTokens(summary.TotalInput),
		formatTokens(summary.TotalOutput),
		"",
		"",
		costStyle.Render(formatCost(summary.TotalCost)),
	}
	printRawRow(totalRow, widths)
}

// RenderModels 渲染按模型汇总的表格
func RenderModels(summary *stats.Summary) {
	renderModelTable("", summary)
}

// RenderModelsBySource 渲染按工具分区的模型汇总表格
func RenderModelsBySource(summaries []stats.SourceSummary, timeRange string) {
	renderSourceHeader(summaries, timeRange)
	for _, s := range summaries {
		renderModelTable(s.Source, s.Summary)
	}
}

func renderModelTable(source string, summary *stats.Summary) {
	if source != "" {
		fmt.Printf("\n  %s\n\n", dimStyle.Render("── "+source+" ──"))
	} else {
		fmt.Printf("\n")
	}

	headers := []string{"Model", "Reqs", "Input", "Output", "Cache W", "Cache R", "Cost"}
	widths := []int{24, 6, 10, 10, 10, 10, 10}

	printRow(headers, widths, headerStyle)
	printSep(widths)

	for _, m := range summary.Models {
		row := []string{
			projectStyle.Render(truncate(m.Name, widths[0])),
			fmt.Sprintf("%d", m.RequestCount),
			formatTokens(m.InputTokens),
			formatTokens(m.OutputTokens),
			formatTokens(m.CacheWrite),
			formatTokens(m.CacheRead),
			costStyle.Render(formatCost(m.TotalCost)),
		}
		printRawRow(row, widths)
	}

	printSep(widths)
	totalRow := []string{
		totalStyle.Render("Total"),
		"",
		formatTokens(summary.TotalInput),
		formatTokens(summary.TotalOutput),
		"",
		"",
		costStyle.Render(formatCost(summary.TotalCost)),
	}
	printRawRow(totalRow, widths)
}

func formatTokens(n int) string {
	if n == 0 {
		return dimStyle.Render("-")
	}
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func formatCost(c float64) string {
	return fmt.Sprintf("$%.2f", c)
}

func truncate(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// 逐 rune 截断，按显示宽度判断
	result := []rune(s)
	for i := len(result); i > 0; i-- {
		candidate := string(result[:i]) + "…"
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
	}
	return "…"
}

func printRow(cols []string, widths []int, style lipgloss.Style) {
	var parts []string
	for i, col := range cols {
		parts = append(parts, style.Render(padRight(col, widths[i])))
	}
	fmt.Printf("  %s\n", strings.Join(parts, "  "))
}

func printRawRow(cols []string, widths []int) {
	var parts []string
	for i, col := range cols {
		parts = append(parts, padRight(col, widths[i]))
	}
	fmt.Printf("  %s\n", strings.Join(parts, "  "))
}

func printSep(widths []int) {
	var parts []string
	for _, w := range widths {
		parts = append(parts, dimStyle.Render(strings.Repeat("─", w)))
	}
	fmt.Printf("  %s\n", strings.Join(parts, "──"))
}

func padRight(s string, width int) string {
	visLen := lipgloss.Width(s)
	if visLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visLen)
}

// RenderSessions 渲染按会话汇总的表格
func RenderSessions(sessions []stats.SessionStats, label string) {
	var totalCost float64
	for _, s := range sessions {
		totalCost += s.TotalCost
	}

	fmt.Printf("\n  %s  Sessions: %d  Total: %s\n\n",
		headerStyle.Render(label),
		len(sessions),
		costStyle.Render(formatCost(totalCost)))

	headers := []string{"Time", "Project", "Model", "Reqs", "Output", "Cost"}
	widths := []int{16, 28, 22, 6, 10, 10}

	printRow(headers, widths, headerStyle)
	printSep(widths)

	loc := time.Now().Location()
	for _, s := range sessions {
		row := []string{
			s.LastTime.In(loc).Format("01-02 15:04"),
			projectStyle.Render(truncate(s.Project, widths[1])),
			truncate(s.Model, widths[2]),
			fmt.Sprintf("%d", s.RequestCount),
			formatTokens(s.OutputTokens),
			costStyle.Render(formatCost(s.TotalCost)),
		}
		printRawRow(row, widths)
	}
	fmt.Println()
}
