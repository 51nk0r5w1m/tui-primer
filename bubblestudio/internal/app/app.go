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

// appTitle and footerHints are the fixed chrome strings rendered in the
// header and footer.  They are used in both bodyHeight() for size measurement
// and View() for rendering.
const (
	appTitle    = "bubblestudio"
	footerHints = "q quit • ? help • tab form demo • esc back"
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
// It measures the actual rendered height of the header and footer using
// lipgloss.Height instead of relying on a fixed 2-line heuristic, so the
// calculation remains correct even if the theme changes vertical padding.
func (m Model) bodyHeight() int {
	headerH := lipgloss.Height(m.theme.Header.Width(m.width).Render(appTitle))
	footerH := lipgloss.Height(m.theme.Footer.Width(m.width).Render(footerHints))
	h := m.height - headerH - footerH
	if h < 0 {
		h = 0
	}
	return h
}

// listInnerSize returns the width and height available to the list after
// subtracting the theme.Body padding so the list fits precisely inside the
// body style box.
func (m Model) listInnerSize() (w, h int) {
	hp := m.theme.Body.GetHorizontalPadding()
	vp := m.theme.Body.GetVerticalPadding()
	w = max(0, m.width-hp)
	h = max(0, m.bodyHeight()-vp)
	return w, h
}

// route returns the next mode given a key event, decoupling mode transition
// logic from the key-dispatch section of Update.  Returning the current mode
// signals "no transition".
func (m Model) route(msg tea.KeyMsg) (mode, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Tab) && m.mode == modeList:
		return modeForm, m.form.Init()
	case key.Matches(msg, m.keys.Back) && m.mode == modeForm:
		return modeList, nil
	}
	return m.mode, nil
}

// Update handles incoming messages and updates state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		w, h := m.listInnerSize()
		m.list.SetSize(w, h)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		}

		// Route mode transitions before forwarding to body components.
		if next, cmd := m.route(msg); next != m.mode {
			m.mode = next
			return m, cmd
		}

		// Don't forward key events to body components while the help overlay is
		// visible — doing so would silently mutate hidden list/form state.
		if m.showHelp {
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
// theme.Body is applied uniformly to all three body states so styling is
// consistent whether the list, form, or help overlay is shown.
func (m Model) View() string {
	header := m.theme.Header.Width(m.width).Render(appTitle)

	bodyStyle := m.theme.Body.Width(m.width).Height(m.bodyHeight())
	var body string
	switch {
	case m.showHelp:
		body = bodyStyle.Render(m.help.View(m.keys))
	case m.mode == modeForm:
		body = bodyStyle.Render(m.form.View())
	default:
		// The list is pre-sized to fit within the body padding (see listInnerSize),
		// so theme.Body wraps it without clipping or overflowing.
		body = bodyStyle.Render(m.list.View())
	}

	footer := m.theme.Footer.Width(m.width).Render(footerHints)

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}
