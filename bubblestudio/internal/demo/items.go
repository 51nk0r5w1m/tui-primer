// Package demo provides synthetic sample data for bubblestudio primitives.
// All data is fictional and is used only to exercise the UI layer.
package demo

import "github.com/tuistudio/bubblestudio/internal/list"

// Components is the sample dataset for the list demo view.
var Components = []list.Item{
	list.NewItem("Table", "Scrollable data grid with column headers"),
	list.NewItem("Form", "Input fields with validation"),
	list.NewItem("Modal", "Overlay dialog for confirmations"),
	list.NewItem("Progress", "Progress bar or spinner"),
	list.NewItem("Logs", "Streaming log output panel"),
	list.NewItem("Wizard", "Multi-step guided workflow"),
	list.NewItem("Dashboard", "Overview with key metrics"),
	list.NewItem("Status", "Status bar with contextual info"),
}
