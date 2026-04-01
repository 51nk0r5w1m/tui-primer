// Package bubblestudio is a runtime library for TUI Studio exports.
// It reads a .tui JSON design file and returns a ready-to-run tea.Model.
package bubblestudio

import (
	"encoding/json"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── JSON schema (mirrors TypeScript ComponentNode) ───────────────────────────

// TUINode is the deserialized form of a ComponentNode from a .tui JSON file.
type TUINode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Props    map[string]interface{} `json:"props"`
	Layout   TUILayout              `json:"layout"`
	Style    TUIStyle               `json:"style"`
	Events   TUIEvents              `json:"events"`
	Children []*TUINode             `json:"children"`
	Hidden   bool                   `json:"hidden"`
}

// TUILayout mirrors LayoutProps.
type TUILayout struct {
	Type      string      `json:"type"`
	Direction string      `json:"direction"`
	Justify   string      `json:"justify"`
	Align     string      `json:"align"`
	Gap       int         `json:"gap"`
	Wrap      bool        `json:"wrap"`
	Columns   int         `json:"columns"`
	Rows      int         `json:"rows"`
	X         float64     `json:"x"`
	Y         float64     `json:"y"`
	Padding   interface{} `json:"padding"`
	Margin    interface{} `json:"margin"`
}

// TUIStyle mirrors StyleProps.
type TUIStyle struct {
	Border          bool   `json:"border"`
	BorderStyle     string `json:"borderStyle"`
	BorderColor     string `json:"borderColor"`
	Color           string `json:"color"`
	BackgroundColor string `json:"backgroundColor"`
	Bold            bool   `json:"bold"`
	Italic          bool   `json:"italic"`
	Underline       bool   `json:"underline"`
}

// TUIEvents mirrors EventHandlers.
type TUIEvents struct {
	OnFocus    string `json:"onFocus"`
	OnBlur     string `json:"onBlur"`
	OnClick    string `json:"onClick"`
	OnSubmit   string `json:"onSubmit"`
	OnChange   string `json:"onChange"`
	OnKeyPress string `json:"onKeyPress"`
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// Handlers holds user-defined callbacks keyed by component name.
// Each map key is the component's Name field from the design file.
type Handlers struct {
	// OnClick is called when a Button is activated (Enter key).
	OnClick map[string]func()
	// OnChange is called when a TextInput value changes; receives the new value.
	OnChange map[string]func(string)
	// OnSubmit is called when a TextInput is submitted (Enter key); receives the value.
	OnSubmit map[string]func(string)
	// OnToggle is called when a Checkbox, Radio, or Toggle changes; receives the new bool.
	OnToggle map[string]func(bool)
	// OnSelect is called when a Select or List item is chosen; receives the selected string.
	OnSelect map[string]func(string)
	// OnTab is called when a Tabs component changes; receives the new tab label.
	OnTab map[string]func(string)
}

// ── Public API ────────────────────────────────────────────────────────────────

// Load reads a .tui JSON design file from path and returns a tea.Model
// that is ready to pass to tea.NewProgram.  Provide Handlers{} for no-op
// callbacks; fill in the maps to wire up your application logic.
func Load(path string, h Handlers) (tea.Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("bubblestudio: read %s: %w", path, err)
	}
	var root TUINode
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("bubblestudio: parse %s: %w", path, err)
	}
	return newScreen(&root, h), nil
}

// ── Screen (root tea.Model) ───────────────────────────────────────────────────

// Screen is the root BubbleTea model produced by Load.
type Screen struct {
	root       *TUINode
	states     map[string]*compState // component ID → state
	focusOrder []string              // IDs of focusable components, in tree order
	focusIdx   int
	handlers   Handlers
	width      int
	height     int
	ready      bool
}

func newScreen(root *TUINode, h Handlers) *Screen {
	s := &Screen{
		root:     root,
		states:   make(map[string]*compState),
		handlers: h,
	}
	// Walk tree: initialise component state and build focus order.
	s.walkInit(root)
	// Focus the first focusable component.
	if len(s.focusOrder) > 0 {
		if st, ok := s.states[s.focusOrder[0]]; ok {
			st.setFocus(true)
		}
	}
	return s
}

func (s *Screen) walkInit(node *TUINode) {
	if node == nil || node.Hidden {
		return
	}
	st := initComponent(node)
	if st != nil {
		s.states[node.ID] = st
		if isFocusable(node.Type) {
			s.focusOrder = append(s.focusOrder, node.ID)
		}
	}
	for _, child := range node.Children {
		s.walkInit(child)
	}
}

// ── tea.Model implementation ─────────────────────────────────────────────────

func (s *Screen) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, st := range s.states {
		if c := st.initCmd(); c != nil {
			cmds = append(cmds, c)
		}
	}
	return tea.Batch(cmds...)
}

func (s *Screen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.ready = true
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return s, tea.Quit
		case "tab":
			s.shiftFocus(1)
			return s, nil
		case "shift+tab":
			s.shiftFocus(-1)
			return s, nil
		}
		// Route to focused component.
		if len(s.focusOrder) > 0 {
			id := s.focusOrder[s.focusIdx]
			node := s.findNode(s.root, id)
			if node != nil {
				if st, ok := s.states[id]; ok {
					cmd := s.updateComponent(st, node, msg)
					return s, cmd
				}
			}
		}
	}
	return s, nil
}

func (s *Screen) View() string {
	if !s.ready {
		return "Loading…"
	}
	return renderNode(s, s.root)
}

// ── Focus management ─────────────────────────────────────────────────────────

func (s *Screen) shiftFocus(delta int) {
	if len(s.focusOrder) == 0 {
		return
	}
	// Blur current.
	if st, ok := s.states[s.focusOrder[s.focusIdx]]; ok {
		st.setFocus(false)
	}
	n := len(s.focusOrder)
	s.focusIdx = ((s.focusIdx + delta) % n + n) % n
	// Focus next.
	if st, ok := s.states[s.focusOrder[s.focusIdx]]; ok {
		st.setFocus(true)
	}
}

func (s *Screen) findNode(node *TUINode, id string) *TUINode {
	if node == nil {
		return nil
	}
	if node.ID == id {
		return node
	}
	for _, child := range node.Children {
		if found := s.findNode(child, id); found != nil {
			return found
		}
	}
	return nil
}

// ── Rendering ────────────────────────────────────────────────────────────────

// renderNode recursively renders a node tree to a Lip Gloss string.
func renderNode(s *Screen, node *TUINode) string {
	if node == nil || node.Hidden {
		return ""
	}

	// If this node has a component state, render via the component.
	if st, ok := s.states[node.ID]; ok {
		focused := len(s.focusOrder) > 0 && s.focusOrder[s.focusIdx] == node.ID
		content := viewComponent(st, node, focused)
		return applyContainerStyle(node, content)
	}

	// Display-only leaf nodes (Text, Menu, Breadcrumb, Tree, Spacer, …).
	if content, isLeaf := renderLeaf(node); isLeaf {
		return applyContainerStyle(node, content)
	}

	// Container node: render children and compose with layout.
	var childViews []string
	for _, child := range node.Children {
		if v := renderNode(s, child); v != "" {
			childViews = append(childViews, v)
		}
	}

	content := composeChildren(node, childViews)
	return applyContainerStyle(node, content)
}

// composeChildren joins child views according to the node's layout.
func composeChildren(node *TUINode, views []string) string {
	if len(views) == 0 {
		return ""
	}

	switch node.Layout.Type {
	case "absolute":
		// TODO: absolute positioning is approximated as vertical stacking.
		// Precise x/y positioning requires knowing terminal dimensions per node.
		return lipgloss.JoinVertical(lipgloss.Left, views...)

	case "grid":
		cols := node.Layout.Columns
		if cols <= 0 {
			cols = 2
		}
		var rows []string
		for i := 0; i < len(views); i += cols {
			end := i + cols
			if end > len(views) {
				end = len(views)
			}
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, views[i:end]...))
		}
		return lipgloss.JoinVertical(lipgloss.Left, rows...)

	default: // flexbox and none
		if node.Layout.Direction == "row" {
			return lipgloss.JoinHorizontal(lipgloss.Top, views...)
		}
		return lipgloss.JoinVertical(lipgloss.Left, views...)
	}
}

// applyContainerStyle wraps content in a Lip Gloss style derived from node.Style.
func applyContainerStyle(node *TUINode, content string) string {
	st := lipgloss.NewStyle()

	if node.Style.Border {
		bs := toBorderStyle(node.Style.BorderStyle)
		st = st.Border(bs)
		if node.Style.BorderColor != "" {
			st = st.BorderForeground(lipgloss.Color(node.Style.BorderColor))
		}
	}
	if node.Style.Color != "" {
		st = st.Foreground(lipgloss.Color(node.Style.Color))
	}
	if node.Style.BackgroundColor != "" {
		st = st.Background(lipgloss.Color(node.Style.BackgroundColor))
	}

	pad := flatPadding(node.Layout.Padding)
	if pad > 0 {
		st = st.Padding(pad)
	}

	return st.Render(content)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func toBorderStyle(s string) lipgloss.Border {
	switch s {
	case "double":
		return lipgloss.DoubleBorder()
	case "rounded":
		return lipgloss.RoundedBorder()
	case "bold":
		return lipgloss.ThickBorder()
	case "hidden":
		return lipgloss.HiddenBorder()
	default:
		return lipgloss.NormalBorder()
	}
}

func flatPadding(v interface{}) int {
	switch p := v.(type) {
	case float64:
		return int(p)
	case int:
		return p
	}
	return 0
}

func propString(props map[string]interface{}, key, fallback string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return fallback
}

func propFloat(props map[string]interface{}, key string, fallback float64) float64 {
	if v, ok := props[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return fallback
}

func propBool(props map[string]interface{}, key string, fallback bool) bool {
	if v, ok := props[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return fallback
}

func propStrings(props map[string]interface{}, key string) []string {
	v, ok := props[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		switch s := item.(type) {
		case string:
			out = append(out, s)
		case map[string]interface{}:
			if label, ok := s["label"].(string); ok {
				out = append(out, label)
			}
		}
	}
	return out
}
