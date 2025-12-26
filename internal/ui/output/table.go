package output

import (
	"fmt"
	"io"
	"strings"
)

// Table represents a text table
type Table struct {
	headers []string
	rows    [][]string
	writer  io.Writer
}

// NewTable creates a new table
func NewTable(writer io.Writer, headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    make([][]string, 0),
		writer:  writer,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(cells ...string) {
	t.rows = append(t.rows, cells)
}

// Render renders the table
func (t *Table) Render() error {
	if len(t.headers) == 0 {
		return fmt.Errorf("no headers defined")
	}
	
	// Calculate column widths
	widths := make([]int, len(t.headers))
	for i, header := range t.headers {
		widths[i] = len(header)
	}
	
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	
	// Render header
	if err := t.renderRow(t.headers, widths); err != nil {
		return err
	}
	
	// Render separator
	if err := t.renderSeparator(widths); err != nil {
		return err
	}
	
	// Render rows
	for _, row := range t.rows {
		if err := t.renderRow(row, widths); err != nil {
			return err
		}
	}
	
	return nil
}

func (t *Table) renderRow(cells []string, widths []int) error {
	var parts []string
	
	for i, cell := range cells {
		width := widths[i]
		if i >= len(widths) {
			width = len(cell)
		}
		parts = append(parts, padRight(cell, width))
	}
	
	_, err := fmt.Fprintf(t.writer, "| %s |\n", strings.Join(parts, " | "))
	return err
}

func (t *Table) renderSeparator(widths []int) error {
	var parts []string
	
	for _, width := range widths {
		parts = append(parts, strings.Repeat("-", width))
	}
	
	_, err := fmt.Fprintf(t.writer, "|-%s-|\n", strings.Join(parts, "-|-"))
	return err
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// List represents a formatted list
type List struct {
	items  []string
	writer io.Writer
	style  ListStyle
}

// ListStyle defines list styling
type ListStyle string

const (
	ListStyleBullet  ListStyle = "bullet"
	ListStyleNumbered ListStyle = "numbered"
	ListStyleDash    ListStyle = "dash"
)

// NewList creates a new list
func NewList(writer io.Writer, style ListStyle) *List {
	return &List{
		items:  make([]string, 0),
		writer: writer,
		style:  style,
	}
}

// Add adds an item to the list
func (l *List) Add(item string) {
	l.items = append(l.items, item)
}

// Render renders the list
func (l *List) Render() error {
	for i, item := range l.items {
		var prefix string
		
		switch l.style {
		case ListStyleBullet:
			prefix = "• "
		case ListStyleNumbered:
			prefix = fmt.Sprintf("%d. ", i+1)
		case ListStyleDash:
			prefix = "- "
		}
		
		if _, err := fmt.Fprintf(l.writer, "%s%s\n", prefix, item); err != nil {
			return err
		}
	}
	
	return nil
}

// KeyValue represents a key-value display
type KeyValue struct {
	pairs  [][2]string
	writer io.Writer
}

// NewKeyValue creates a new key-value display
func NewKeyValue(writer io.Writer) *KeyValue {
	return &KeyValue{
		pairs:  make([][2]string, 0),
		writer: writer,
	}
}

// Add adds a key-value pair
func (kv *KeyValue) Add(key, value string) {
	kv.pairs = append(kv.pairs, [2]string{key, value})
}

// Render renders the key-value pairs
func (kv *KeyValue) Render() error {
	// Find max key width
	maxWidth := 0
	for _, pair := range kv.pairs {
		if len(pair[0]) > maxWidth {
			maxWidth = len(pair[0])
		}
	}
	
	// Render pairs
	for _, pair := range kv.pairs {
		paddedKey := padRight(pair[0], maxWidth)
		if _, err := fmt.Fprintf(kv.writer, "%s: %s\n", paddedKey, pair[1]); err != nil {
			return err
		}
	}
	
	return nil
}
