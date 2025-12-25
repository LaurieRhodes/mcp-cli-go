package console

import (
	"fmt"
	"strings"
	"time"
)

// Spinner represents a loading spinner
type Spinner struct {
	message string
	frames  []string
	running bool
	done    chan bool
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		done:    make(chan bool),
	}
}

// Start starts the spinner
func (s *Spinner) Start() {
	s.running = true
	
	go func() {
		i := 0
		for s.running {
			frame := s.frames[i%len(s.frames)]
			fmt.Printf("\r%s %s", Cyan(frame), s.message)
			
			select {
			case <-s.done:
				return
			case <-time.After(100 * time.Millisecond):
				i++
			}
		}
	}()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	s.running = false
	s.done <- true
	fmt.Print("\r" + strings.Repeat(" ", len(s.message)+10) + "\r")
}

// Success stops the spinner with a success message
func (s *Spinner) Success(message string) {
	s.Stop()
	PrintSuccess(message)
}

// Fail stops the spinner with an error message
func (s *Spinner) Fail(message string) {
	s.Stop()
	PrintError(message)
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.message = message
}

// ProgressBar represents a progress bar
type ProgressBar struct {
	total   int
	current int
	width   int
	message string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, message string) *ProgressBar {
	return &ProgressBar{
		total:   total,
		current: 0,
		width:   40,
		message: message,
	}
}

// Update updates the progress bar
func (pb *ProgressBar) Update(current int) {
	pb.current = current
	pb.Render()
}

// Increment increments the progress bar
func (pb *ProgressBar) Increment() {
	pb.current++
	pb.Render()
}

// Render renders the progress bar
func (pb *ProgressBar) Render() {
	percent := float64(pb.current) / float64(pb.total)
	filled := int(percent * float64(pb.width))
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)
	
	fmt.Printf("\r%s [%s] %d/%d (%.1f%%)",
		pb.message,
		Green(bar),
		pb.current,
		pb.total,
		percent*100,
	)
	
	if pb.current >= pb.total {
		fmt.Println()
	}
}

// Complete marks the progress bar as complete
func (pb *ProgressBar) Complete() {
	pb.current = pb.total
	pb.Render()
}

// StepIndicator shows step-by-step progress
type StepIndicator struct {
	steps   []string
	current int
}

// NewStepIndicator creates a new step indicator
func NewStepIndicator(steps []string) *StepIndicator {
	return &StepIndicator{
		steps:   steps,
		current: 0,
	}
}

// Start starts the step indicator
func (si *StepIndicator) Start() {
	fmt.Println(Bold("Steps:"))
	for i, step := range si.steps {
		if i == 0 {
			fmt.Printf("  %s %s\n", Cyan("▶"), step)
		} else {
			fmt.Printf("  %s %s\n", Dim("○"), Dim(step))
		}
	}
}

// Next moves to the next step
func (si *StepIndicator) Next() {
	if si.current < len(si.steps) {
		si.current++
		si.render()
	}
}

// Complete marks a step as complete
func (si *StepIndicator) Complete() {
	si.current++
	si.render()
}

// Fail marks a step as failed
func (si *StepIndicator) Fail(err error) {
	si.renderFailed(err)
}

func (si *StepIndicator) render() {
	// Move cursor up
	fmt.Print("\033[" + fmt.Sprintf("%d", len(si.steps)) + "A")
	
	for i, step := range si.steps {
		if i < si.current {
			fmt.Printf("  %s %s\n", Green("✓"), step)
		} else if i == si.current {
			fmt.Printf("  %s %s\n", Cyan("▶"), step)
		} else {
			fmt.Printf("  %s %s\n", Dim("○"), Dim(step))
		}
	}
}

func (si *StepIndicator) renderFailed(err error) {
	// Move cursor up
	fmt.Print("\033[" + fmt.Sprintf("%d", len(si.steps)) + "A")
	
	for i, step := range si.steps {
		if i < si.current {
			fmt.Printf("  %s %s\n", Green("✓"), step)
		} else if i == si.current {
			fmt.Printf("  %s %s - %s\n", Red("✗"), step, Red(err.Error()))
		} else {
			fmt.Printf("  %s %s\n", Dim("○"), Dim(step))
		}
	}
}
