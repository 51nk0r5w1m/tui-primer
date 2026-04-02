package demo

import (
	"strings"

	"github.com/tuistudio/bubblestudio/internal/form"
)

// FormFields returns the sample set of fields for the form demo view.
// All labels and defaults are generic and domain-agnostic.
func FormFields() []form.Field {
	return []form.Field{
		form.NewField("Title", "enter a title", func(v string) string {
			if strings.TrimSpace(v) == "" {
				return "required"
			}
			return ""
		}),
		form.NewField("Description", "enter a description (optional)", nil),
		form.NewField("Version", "e.g. 1.0.0", nil),
	}
}
