package display

import (
	"fmt"
	"strings"

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
	fmt.Printf("\n  %s  Total: %s\n\n",
		headerStyle.Render(summary.TimeRange),
		costStyle.Render(formatCost(summary.TotalCost)))

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
	fmt.Println()
}

// RenderModels 渲染按模型汇总的表格
func RenderModels(summary *stats.Summary) {
	fmt.Printf("\n  %s  Total: %s\n\n",
		headerStyle.Render(summary.TimeRange),
		costStyle.Render(formatCost(summary.TotalCost)))

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
	fmt.Println()
}

func formatTokens(n int) string {
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
	if len(s) <= maxWidth {
		return s
	}
	return s[:maxWidth-1] + "…"
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
