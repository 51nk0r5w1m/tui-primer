package bubblestudio

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Focus check ───────────────────────────────────────────────────────────────

// isFocusable returns true for interactive component types that join the focus ring.
func isFocusable(t string) bool {
	switch t {
	case "TextInput", "Button", "Checkbox", "Radio", "Toggle", "Select", "List", "Tabs":
		return true
	}
	return false
}

// ── compState ─────────────────────────────────────────────────────────────────

// compState holds the runtime state for a single TUI component.
// Only the field relevant to the component kind is populated.
type compState struct {
	kind      string
	focused   bool
	// Bubbles models
	textInput textinput.Model
	list      list.Model
	table     table.Model
	spin      spinner.Model
	prog      progress.Model
	// Simple state
	checked bool    // Checkbox / Radio / Toggle
	tabIdx  int     // Tabs: currently active tab index
	value   float64 // ProgressBar value (0‥100)
}

// initComponent creates a compState for nodes that need runtime state.
// Returns nil for pure layout containers (Screen, Box, Grid, Spacer, Modal, Text, …).
func initComponent(node *TUINode) *compState {
	switch node.Type {
	case "TextInput":
		ti := textinput.New()
		ti.Placeholder = propString(node.Props, "placeholder", "")
		ti.SetValue(propString(node.Props, "value", ""))
		ti.CharLimit = 256
		return &compState{kind: node.Type, textInput: ti}

	case "Button":
		return &compState{kind: node.Type}

	case "Checkbox":
		return &compState{kind: node.Type, checked: propBool(node.Props, "checked", false)}

	case "Radio":
		return &compState{kind: node.Type, checked: propBool(node.Props, "checked", false)}

	case "Toggle":
		return &compState{kind: node.Type, checked: propBool(node.Props, "value", false)}

	case "Select":
		items := listItems(propStrings(node.Props, "options"))
		l := list.New(items, list.NewDefaultDelegate(), 20, 6)
		l.SetShowHelp(false)
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		return &compState{kind: node.Type, list: l}

	case "List":
		items := listItems(propStrings(node.Props, "items"))
		l := list.New(items, list.NewDefaultDelegate(), 20, 10)
		l.SetShowHelp(false)
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		return &compState{kind: node.Type, list: l}

	case "Table":
		cols, rows := tableData(node.Props)
		t := table.New(
			table.WithColumns(cols),
			table.WithRows(rows),
			table.WithFocused(false),
			table.WithHeight(8),
		)
		return &compState{kind: node.Type, table: t}

	case "Spinner":
		sp := spinner.New()
		sp.Spinner = spinner.Dot
		return &compState{kind: node.Type, spin: sp}

	case "ProgressBar":
		v := propFloat(node.Props, "value", 0)
		max := propFloat(node.Props, "max", 100)
		pct := 0.0
		if max > 0 {
			pct = v / max
		}
		pr := progress.New(progress.WithDefaultGradient())
		return &compState{kind: node.Type, prog: pr, value: pct}

	case "Tabs":
		return &compState{kind: node.Type, tabIdx: 0}

	default:
		// Text, Box, Grid, Screen, Modal, Spacer, Menu, Breadcrumb, Tree —
		// these are pure layout / display; no interactive state needed.
		return nil
	}
}

// initCmd returns a tea.Cmd to fire on startup (e.g. spinner tick).
func (st *compState) initCmd() tea.Cmd {
	switch st.kind {
	case "Spinner":
		return st.spin.Tick
	case "TextInput":
		if st.focused {
			return textinput.Blink
		}
	}
	return nil
}

// setFocus focuses or blurs the component.
func (st *compState) setFocus(on bool) {
	st.focused = on
	switch st.kind {
	case "TextInput":
		if on {
			st.textInput.Focus()
		} else {
			st.textInput.Blur()
		}
	case "Table":
		st.table.SetStyles(tableStyles(on))
	}
}

// ── Update routing ────────────────────────────────────────────────────────────

// updateComponent delegates a key message to the focused component.
// It calls any registered handlers and returns a tea.Cmd if needed.
func (s *Screen) updateComponent(st *compState, node *TUINode, msg tea.Msg) tea.Cmd {
	switch st.kind {
	case "TextInput":
		var cmd tea.Cmd
		prevVal := st.textInput.Value()
		st.textInput, cmd = st.textInput.Update(msg)
		newVal := st.textInput.Value()
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			if fn := s.handlers.OnSubmit[node.Name]; fn != nil {
				fn(newVal)
			}
		} else if newVal != prevVal {
			if fn := s.handlers.OnChange[node.Name]; fn != nil {
				fn(newVal)
			}
		}
		return cmd

	case "Button":
		if key, ok := msg.(tea.KeyMsg); ok && (key.String() == "enter" || key.String() == " ") {
			if fn := s.handlers.OnClick[node.Name]; fn != nil {
				fn()
			}
		}

	case "Checkbox", "Radio", "Toggle":
		if key, ok := msg.(tea.KeyMsg); ok && (key.String() == "enter" || key.String() == " ") {
			st.checked = !st.checked
			if fn := s.handlers.OnToggle[node.Name]; fn != nil {
				fn(st.checked)
			}
		}

	case "Select", "List":
		var cmd tea.Cmd
		st.list, cmd = st.list.Update(msg)
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			if item, ok := st.list.SelectedItem().(simpleItem); ok {
				if fn := s.handlers.OnSelect[node.Name]; fn != nil {
					fn(string(item))
				}
			}
		}
		return cmd

	case "Table":
		var cmd tea.Cmd
		st.table, cmd = st.table.Update(msg)
		return cmd

	case "Tabs":
		tabs := propStrings(node.Props, "tabs")
		if len(tabs) == 0 {
			break
		}
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "right", "l":
				if st.tabIdx < len(tabs)-1 {
					st.tabIdx++
					if fn := s.handlers.OnTab[node.Name]; fn != nil {
						fn(tabs[st.tabIdx])
					}
				}
			case "left", "h":
				if st.tabIdx > 0 {
					st.tabIdx--
					if fn := s.handlers.OnTab[node.Name]; fn != nil {
						fn(tabs[st.tabIdx])
					}
				}
			}
		}
	}
	return nil
}

// ── View ──────────────────────────────────────────────────────────────────────

// viewComponent renders a single component to a string.
func viewComponent(st *compState, node *TUINode, focused bool) string {
	focusColor := lipgloss.Color("205") // pink/magenta focus indicator

	switch st.kind {
	case "TextInput":
		return st.textInput.View()

	case "Button":
		label := propString(node.Props, "label", "Button")
		s := lipgloss.NewStyle().Padding(0, 2)
		if focused {
			s = s.Foreground(lipgloss.Color("0")).Background(focusColor)
		} else {
			s = s.Border(lipgloss.NormalBorder())
		}
		return s.Render(label)

	case "Checkbox":
		label := propString(node.Props, "label", "")
		icon := "[ ]"
		if st.checked {
			icon = "[✓]"
		}
		s := lipgloss.NewStyle()
		if focused {
			s = s.Foreground(focusColor)
		}
		return s.Render(fmt.Sprintf("%s %s", icon, label))

	case "Radio":
		label := propString(node.Props, "label", "")
		icon := "(○)"
		if st.checked {
			icon = "(◉)"
		}
		s := lipgloss.NewStyle()
		if focused {
			s = s.Foreground(focusColor)
		}
		return s.Render(fmt.Sprintf("%s %s", icon, label))

	case "Toggle":
		label := propString(node.Props, "label", "")
		indicator := "[OFF]"
		if st.checked {
			indicator = "[ ON]"
		}
		s := lipgloss.NewStyle()
		if focused {
			s = s.Foreground(focusColor)
		}
		return s.Render(fmt.Sprintf("%s %s", indicator, label))

	case "Select", "List":
		return st.list.View()

	case "Table":
		return st.table.View()

	case "Spinner":
		return st.spin.View()

	case "ProgressBar":
		width := 40
		return st.prog.ViewAs(st.value) + fmt.Sprintf(" %.0f%%", st.value*100)
		_ = width

	case "Tabs":
		tabs := propStrings(node.Props, "tabs")
		if len(tabs) == 0 {
			return ""
		}
		var parts []string
		for i, tab := range tabs {
			s := lipgloss.NewStyle().Padding(0, 1)
			if i == st.tabIdx {
				s = s.Underline(true)
				if focused {
					s = s.Foreground(focusColor)
				}
			}
			parts = append(parts, s.Render(tab))
		}
		return lipgloss.JoinHorizontal(lipgloss.Bottom, parts...)
	}

	return ""
}

// ── Non-interactive leaf views ────────────────────────────────────────────────
// These are called from renderNode for nodes that have no compState.

// viewText renders a Text node.
func viewText(node *TUINode) string {
	content := propString(node.Props, "content", "")
	s := lipgloss.NewStyle()
	if node.Style.Color != "" {
		s = s.Foreground(lipgloss.Color(node.Style.Color))
	}
	if node.Style.Bold {
		s = s.Bold(true)
	}
	if node.Style.Italic {
		s = s.Italic(true)
	}
	if node.Style.Underline {
		s = s.Underline(true)
	}
	return s.Render(content)
}

// viewMenu renders a Menu node.
func viewMenu(node *TUINode) string {
	items := propStrings(node.Props, "items")
	if len(items) == 0 {
		return ""
	}
	s := lipgloss.NewStyle()
	rendered := make([]string, len(items))
	for i, item := range items {
		rendered[i] = s.Padding(0, 1).Render(item)
	}
	if node.Layout.Direction == "row" {
		return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

// viewBreadcrumb renders a Breadcrumb node.
func viewBreadcrumb(node *TUINode) string {
	items := propStrings(node.Props, "items")
	sep := propString(node.Props, "separator", "/")
	return strings.Join(items, " "+sep+" ")
}

// viewTree renders a Tree node with ASCII art prefixes.
func viewTree(node *TUINode) string {
	v, ok := node.Props["items"]
	if !ok {
		return ""
	}
	arr, ok := v.([]interface{})
	if !ok {
		return ""
	}
	var lines []string
	var walk func(items []interface{}, depth int)
	walk = func(items []interface{}, depth int) {
		for _, item := range items {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			label, _ := m["label"].(string)
			prefix := strings.Repeat("  ", depth)
			if depth > 0 {
				prefix += "├─ "
			}
			lines = append(lines, prefix+label)
			if children, ok := m["children"].([]interface{}); ok {
				walk(children, depth+1)
			}
		}
	}
	walk(arr, 0)
	return strings.Join(lines, "\n")
}

// Override renderNode to handle display-only leaf types.
func init() {
	// Monkey-patch via a wrapper would need global state; instead the rendering
	// logic below is merged into renderNode via the helper renderLeaf.
	_ = viewText  // used by renderLeaf
	_ = viewMenu
	_ = viewBreadcrumb
	_ = viewTree
}

// renderLeaf renders a non-container, non-interactive leaf node.
// Called by renderNode when there is no compState for the node.
func renderLeaf(node *TUINode) (string, bool) {
	switch node.Type {
	case "Text":
		return viewText(node), true
	case "Spacer":
		return "", true
	case "Menu":
		return viewMenu(node), true
	case "Breadcrumb":
		return viewBreadcrumb(node), true
	case "Tree":
		return viewTree(node), true
	}
	// TODO: no equivalent for this component type; placeholder shown.
	return fmt.Sprintf("[%s]", node.Type), false
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// simpleItem implements list.Item for plain string items.
type simpleItem string

func (s simpleItem) FilterValue() string { return string(s) }
func (s simpleItem) Title() string       { return string(s) }
func (s simpleItem) Description() string { return "" }

func listItems(ss []string) []list.Item {
	out := make([]list.Item, len(ss))
	for i, s := range ss {
		out[i] = simpleItem(s)
	}
	return out
}

func tableData(props map[string]interface{}) ([]table.Column, []table.Row) {
	colNames := propStrings(props, "columns")
	if len(colNames) == 0 {
		colNames = []string{"Column 1", "Column 2"}
	}
	cols := make([]table.Column, len(colNames))
	for i, c := range colNames {
		cols[i] = table.Column{Title: c, Width: 14}
	}

	var rows []table.Row
	if v, ok := props["rows"]; ok {
		if arr, ok := v.([]interface{}); ok {
			for _, rowRaw := range arr {
				if rowArr, ok := rowRaw.([]interface{}); ok {
					row := make(table.Row, len(cols))
					for i, cell := range rowArr {
						if i < len(cols) {
							row[i] = fmt.Sprintf("%v", cell)
						}
					}
					rows = append(rows, row)
				}
			}
		}
	}
	return cols, rows
}

func tableStyles(focused bool) table.Styles {
	s := table.DefaultStyles()
	if focused {
		s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
	} else {
		s.Selected = lipgloss.NewStyle()
	}
	return s
}
