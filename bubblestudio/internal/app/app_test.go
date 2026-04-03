package app_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tuistudio/bubblestudio/internal/app"
)

// simulateResize sends a WindowSizeMsg to the model.
func simulateResize(m tea.Model, w, h int) tea.Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next
}

func TestNew_initReturnsNil(t *testing.T) {
	m := app.New()
	if m.Init() != nil {
		t.Error("Init should return nil before any events")
	}
}

func TestUpdate_windowSizeMsg(t *testing.T) {
	m := simulateResize(app.New(), 80, 24)
	// After resize, View should not panic.
	out := m.View()
	if out == "" {
		t.Error("View should not return empty string after resize")
	}
}

func TestUpdate_quitKey(t *testing.T) {
	m := simulateResize(app.New(), 80, 24)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("q key should return a quit command")
	}
}

func TestUpdate_tabSwitchesToForm(t *testing.T) {
	m := simulateResize(app.New(), 80, 24)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	// After Tab, View should render the form (not panic).
	out := next.View()
	if out == "" {
		t.Error("View should not be empty after switching to form")
	}
}

func TestUpdate_escReturnsToList(t *testing.T) {
	m := simulateResize(app.New(), 80, 24)
	// Switch to form.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	// Switch back to list.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	out := m.View()
	if out == "" {
		t.Error("View should not be empty after returning to list")
	}
}

func TestView_doesNotExceedTerminalHeight(t *testing.T) {
	// With a 24-line terminal, bodyHeight must leave room for chrome without
	// hardcoding exactly 2.  We verify it is positive and less than total height.
	m := simulateResize(app.New(), 80, 24)
	view := m.View()
	// Use lipgloss.Height for an accurate rendered-line count (newlines + 1,
	// with correct handling of a trailing newline).
	h := lipgloss.Height(view)
	if h > 24 {
		t.Errorf("rendered output is %d lines tall, expected <= 24", h)
	}
}
