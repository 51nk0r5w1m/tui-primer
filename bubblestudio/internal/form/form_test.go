package form_test

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tuistudio/bubblestudio/internal/form"
)

// requireField is erroring on an empty string.
var requireField form.ValidateFn = func(v string) error {
	if v == "" {
		return errors.New("required")
	}
	return nil
}

func TestNew_emptyFields(t *testing.T) {
	m := form.New()
	if !m.Valid() {
		t.Error("empty form should be valid")
	}
	if m.Init() != nil {
		t.Error("Init on empty form should return nil")
	}
}

func TestNew_focusesFirstField(t *testing.T) {
	f1 := form.NewField("A", "ph", nil)
	f2 := form.NewField("B", "ph", nil)
	m := form.New(f1, f2)
	// Type a character; it should land in the first (focused) field, not the second.
	charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	m, _ = m.Update(charMsg)
	vals := m.Values()
	if got := vals["A"]; got != "x" {
		t.Errorf("expected first field to receive input; got %q", got)
	}
	if got := vals["B"]; got != "" {
		t.Errorf("expected second field to remain empty; got %q", got)
	}
}

func TestValid_falseWhenRequired(t *testing.T) {
	f := form.NewField("Name", "", requireField)
	m := form.New(f)
	if m.Valid() {
		t.Error("form with empty required field should be invalid")
	}
}

func TestValues_returnedByLabel(t *testing.T) {
	f := form.NewField("Key", "ph", nil)
	m := form.New(f)
	vals := m.Values()
	if _, ok := vals["Key"]; !ok {
		t.Error("Values() should contain entry for label 'Key'")
	}
}

func TestUpdate_tabNavigates(t *testing.T) {
	f1 := form.NewField("A", "ph", nil)
	f2 := form.NewField("B", "ph", nil)
	m := form.New(f1, f2)

	// Pressing Tab should move focus to the second field.
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	m, _ = m.Update(tabMsg)

	// Typing a character after Tab should affect the second field, not the first.
	charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	m, _ = m.Update(charMsg)

	vals := m.Values()
	if got := vals["B"]; got != "x" {
		t.Errorf("expected second field to receive input after Tab; got %q", got)
	}
	if got := vals["A"]; got != "" {
		t.Errorf("expected first field to remain unchanged after Tab; got %q", got)
	}
}

func TestValidateFn_errorType(t *testing.T) {
	// ValidateFn must accept a func(string) error — verify compile-time contract.
	var fn form.ValidateFn = func(v string) error {
		if v == "" {
			return errors.New("empty")
		}
		return nil
	}
	if err := fn(""); err == nil {
		t.Error("validator should return non-nil error for empty string")
	}
	if err := fn("hello"); err != nil {
		t.Errorf("validator should return nil for non-empty string, got %v", err)
	}
}
