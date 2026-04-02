// Package theme defines lipgloss styles for the tui-factory shell.
package theme

import "github.com/charmbracelet/lipgloss"

// Theme holds the styles for the three layout regions.
type Theme struct {
	Header lipgloss.Style
	Footer lipgloss.Style
	Body   lipgloss.Style
}

// Default returns a minimal Theme.
func Default() Theme {
	return Theme{
		Header: lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Footer: lipgloss.NewStyle().Faint(true).Padding(0, 1),
		Body:   lipgloss.NewStyle().Padding(1, 2),
	}
}
