// Package app is the root Bubble Tea model for the bubblestudio shell.
package app

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tuistudio/bubblestudio/internal/demo"
	"github.com/tuistudio/bubblestudio/internal/form"
	"github.com/tuistudio/bubblestudio/internal/keymap"
	"github.com/tuistudio/bubblestudio/internal/list"
	"github.com/tuistudio/bubblestudio/internal/theme"
)

// mode identifies the active body view.
type mode int

const (
	modeList mode = iota
	modeForm
)

// Model is the root application model.
type Model struct {
	keys     keymap.KeyMap
	theme    theme.Theme
	help     help.Model
	list     list.Model
	form     form.Model
	mode     mode
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
		list:  list.New(demo.Components, "Components", 0, 0), // sized on first WindowSizeMsg
		form:  form.New(demo.FormFields()...),
	}
}

// Init satisfies tea.Model; no I/O commands on startup.
func (m Model) Init() tea.Cmd {
	return nil
}

// bodyHeight returns the number of lines available for the body region.
// It reserves 1 line each for the header and footer.
func (m Model) bodyHeight() int {
	h := m.height - 2
	if h < 0 {
		h = 0
	}
	return h
}

// Update handles incoming messages and updates state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, m.bodyHeight())
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		// Tab switches from list to form; in form mode Tab navigates fields.
		case key.Matches(msg, m.keys.Tab) && m.mode == modeList:
			m.mode = modeForm
			return m, m.form.Init()
		// Esc returns from form to list.
		case key.Matches(msg, m.keys.Back) && m.mode == modeForm:
			m.mode = modeList
			return m, nil
		}
	}

	// Forward to the active body component.
	var cmd tea.Cmd
	switch m.mode {
	case modeForm:
		m.form, cmd = m.form.Update(msg)
	default:
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

// View renders the header / body / footer layout.
func (m Model) View() string {
	header := m.theme.Header.Width(m.width).Render("bubblestudio")

	var body string
	switch {
	case m.showHelp:
		body = m.theme.Body.Width(m.width).Height(m.bodyHeight()).Render(m.help.View(m.keys))
	case m.mode == modeForm:
		body = m.theme.Body.Width(m.width).Height(m.bodyHeight()).Render(m.form.View())
	default:
		// List manages its own sizing; no outer style wrapper needed.
		body = m.list.View()
	}

	footer := m.theme.Footer.Width(m.width).Render("q quit • ? help • tab form demo • esc back")

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}
