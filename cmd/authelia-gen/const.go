// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

const (
	dirCurrent = "./"
	dirLocales = "internal/server/locales"
	dirWeb     = "web"

	subPathCmd      = "cmd"
	subPathInternal = "internal"

	fileCICommitLintConfig = ".commitlintrc.js"
	fileWebI18NIndex       = "src/i18n/index.ts"
	fileWebPackage         = "package.json"

	fileDocsCommitMessageGuidelines = "docs/content/en/contributing/guidelines/commit-message.md"

	fileCodeConfigKeys  = "internal/configuration/schema/keys.go"
	fileServerGenerated = "internal/server/gen.go"
	fileScriptsGen      = "cmd/authelia-scripts/cmd/gen.go"

	dirDocs             = "docs"
	dirDocsContent      = "content"
	dirDocsData         = "data"
	dirDocsCLIReference = "en/reference/cli"

	fileDocsDataLanguages  = "languages.json"
	fileDocsDataMisc       = "misc.json"
	fileDocsDataConfigKeys = "configkeys.json"

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
	cmdUseDocsData               = "data"
	cmdUseDocsDataMisc           = "misc"
	cmdUseGitHub                 = "github"
	cmdUseGitHubIssueTemplates   = "issue-templates"
	cmdUseGitHubIssueTemplatesFR = "feature-request"
	cmdUseGitHubIssueTemplatesBR = "bug-report"
	cmdUseLocales                = "locales"
	cmdUseCommitLint             = "commit-lint"
	cmdUseCode                   = "code"
	cmdUseCodeScripts            = "scripts"
	cmdUseKeys                   = "keys"
	cmdUseServer                 = "server"
)

const (
	cmdFlagRoot                        = "dir.root"
	cmdFlagWeb                         = "dir.web"
	cmdFlagFileWebI18N                 = "file.web.i18n"
	cmdFlagFileWebPackage              = "file.web.package"
	cmdFlagDocs                        = "dir.docs"
	cmdFlagDirLocales                  = "dir.locales"
	cmdFlagDocsCLIReference            = "dir.docs.cli-reference"
	cmdFlagDocsContent                 = "dir.docs.content"
	cmdFlagDocsData                    = "dir.docs.data"
	cmdFlagDocsDataMisc                = "file.docs.data.misc"
	cmdFlagDocsDataKeys                = "file.docs.data.keys"
	cmdFlagDocsDataLanguages           = "file.docs.data.languages"
	cmdFlagFileConfigKeys              = "file.configuration-keys"
	cmdFlagFileScriptsGen              = "file.scripts.gen"
	cmdFlagFileServerGenerated         = "file.server.generated"
	cmdFlagFileConfigCommitLint        = "file.commit-lint-config"
	cmdFlagFileDocsCommitMsgGuidelines = "file.docs-commit-msg-guidelines"
	cmdFlagFeatureRequest              = "file.feature-request"
	cmdFlagBugReport                   = "file.bug-report"

	cmdFlagExclude           = "exclude"
	cmdFlagVersions          = "versions"
	cmdFlagCwd               = "cwd"
	cmdFlagPackageConfigKeys = "package.configuration.keys"
	cmdFlagPackageScriptsGen = "package.scripts.gen"
)

const (
	codeCSPProductionDefaultSrc  = "'self'"
	codeCSPDevelopmentDefaultSrc = "'self' 'unsafe-eval'"
	codeCSPNonce                 = "${NONCE}"
)

var (
	codeCSPValuesCommon = []CSPValue{
		{Name: "default-src", Value: ""},
		{Name: "frame-src", Value: "'none'"},
		{Name: "object-src", Value: "'none'"},
		{Name: "style-src", Value: "'self' 'nonce-%s'"},
		{Name: "frame-ancestors", Value: "'none'"},
		{Name: "base-uri", Value: "'self'"},
	}

	codeCSPValuesProduction = []CSPValue{}

	codeCSPValuesDevelopment = []CSPValue{}
)
