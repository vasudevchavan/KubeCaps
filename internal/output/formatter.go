package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// PrintRecommendation prints a resource recommendation in a formatted way.
func PrintRecommendation(rec *types.Recommendation) {
	header := color.New(color.FgCyan, color.Bold)
	label := color.New(color.FgWhite, color.Bold)
	good := color.New(color.FgGreen)
	warn := color.New(color.FgYellow)

	header.Printf("━━━ %s/%s (%s) ", rec.WorkloadKind, rec.WorkloadName, rec.Namespace)
	header.Println(strings.Repeat("━", 40))

	fmt.Println()
	label.Print("  📊 Recommended Resources (per replica):\n")
	fmt.Println()

	// CPU
	label.Print("     CPU Request:  ")
	good.Printf("%-10s", rec.CPURequest)
	formatDiff(rec.CurrentCPURequestRaw, rec.CPURequestRaw, rec.CPUReqDiffPercent)
	
	label.Print("     CPU Limit:    ")
	good.Printf("%-10s", rec.CPULimit)
	formatDiff(rec.CurrentCPULimitRaw, rec.CPULimitRaw, rec.CPULimitDiffPercent)

	// Memory
	label.Print("     Mem Request:  ")
	good.Printf("%-10s", rec.MemoryRequest)
	formatDiff(rec.CurrentMemRequestRaw, rec.MemRequestRaw, rec.MemReqDiffPercent)

	label.Print("     Mem Limit:    ")
	good.Printf("%-10s", rec.MemoryLimit)
	formatDiff(rec.CurrentMemLimitRaw, rec.MemLimitRaw, rec.MemLimitDiffPercent)

	fmt.Println()

	// Confidence
	label.Print("  🎯 Confidence:   ")
	confPct := rec.Confidence * 100
	if confPct >= 80 {
		good.Printf("%.0f%%\n", confPct)
	} else {
		warn.Printf("%.0f%%\n", confPct)
	}

	// Time window
	label.Printf("  ⏱  Time Window:  %dh\n", rec.TimeWindowHours)

	// Explanation / Insights
	if len(rec.Insights) > 0 {
		fmt.Println()
		label.Println("  💡 Analysis & Optimization Strategy:")
		infoColor := color.New(color.FgCyan)
		
		infoColor.Printf("     Classification: %s\n", rec.Type)
		if rec.Recommendations.HPA.Enabled {
			infoColor.Printf("     [HPA] Enable with minReplicas=%d, maxReplicas=%d on CPU=%s\n", rec.Recommendations.HPA.MinReplicas, rec.Recommendations.HPA.MaxReplicas, rec.Recommendations.HPA.TargetValue)
		} else {
			dim := color.New(color.FgHiBlack)
			dim.Println("     [HPA] Disabled (Workload is steady, no high-variance scaling needed)")
		}

		infoColor.Printf("     [VPA] Set Mode=%s (CPU Buffers = %s, Mem Buffers = %s)\n", rec.Recommendations.VPA.Mode, rec.Recommendations.VPA.CPU, rec.Recommendations.VPA.Memory)

		if rec.Recommendations.KEDA.Enabled {
			infoColor.Printf("     [KEDA] Recommend Trigger='%s' at threshold %s (cooldown=%ds)\n", rec.Recommendations.KEDA.Trigger, rec.Recommendations.KEDA.Threshold, rec.Recommendations.KEDA.CooldownPeriod)
		}

		fmt.Println()
		for _, ins := range rec.Insights {
			fmt.Printf("     • %s\n", ins)
		}
	}

	// Risk Profile
	fmt.Println()
	label.Println("  ⚠️  Risk Profile:")
	riskColor := color.New(color.FgGreen)
	if rec.Risk.Level == "high" {
		riskColor = color.New(color.FgRed, color.Bold)
	} else if rec.Risk.Level == "medium" {
		riskColor = color.New(color.FgMagenta)
	}
	
	riskColor.Printf("     Level: %s\n", rec.Risk.Level)
	for _, note := range rec.Risk.Notes {
		color.New(color.FgHiBlack).Printf("     - %s\n", note)
	}

	fmt.Println()
}

// formatDiff prints the current value and percentage diff.
func formatDiff(current, rec, diffPct float64) {
	dim := color.New(color.FgHiBlack)
	up := color.New(color.FgRed)
	down := color.New(color.FgGreen)
	
	if current == 0 {
		dim.Println("(currently not set)")
		return
	}

	if diffPct > 10 {
		up.Printf("(current: %v, +%.0f%%)\n", formatValue(current), diffPct)
	} else if diffPct < -10 {
		down.Printf("(current: %v, %.0f%%)\n", formatValue(current), diffPct)
	} else {
		dim.Printf("(current: %v, ~%.0f%%)\n", formatValue(current), diffPct)
	}
}

// formatValue formats a raw CPU/Mem float to string logic similar to engine.go
func formatValue(val float64) string {
	// If it's small, it's CPU, but memory is in bytes, so > 1M usually
	if val < 1000 {
		// CPU logic
		if val < 1.0 {
			return fmt.Sprintf("%dm", int(val*1000))
		}
		return fmt.Sprintf("%.1f", val)
	}
	// Memory logic
	gi := val / (1024 * 1024 * 1024)
	if gi >= 1.0 {
		return fmt.Sprintf("%.1fGi", gi)
	}
	mi := val / (1024 * 1024)
	return fmt.Sprintf("%.0fMi", mi)
}


// PrintEvaluationResult prints a complete evaluation result.
func PrintEvaluationResult(result *types.EvaluationResult) {
	header := color.New(color.FgCyan, color.Bold)
	label := color.New(color.FgWhite, color.Bold)

	header.Printf("\n╔══════════════════════════════════════════════════════════════╗\n")
	header.Printf("║  KubeCaps Autoscaling Evaluation Report                     ║\n")
	header.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")

	// Workload info
	label.Printf("  📦 Workload:    %s/%s (%s)\n", result.Workload.Kind, result.Workload.Name, result.Workload.Namespace)
	label.Printf("  📅 Analyzed:    %s\n", result.AnalyzedAt.Format("2006-01-02 15:04:05"))
	label.Printf("  ⏱  Window:      %s\n", result.TimeWindow)
	fmt.Println()

	// Overall score
	printOverallScore(result.OverallScore)
	fmt.Println()

	// Component scores
	label.Println("  📊 Component Scores:")
	fmt.Println()
	for _, cs := range result.Scores {
		printComponentScore(cs)
	}
	fmt.Println()

	// Autoscaling configs detected
	printDetectedConfigs(result)

	// Insights
	if len(result.Insights) > 0 {
		label.Println("  💡 Insights:")
		fmt.Println()
		for _, insight := range result.Insights {
			printInsight(insight)
		}
		fmt.Println()
	}

	// Risks
	if len(result.Risks) > 0 {
		label.Println("  ⚠️  Risks:")
		fmt.Println()
		for _, risk := range result.Risks {
			printRisk(risk)
		}
		fmt.Println()
	}

	// Recommendation
	if result.Recommendation != nil {
		label.Println("  🎯 Resource Recommendation:")
		fmt.Println()
		PrintRecommendation(result.Recommendation)
	}

	header.Println(strings.Repeat("━", 64))
}

func printOverallScore(score float64) {
	label := color.New(color.FgWhite, color.Bold)
	label.Print("  🏆 Overall Optimization Score:  ")

	scoreColor := getScoreColor(score)
	grade := scoreToGrade(score)
	scoreColor.Printf("%.1f/10 (%s)\n", score, grade)

	// Score bar
	fmt.Print("     ")
	filled := int(score)
	empty := 10 - filled
	green := color.New(color.FgGreen)
	dim := color.New(color.FgHiBlack)
	for i := 0; i < filled; i++ {
		green.Print("█")
	}
	for i := 0; i < empty; i++ {
		dim.Print("░")
	}
	fmt.Println()
}

func printComponentScore(cs types.ComponentScore) {
	label := color.New(color.FgWhite)
	scoreColor := getScoreColor(cs.Score)

	if cs.Score == 0 {
		dim := color.New(color.FgHiBlack)
		label.Printf("     %-6s ", cs.Component)
		dim.Printf("Not configured\n")
		return
	}

	label.Printf("     %-6s ", cs.Component)
	scoreColor.Printf("%.1f/%.0f", cs.Score, cs.MaxScore)
	dim := color.New(color.FgHiBlack)
	dim.Printf("  %s\n", cs.Details)
}

func printDetectedConfigs(result *types.EvaluationResult) {
	label := color.New(color.FgWhite, color.Bold)
	label.Println("  🔍 Detected Autoscaling:")
	fmt.Println()

	check := color.New(color.FgGreen)
	cross := color.New(color.FgRed)

	if result.HPAConfig != nil && result.HPAConfig.Found {
		check.Printf("     ✓ HPA: %s (min=%d, max=%d, current=%d)\n",
			result.HPAConfig.Name, result.HPAConfig.MinReplicas,
			result.HPAConfig.MaxReplicas, result.HPAConfig.CurrentReplicas)
	} else {
		cross.Println("     ✗ HPA: Not configured")
	}

	if result.VPAConfig != nil && result.VPAConfig.Found {
		check.Printf("     ✓ VPA: %s (mode=%s)\n",
			result.VPAConfig.Name, result.VPAConfig.UpdateMode)
	} else {
		cross.Println("     ✗ VPA: Not configured")
	}

	if result.KEDAConfig != nil && result.KEDAConfig.Found {
		triggerTypes := []string{}
		for _, t := range result.KEDAConfig.Triggers {
			triggerTypes = append(triggerTypes, t.Type)
		}
		check.Printf("     ✓ KEDA: %s (triggers: %s)\n",
			result.KEDAConfig.Name, strings.Join(triggerTypes, ", "))
	} else {
		cross.Println("     ✗ KEDA: Not configured")
	}
	fmt.Println()
}

func printInsight(insight types.Insight) {
	var icon string
	var c *color.Color

	switch insight.Severity {
	case "critical":
		icon = "🔴"
		c = color.New(color.FgRed)
	case "warning":
		icon = "🟡"
		c = color.New(color.FgYellow)
	default:
		icon = "🔵"
		c = color.New(color.FgCyan)
	}

	c.Printf("     %s [%s] %s\n", icon, strings.ToUpper(insight.Severity), insight.Title)
	fmt.Printf("        %s\n", insight.Description)
	if insight.Evidence != "" {
		dim := color.New(color.FgHiBlack)
		dim.Printf("        Evidence: %s\n", insight.Evidence)
	}
	if insight.Suggestion != "" {
		fmt.Printf("        → %s\n", insight.Suggestion)
	}
	fmt.Println()
}

func printRisk(risk types.RiskFlag) {
	var icon string
	var c *color.Color

	switch risk.Level {
	case "high":
		icon = "🚨"
		c = color.New(color.FgRed, color.Bold)
	case "medium":
		icon = "⚠️"
		c = color.New(color.FgYellow)
	default:
		icon = "ℹ️"
		c = color.New(color.FgCyan)
	}

	c.Printf("     %s [%s] %s\n", icon, strings.ToUpper(risk.Level), risk.Type)
	fmt.Printf("        %s\n", risk.Description)
	if risk.Impact != "" {
		fmt.Printf("        Impact: %s\n", risk.Impact)
	}
}

func getScoreColor(score float64) *color.Color {
	switch {
	case score >= 8:
		return color.New(color.FgGreen, color.Bold)
	case score >= 6:
		return color.New(color.FgYellow, color.Bold)
	default:
		return color.New(color.FgRed, color.Bold)
	}
}

func scoreToGrade(score float64) string {
	switch {
	case score >= 9:
		return "A+"
	case score >= 8:
		return "A"
	case score >= 7:
		return "B"
	case score >= 6:
		return "C"
	case score >= 5:
		return "D"
	default:
		return "F"
	}
}
