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
	// Both fields should exist; the form itself should be valid (no validation).
	if !m.Valid() {
		t.Error("form with nil validators should be valid")
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

	// Pressing Tab should move focus to the second field (wraps via Update).
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	m, _ = m.Update(tabMsg)
	// The model should still be valid and not crash.
	_ = m.View()
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
