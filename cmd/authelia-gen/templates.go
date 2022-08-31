package main

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates/*
var templatesFS embed.FS

var (
	funcMap = template.FuncMap{
		"stringsContains": strings.Contains,
		"join":            strings.Join,
		"joinX":           fmJoinX,
	}

	tmplCodeConfigurationSchemaKeys = template.Must(newTMPL("internal_configuration_schema_keys.go"))
	tmplGitHubIssueTemplateBug      = template.Must(newTMPL("github_issue_template_bug_report.yml"))
	tmplIssueTemplateFeature        = template.Must(newTMPL("github_issue_template_feature.yml"))
	tmplWebI18NIndex                = template.Must(newTMPL("web_i18n_index.ts"))
	tmplDotCommitLintRC             = template.Must(newTMPL("dot_commitlintrc.js"))
	tmplDocsCommitMessageGuidelines = template.Must(newTMPL("docs-contributing-development-commitmsg.md"))
	tmplScriptsGen                  = template.Must(newTMPL("cmd-authelia-scripts-gen.go"))
)

func fmJoinX(elems []string, sep string, n int, p string) string {
	buf := strings.Builder{}

	c := 0
	e := len(elems) - 1

	for i := 0; i <= e; i++ {
		if c+len(elems[i])+1 > n {
			c = 0

			buf.WriteString(p)
		}

		c += len(elems[i]) + 1

		buf.WriteString(elems[i])

		if i < e {
			buf.WriteString(sep)
		}
	}

	return buf.String()
}

func newTMPL(name string) (tmpl *template.Template, err error) {
	return template.New(name).Funcs(funcMap).Parse(mustLoadTmplFS(name))
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
