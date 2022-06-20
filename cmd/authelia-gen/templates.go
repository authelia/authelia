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
		"stringsContains": stringsContains,
	}

	tmplCodeConfigurationSchemaKeys = template.Must(newTMPL("internal_configuration_schema_keys.go"))
	tmplGitHubIssueTemplateBug      = template.Must(newTMPL("github_issue_template_bug_report.yml"))
	tmplIssueTemplateFeature        = template.Must(newTMPL("github_issue_template_feature.yml"))
	tmplWebI18NIndex                = template.Must(newTMPL("web_i18n_index.ts"))
	tmplDotCommitLintRC             = template.Must(newTMPL("dot_commitlintrc.js"))
)

func newTMPL(name string) (tmpl *template.Template, err error) {
	return template.New(name).Funcs(funcMap).Parse(MustLoadTmplFS(name))
}

func MustLoadTmplFS(tmpl string) string {
	var (
		content []byte
		err     error
	)

	if content, err = templatesFS.ReadFile(fmt.Sprintf("templates/%s.tmpl", tmpl)); err != nil {
		panic(err)
	}

	return string(content)
}

func stringsContains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
