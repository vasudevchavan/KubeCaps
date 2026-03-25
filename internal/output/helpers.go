package output

import (
	"github.com/fatih/color"
)

// Separator returns a visual separator line
func Separator() string {
	return "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

// NewLabelPrinter returns a color printer for labels
func NewLabelPrinter() *color.Color {
	return color.New(color.FgWhite, color.Bold)
}

// NewValuePrinter returns a color printer for values
func NewValuePrinter() *color.Color {
	return color.New(color.FgCyan)
}

// GetSeverityColor returns a color based on severity level
func GetSeverityColor(severity string) *color.Color {
	switch severity {
	case "critical":
		return color.New(color.FgRed, color.Bold)
	case "high":
		return color.New(color.FgRed)
	case "medium":
		return color.New(color.FgYellow)
	case "low":
		return color.New(color.FgGreen)
	default:
		return color.New(color.FgWhite)
	}
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	green := color.New(color.FgGreen, color.Bold)
	green.Printf("✅ %s\n", message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	yellow := color.New(color.FgYellow, color.Bold)
	yellow.Printf("⚠️  %s\n", message)
}

// PrintError prints an error message
func PrintError(message string) {
	red := color.New(color.FgRed, color.Bold)
	red.Printf("❌ %s\n", message)
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	cyan := color.New(color.FgCyan)
	cyan.Printf("ℹ️  %s\n", message)
}

// Made with Bob
