package templates

import (
	"text/template"
)

// HTMLPlainTextTemplate is the template type which contains both the html and txt versions of a template.
type HTMLPlainTextTemplate struct {
	html *template.Template
	txt  *template.Template
}

// Get returns the appropriate template given the format.
func (f HTMLPlainTextTemplate) Get(format Format) (t *template.Template) {
	switch format {
	case HTMLFormat:
		return f.html
	case PlainTextFormat:
		return f.txt
	default:
		return f.html
	}
}
