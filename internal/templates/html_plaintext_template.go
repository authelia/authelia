package templates

import (
	th "html/template"
	tt "text/template"
)

// EmailTemplate is the template type which contains both the html and txt versions of a template.
type EmailTemplate struct {
	HTML *th.Template
	Text *tt.Template
}

// Get returns the appropriate template given the format.
func (t *EmailTemplate) Get(format Format) (tmpl Template) {
	switch format {
	case HTMLFormat:
		return t.HTML
	case PlainTextFormat:
		return t.Text
	default:
		return t.HTML
	}
}
