package main

import (
	"embed"
	"fmt"
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
	tmplDotCommitLintRC             = template.Must(newTMPL("dot_commitlintrc.cjs"))
	tmplDocsCommitMessageGuidelines = template.Must(newTMPL("docs-contributing-development-commitmsg.md"))
	tmplScriptsGen                  = template.Must(newTMPL("cmd-authelia-scripts-gen.go"))
	tmplServer                      = template.Must(newTMPL("server_gen.go"))
)

func newTMPL(name string) (tmpl *template.Template, err error) {
	funcs := templates.FuncMap()

	funcs["joinX"] = templates.FuncStringJoinX

	return template.New(name).Funcs(funcs).Parse(mustLoadTmplFS(name))
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
