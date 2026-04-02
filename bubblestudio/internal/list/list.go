// Package list provides a reusable selectable list component.
// Navigation: ↑/↓ or k/j to move, enter to select.
package list

import (
	bubblelist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Item is a single entry shown in the list.
type Item struct {
	title string
	desc  string
}

// NewItem constructs an Item with the given title and description.
func NewItem(title, desc string) Item { return Item{title: title, desc: desc} }

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.desc }
func (i Item) FilterValue() string { return i.title }

// Model wraps bubbles/list for use as a reusable body component.
type Model struct {
	list bubblelist.Model
}

// New returns a Model loaded with items and sized to w×h.
// Pass title as the list heading; pass items using NewItem or any type
// that satisfies bubbles/list.Item.
func New(items []Item, title string, w, h int) Model {
	bi := make([]bubblelist.Item, len(items))
	for i, it := range items {
		bi[i] = it
	}
	l := bubblelist.New(bi, bubblelist.NewDefaultDelegate(), w, h)
	l.Title = title
	l.SetShowHelp(false) // help is managed by the parent app
	return Model{list: l}
}

// SetSize resizes the list to fit the available area.
func (m *Model) SetSize(w, h int) {
	m.list.SetSize(w, h)
}

// Selected returns the currently highlighted item, or nil if the list is empty.
func (m Model) Selected() bubblelist.Item {
	return m.list.SelectedItem()
}

// Update forwards messages to the inner list model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the list.
func (m Model) View() string {
	return m.list.View()
}
