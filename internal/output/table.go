package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// PrintComparisonTable prints a current vs recommended config comparison table.
func PrintComparisonTable(rows []types.ComparisonRow) {
	if len(rows) == 0 {
		return
	}

	header := color.New(color.FgCyan, color.Bold)
	header.Println("\n  📊 Current vs Recommended Configuration:")
	fmt.Println()

	// Calculate column widths
	aspectW, currentW, recW, deltaW := 14, 10, 13, 10
	for _, row := range rows {
		if len(row.Aspect) > aspectW {
			aspectW = len(row.Aspect)
		}
		if len(row.CurrentValue) > currentW {
			currentW = len(row.CurrentValue)
		}
		if len(row.Recommended) > recW {
			recW = len(row.Recommended)
		}
		if len(row.Delta) > deltaW {
			deltaW = len(row.Delta)
		}
	}
	aspectW += 2
	currentW += 2
	recW += 2
	deltaW += 2

	// Header
	sep := fmt.Sprintf("     ├%s┼%s┼%s┼%s┤",
		strings.Repeat("─", aspectW), strings.Repeat("─", currentW),
		strings.Repeat("─", recW), strings.Repeat("─", deltaW))
	topBorder := fmt.Sprintf("     ┌%s┬%s┬%s┬%s┐",
		strings.Repeat("─", aspectW), strings.Repeat("─", currentW),
		strings.Repeat("─", recW), strings.Repeat("─", deltaW))
	bottomBorder := fmt.Sprintf("     └%s┴%s┴%s┴%s┘",
		strings.Repeat("─", aspectW), strings.Repeat("─", currentW),
		strings.Repeat("─", recW), strings.Repeat("─", deltaW))

	fmt.Println(topBorder)
	headerLabel := color.New(color.FgWhite, color.Bold)
	fmt.Print("     │")
	headerLabel.Printf(" %-*s", aspectW-1, "Aspect")
	fmt.Print("│")
	headerLabel.Printf(" %-*s", currentW-1, "Current")
	fmt.Print("│")
	headerLabel.Printf(" %-*s", recW-1, "Recommended")
	fmt.Print("│")
	headerLabel.Printf(" %-*s", deltaW-1, "Delta")
	fmt.Println("│")
	fmt.Println(sep)

	// Data rows
	for _, row := range rows {
		fmt.Printf("     │ %-*s│ %-*s│ %-*s│ %-*s│\n",
			aspectW-1, row.Aspect,
			currentW-1, row.CurrentValue,
			recW-1, row.Recommended,
			deltaW-1, row.Delta)
	}
	fmt.Println(bottomBorder)
	fmt.Println()
}

// PrintCorrelationTable prints correlation analysis results as a table.
func PrintCorrelationTable(correlations []types.CorrelationResult) {
	if len(correlations) == 0 {
		return
	}

	header := color.New(color.FgCyan, color.Bold)
	header.Println("\n  📈 Metric Correlation Analysis:")
	fmt.Println()

	for _, corr := range correlations {
		label := color.New(color.FgWhite)
		label.Printf("     %s ↔ %s: ", corr.MetricA, corr.MetricB)

		corrColor := getCorrelationColor(corr.Correlation)
		corrColor.Printf("%.2f", corr.Correlation)
		fmt.Println()

		dim := color.New(color.FgHiBlack)
		dim.Printf("       %s\n", corr.Interpretation)
		fmt.Println()
	}
}

// BuildComparisonRows builds comparison rows from current config and recommendation.
func BuildComparisonRows(rec *types.Recommendation, cpuReq, cpuLim, memReq, memLim float64) []types.ComparisonRow {
	var rows []types.ComparisonRow

	if cpuReq > 0 {
		rows = append(rows, types.ComparisonRow{
			Aspect:       "CPU Request",
			CurrentValue: formatCPUCores(cpuReq),
			Recommended:  rec.CPURequest,
			Delta:        formatDelta(cpuReq, rec.CPURequestRaw),
		})
	}

	if cpuLim > 0 {
		rows = append(rows, types.ComparisonRow{
			Aspect:       "CPU Limit",
			CurrentValue: formatCPUCores(cpuLim),
			Recommended:  rec.CPULimit,
			Delta:        formatDelta(cpuLim, rec.CPULimitRaw),
		})
	}

	if memReq > 0 {
		rows = append(rows, types.ComparisonRow{
			Aspect:       "Memory Request",
			CurrentValue: formatMemBytes(memReq),
			Recommended:  rec.MemoryRequest,
			Delta:        formatDelta(memReq, rec.MemRequestRaw),
		})
	}

	if memLim > 0 {
		rows = append(rows, types.ComparisonRow{
			Aspect:       "Memory Limit",
			CurrentValue: formatMemBytes(memLim),
			Recommended:  rec.MemoryLimit,
			Delta:        formatDelta(memLim, rec.MemLimitRaw),
		})
	}

	return rows
}

func formatCPUCores(cores float64) string {
	if cores < 1.0 {
		return fmt.Sprintf("%dm", int(cores*1000))
	}
	return fmt.Sprintf("%.1f", cores)
}

func formatMemBytes(bytes float64) string {
	gi := bytes / (1024 * 1024 * 1024)
	if gi >= 1.0 {
		return fmt.Sprintf("%.1fGi", gi)
	}
	mi := bytes / (1024 * 1024)
	return fmt.Sprintf("%.0fMi", mi)
}

func formatDelta(current, recommended float64) string {
	if current == 0 {
		return "N/A"
	}
	pctChange := ((recommended - current) / current) * 100

	if pctChange > 0 {
		return fmt.Sprintf("↑ +%.0f%%", pctChange)
	} else if pctChange < 0 {
		return fmt.Sprintf("↓ %.0f%%", pctChange)
	}
	return "→ 0%"
}

func getCorrelationColor(corr float64) *color.Color {
	abs := corr
	if abs < 0 {
		abs = -abs
	}
	switch {
	case abs > 0.7:
		return color.New(color.FgGreen, color.Bold)
	case abs > 0.4:
		return color.New(color.FgYellow)
	default:
		return color.New(color.FgHiBlack)
	}
}

// PrintSummaryLine prints a one-line summary suitable for list outputs.
func PrintSummaryLine(workload types.WorkloadInfo, overallScore float64) {
	scoreC := getScoreColor(overallScore)
	grade := scoreToGrade(overallScore)
	fmt.Printf("  %-12s %-30s %-10s  ", workload.Kind, workload.Name, workload.Namespace)
	scoreC.Printf("%.1f/10 (%s)\n", overallScore, grade)
}
