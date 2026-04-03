// Package form provides a reusable multi-field input component.
// Focus moves between fields with tab/shift-tab or ↑/↓.
// Validation is run on every keystroke of the focused field.
package form

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ValidateFn validates a field value and returns a non-nil error when invalid.
type ValidateFn func(string) error

var (
	labelFocused  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	labelBlurred  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	inputFocused  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("212")).Padding(0, 1)
	inputBlurred  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
)

// Field holds the state for a single form input.
type Field struct {
	Label    string
	input    textinput.Model
	validate ValidateFn
	err      error
}

// NewField creates a Field with the given label and placeholder.
// validate may be nil to skip validation.
func NewField(label, placeholder string, validate ValidateFn) Field {
	ti := textinput.New()
	ti.Placeholder = placeholder
	return Field{
		Label:    label,
		input:    ti,
		validate: validate,
	}
}

// Model manages a collection of focusable input fields.
type Model struct {
	fields  []Field
	focused int
}

// New returns a Model with the given fields; the first field is focused.
func New(fields ...Field) Model {
	m := Model{fields: make([]Field, len(fields))}
	copy(m.fields, fields)
	if len(m.fields) > 0 {
		m.fields[0].input.Focus()
	}
	return m
}

// Init returns the blink command needed to animate the text cursor.
func (m Model) Init() tea.Cmd {
	if len(m.fields) == 0 {
		return nil
	}
	return textinput.Blink
}

// focus transitions focus to field i, blurring all others.
func (m Model) focus(i int) (Model, tea.Cmd) {
	for j := range m.fields {
		if j == i {
			m.fields[j].input.Focus()
		} else {
			m.fields[j].input.Blur()
		}
	}
	m.focused = i
	return m, textinput.Blink
}

// Update handles navigation keys and delegates typing to the focused field.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if len(m.fields) == 0 {
		return m, nil
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyTab, tea.KeyDown:
			return m.focus((m.focused + 1) % len(m.fields))
		case tea.KeyShiftTab, tea.KeyUp:
			return m.focus((m.focused - 1 + len(m.fields)) % len(m.fields))
		}
	}

	// Delegate to the focused field's text input.
	var cmd tea.Cmd
	m.fields[m.focused].input, cmd = m.fields[m.focused].input.Update(msg)
	// Run validation on the active field after each update.
	if m.fields[m.focused].validate != nil {
		m.fields[m.focused].err = m.fields[m.focused].validate(m.fields[m.focused].input.Value())
	}
	return m, cmd
}

// Valid returns true when all fields pass their validation functions.
func (m Model) Valid() bool {
	for _, f := range m.fields {
		if f.validate != nil && f.validate(f.input.Value()) != nil {
			return false
		}
	}
	return true
}

// Values returns a map of field label → current value.
func (m Model) Values() map[string]string {
	out := make(map[string]string, len(m.fields))
	for _, f := range m.fields {
		out[f.Label] = f.input.Value()
	}
	return out
}

// View renders all fields with their labels, inputs, and inline error text.
func (m Model) View() string {
	var b strings.Builder
	for i, f := range m.fields {
		if i == m.focused {
			b.WriteString(labelFocused.Render(f.Label) + "\n")
			b.WriteString(inputFocused.Render(f.input.View()) + "\n")
		} else {
			b.WriteString(labelBlurred.Render(f.Label) + "\n")
			b.WriteString(inputBlurred.Render(f.input.View()) + "\n")
		}
		if f.err != nil {
			b.WriteString(errorStyle.Render("  ✗ "+f.err.Error()) + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}
