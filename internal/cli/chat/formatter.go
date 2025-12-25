package chat

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// TerminalFormatter handles pretty terminal output
type TerminalFormatter struct {
	glamourRenderer *glamour.TermRenderer
	noColor         bool
}

// NewTerminalFormatter creates a new formatter
func NewTerminalFormatter(noColor bool) (*TerminalFormatter, error) {
	var renderer *glamour.TermRenderer
	var err error
	
	if !noColor {
		renderer, err = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(100),
		)
		if err != nil {
			return nil, err
		}
	}
	
	return &TerminalFormatter{
		glamourRenderer: renderer,
		noColor:         noColor,
	}, nil
}

// FormatUserPrompt formats user input
func (f *TerminalFormatter) FormatUserPrompt(text string) string {
	if f.noColor {
		return "You: " + text
	}
	
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("36")). // Cyan
		Bold(true)
	
	return style.Render("You: ") + text
}

// FormatThinking formats the "thinking" indicator
func (f *TerminalFormatter) FormatThinking() string {
	if f.noColor {
		return "⏳ Thinking..."
	}
	
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")). // Dim gray
		Italic(true)
	
	return style.Render("⏳ Thinking...")
}

// FormatAssistantResponse formats AI response with markdown rendering
func (f *TerminalFormatter) FormatAssistantResponse(markdown string) string {
	if f.noColor || f.glamourRenderer == nil {
		// Fallback: just strip markdown fences
		return f.stripMarkdownFences(markdown)
	}
	
	// Render with Glamour
	rendered, err := f.glamourRenderer.Render(markdown)
	if err != nil {
		// Fallback on error
		return f.stripMarkdownFences(markdown)
	}
	
	return strings.TrimSpace(rendered)
}

// FormatToolExecution formats tool execution info
func (f *TerminalFormatter) FormatToolExecution(toolName string, args map[string]interface{}) string {
	if f.noColor {
		return "⚡ Executing: " + toolName
	}
	
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")). // Yellow
		Bold(true)
	
	return style.Render("⚡ " + toolName)
}

// FormatToolResult formats tool result
func (f *TerminalFormatter) FormatToolResult(result string) string {
	if f.noColor {
		return result
	}
	
	// Add subtle indent and border
	style := lipgloss.NewStyle().
		PaddingLeft(2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(lipgloss.Color("240"))
	
	return style.Render(result)
}

// FormatSuccess formats success message
func (f *TerminalFormatter) FormatSuccess(text string) string {
	if f.noColor {
		return "✓ " + text
	}
	
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")). // Green
		Bold(true)
	
	return style.Render("✓ " + text)
}

// FormatError formats error message
func (f *TerminalFormatter) FormatError(text string) string {
	if f.noColor {
		return "✗ " + text
	}
	
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red
		Bold(true)
	
	return style.Render("✗ " + text)
}

// stripMarkdownFences removes markdown code fences for plain output
func (f *TerminalFormatter) stripMarkdownFences(text string) string {
	// Remove ```language and ``` markers
	re := regexp.MustCompile("(?m)^```[a-z]*\\n")
	text = re.ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, "```", "")
	return text
}

// FormatSeparator creates a visual separator
func (f *TerminalFormatter) FormatSeparator() string {
	if f.noColor {
		return strings.Repeat("─", 60)
	}
	
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	
	return style.Render(strings.Repeat("─", 60))
}
