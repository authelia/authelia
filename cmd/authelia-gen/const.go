package main

const (
	dirCurrent = "./"
	dirLocales = "internal/server/locales"

	subPathCmd      = "cmd"
	subPathInternal = "internal"

	fileCICommitLintConfig = "web/.commitlintrc.js"
	fileWebI18NIndex       = "web/src/i18n/index.ts"

	fileDocsCommitMessageGuidelines = "docs/content/en/contributing/development/guidelines-commit-message.md"

	fileCodeConfigKeys = "internal/configuration/schema/keys.go"
	fileScriptsGen     = "cmd/authelia-scripts/cmd/gen.go"

	dirDocsContent      = "docs/content"
	dirDocsCLIReference = dirDocsContent + "/en/reference/cli"

	fileDocsDataLanguages = "docs/data/languages.json"

	fileGitHubIssueTemplateFR = ".github/ISSUE_TEMPLATE/feature-request.yml"
	fileGitHubIssueTemplateBR = ".github/ISSUE_TEMPLATE/bug-report.yml"
)

const (
	dateFmtRFC2822 = "Mon, _2 Jan 2006 15:04:05 -0700"
	dateFmtYAML    = "2006-01-02T15:04:05-07:00"
)

const (
	delimiterLineFrontMatter = "---"

	localeDefault          = "en"
	localeNamespaceDefault = "portal"
)

const (
	pkgConfigSchema = "schema"
	pkgScriptsGen   = "cmd"
)

const (
	cmdUseRoot                   = "authelia-gen"
	cmdUseCompletion             = "completion"
	cmdUseDocs                   = "docs"
	cmdUseDocsDate               = "date"
	cmdUseDocsCLI                = "cli"
	cmdUseGitHub                 = "github"
	cmdUseGitHubIssueTemplates   = "issue-templates"
	cmdUseGitHubIssueTemplatesFR = "feature-request"
	cmdUseGitHubIssueTemplatesBR = "bug-report"
	cmdUseLocales                = "locales"
	cmdUseCommitLint             = "commit-lint"
	cmdUseCode                   = "code"
	cmdUseCodeScripts            = "scripts"
	cmdUseCodeKeys               = "keys"
)

const (
	cmdFlagRoot                        = "dir.root"
	cmdFlagExclude                     = "exclude"
	cmdFlagVersions                    = "versions"
	cmdFlagDirLocales                  = "dir.locales"
	cmdFlagDocsCLIReference            = "dir.docs.cli-reference"
	cmdFlagDocsContent                 = "dir.docs.content"
	cmdFlagDocsDataLanguages           = "file.docs.data.languages"
	cmdFlagCwd                         = "cwd"
	cmdFlagFileConfigKeys              = "file.configuration-keys"
	cmdFlagFileScriptsGen              = "file.scripts.gen"
	cmdFlagFileConfigCommitLint        = "file.commit-lint-config"
	cmdFlagFileDocsCommitMsgGuidelines = "file.docs-commit-msg-guidelines"
	cmdFlagFileWebI18N                 = "file.web-i18n"
	cmdFlagFeatureRequest              = "file.feature-request"
	cmdFlagBugReport                   = "file.bug-report"
	cmdFlagPackageConfigKeys           = "package.configuration.keys"
	cmdFlagPackageScriptsGen           = "package.scripts.gen"
)
