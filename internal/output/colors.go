package output

import "github.com/fatih/color"

// Color theme definitions for consistent terminal styling.

var (
	// HeaderStyle for section headers.
	HeaderStyle = color.New(color.FgCyan, color.Bold)
	// LabelStyle for field labels.
	LabelStyle = color.New(color.FgWhite, color.Bold)
	// GoodStyle for positive values.
	GoodStyle = color.New(color.FgGreen)
	// WarnStyle for warning values.
	WarnStyle = color.New(color.FgYellow)
	// BadStyle for negative/critical values.
	BadStyle = color.New(color.FgRed)
	// DimStyle for de-emphasized text.
	DimStyle = color.New(color.FgHiBlack)
	// SuccessStyle for success indicators.
	SuccessStyle = color.New(color.FgGreen, color.Bold)
	// ErrorStyle for error indicators.
	ErrorStyle = color.New(color.FgRed, color.Bold)
)

// Severity icons.
const (
	IconCritical = "🔴"
	IconWarning  = "🟡"
	IconInfo     = "🔵"
	IconSuccess  = "✅"
	IconFail     = "❌"
	IconRisk     = "⚠️"
	IconScore    = "🏆"
	IconAnalyze  = "🔍"
	IconTime     = "⏱"
	IconWorkload = "📦"
	IconChart    = "📊"
	IconInsight  = "💡"
	IconTarget   = "🎯"
)
