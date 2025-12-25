package console

import "fmt"

// ANSI color codes
const (
	ColorReset   = "\033[0m"
	ColorBold    = "\033[1m"
	ColorDim     = "\033[2m"
	ColorItalic  = "\033[3m"
	
	// Foreground colors
	ColorBlack   = "\033[30m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	
	// Bright foreground colors
	ColorBrightBlack   = "\033[90m"
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"
)

var colorsEnabled = true

// SetColorsEnabled enables or disables color output
func SetColorsEnabled(enabled bool) {
	colorsEnabled = enabled
}

// Colorize wraps text in color codes
func Colorize(text, color string) string {
	if !colorsEnabled {
		return text
	}
	return color + text + ColorReset
}

// Bold makes text bold
func Bold(text string) string {
	return Colorize(text, ColorBold)
}

// Dim makes text dimmed
func Dim(text string) string {
	return Colorize(text, ColorDim)
}

// Italic makes text italic
func Italic(text string) string {
	return Colorize(text, ColorItalic)
}

// Red colors text red
func Red(text string) string {
	return Colorize(text, ColorRed)
}

// Green colors text green
func Green(text string) string {
	return Colorize(text, ColorGreen)
}

// Yellow colors text yellow
func Yellow(text string) string {
	return Colorize(text, ColorYellow)
}

// Blue colors text blue
func Blue(text string) string {
	return Colorize(text, ColorBlue)
}

// Magenta colors text magenta
func Magenta(text string) string {
	return Colorize(text, ColorMagenta)
}

// Cyan colors text cyan
func Cyan(text string) string {
	return Colorize(text, ColorCyan)
}

// BrightBlue colors text bright blue
func BrightBlue(text string) string {
	return Colorize(text, ColorBrightBlue)
}

// BrightGreen colors text bright green
func BrightGreen(text string) string {
	return Colorize(text, ColorBrightGreen)
}

// BrightYellow colors text bright yellow
func BrightYellow(text string) string {
	return Colorize(text, ColorBrightYellow)
}

// BrightCyan colors text bright cyan
func BrightCyan(text string) string {
	return Colorize(text, ColorBrightCyan)
}

// Success formats success messages
func Success(text string) string {
	return Green("✓ " + text)
}

// Error formats error messages
func Error(text string) string {
	return Red("✗ " + text)
}

// Warning formats warning messages
func Warning(text string) string {
	return Yellow("⚠ " + text)
}

// Info formats info messages
func Info(text string) string {
	return Blue("ℹ " + text)
}

// PrintSuccess prints a success message
func PrintSuccess(format string, args ...interface{}) {
	fmt.Println(Success(fmt.Sprintf(format, args...)))
}

// PrintError prints an error message
func PrintError(format string, args ...interface{}) {
	fmt.Println(Error(fmt.Sprintf(format, args...)))
}

// PrintWarning prints a warning message
func PrintWarning(format string, args ...interface{}) {
	fmt.Println(Warning(fmt.Sprintf(format, args...)))
}

// PrintInfo prints an info message
func PrintInfo(format string, args ...interface{}) {
	fmt.Println(Info(fmt.Sprintf(format, args...)))
}
