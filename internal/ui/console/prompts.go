package console

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Prompter handles user input prompts
type Prompter struct {
	scanner *bufio.Scanner
}

// NewPrompter creates a new prompter
func NewPrompter() *Prompter {
	return &Prompter{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// Prompt displays a prompt and returns user input
func (p *Prompter) Prompt(message string) (string, error) {
	fmt.Print(message)

	if !p.scanner.Scan() {
		if err := p.scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("input closed")
	}

	return p.scanner.Text(), nil
}

// PromptWithDefault displays a prompt with a default value
func (p *Prompter) PromptWithDefault(message, defaultValue string) (string, error) {
	prompt := fmt.Sprintf("%s [%s]: ", message, Dim(defaultValue))

	input, err := p.Prompt(prompt)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(input) == "" {
		return defaultValue, nil
	}

	return input, nil
}

// PromptYesNo displays a yes/no prompt
func (p *Prompter) PromptYesNo(message string, defaultYes bool) (bool, error) {
	var prompt string
	if defaultYes {
		prompt = fmt.Sprintf("%s [%s/%s]: ", message, Bold("Y"), "n")
	} else {
		prompt = fmt.Sprintf("%s [%s/%s]: ", message, "y", Bold("N"))
	}

	input, err := p.Prompt(prompt)
	if err != nil {
		return false, err
	}

	input = strings.ToLower(strings.TrimSpace(input))

	if input == "" {
		return defaultYes, nil
	}

	return input == "y" || input == "yes", nil
}

// PromptChoice displays a multiple choice prompt
func (p *Prompter) PromptChoice(message string, choices []string, defaultIndex int) (int, error) {
	fmt.Println(message)

	for i, choice := range choices {
		if i == defaultIndex {
			fmt.Printf("  %d) %s\n", i+1, Bold(choice))
		} else {
			fmt.Printf("  %d) %s\n", i+1, choice)
		}
	}

	prompt := fmt.Sprintf("Enter choice [%s]: ", Dim(fmt.Sprintf("%d", defaultIndex+1)))

	input, err := p.Prompt(prompt)
	if err != nil {
		return 0, err
	}

	if strings.TrimSpace(input) == "" {
		return defaultIndex, nil
	}

	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil {
		return 0, fmt.Errorf("invalid choice")
	}

	if choice < 1 || choice > len(choices) {
		return 0, fmt.Errorf("choice out of range")
	}

	return choice - 1, nil
}

// PromptPassword displays a password prompt (note: doesn't hide input in simple version)
func (p *Prompter) PromptPassword(message string) (string, error) {
	fmt.Print(Dim(message + ": "))

	if !p.scanner.Scan() {
		if err := p.scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("input closed")
	}

	return p.scanner.Text(), nil
}

// ConfirmAction asks for confirmation before performing an action
func (p *Prompter) ConfirmAction(action string) (bool, error) {
	return p.PromptYesNo(
		fmt.Sprintf("Are you sure you want to %s?", Yellow(action)),
		false,
	)
}

// PromptMultiline prompts for multiple lines of input
func (p *Prompter) PromptMultiline(message, endMarker string) (string, error) {
	fmt.Printf("%s (end with '%s' on a new line):\n", message, endMarker)

	var lines []string

	for p.scanner.Scan() {
		line := p.scanner.Text()

		if line == endMarker {
			break
		}

		lines = append(lines, line)
	}

	if err := p.scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}
