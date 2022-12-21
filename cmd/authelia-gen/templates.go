package main

import (
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/authelia/authelia/v4/internal/templates"
)

//go:embed templates/*
var templatesFS embed.FS

var (
	tmplCodeConfigurationSchemaKeys = template.Must(newTMPL("internal_configuration_schema_keys.go"))
	tmplGitHubIssueTemplateBug      = template.Must(newTMPL("github_issue_template_bug_report.yml"))
	tmplIssueTemplateFeature        = template.Must(newTMPL("github_issue_template_feature.yml"))
	tmplWebI18NIndex                = template.Must(newTMPL("web_i18n_index.ts"))
	tmplDotCommitLintRC             = template.Must(newTMPL("dot_commitlintrc.js"))
	tmplDocsCommitMessageGuidelines = template.Must(newTMPL("docs-contributing-development-commitmsg.md"))
	tmplScriptsGen                  = template.Must(newTMPL("cmd-authelia-scripts-gen.go"))
	tmplServer                      = template.Must(newTMPL("server_gen.go"))
)

func newTMPL(name string) (tmpl *template.Template, err error) {
	return template.New(name).
		Funcs(template.FuncMap{
			"stringsContains": strings.Contains,
			"join":            strings.Join,
			"joinX":           templates.StringJoinXFunc,
		}).
		Parse(mustLoadTmplFS(name))
}

func mustLoadTmplFS(tmpl string) string {
	var (
		content []byte
		err     error
	)

	if content, err = templatesFS.ReadFile(fmt.Sprintf("templates/%s.tmpl", tmpl)); err != nil {
		panic(err)
	}

	return string(content)
}
