// Package app is the root Bubble Tea model for the tui-factory shell.
package app

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tuistudio/bubblestudio/internal/keymap"
	"github.com/tuistudio/bubblestudio/internal/theme"
)

// Model is the root application model.
type Model struct {
	keys     keymap.KeyMap
	theme    theme.Theme
	help     help.Model
	width    int
	height   int
	showHelp bool
}

// New returns a ready-to-use Model.
func New() Model {
	return Model{
		keys:  keymap.Default(),
		theme: theme.Default(),
		help:  help.New(),
	}
}

// Init satisfies tea.Model; no I/O commands on startup.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
		}
	}
	return m, nil
}

// View renders the header / body / footer layout.
func (m Model) View() string {
	header := m.theme.Header.Render("tui-factory")

	var body string
	if m.showHelp {
		body = m.theme.Body.Render(m.help.View(m.keys))
	} else {
		body = m.theme.Body.Render("")
	}

	footer := m.theme.Footer.Render("q quit • ? help")

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}
