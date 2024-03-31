package main

const (
	dirCurrent = "./"
	dirLocales = "internal/server/locales"
	dirWeb     = "web"

	subPathCmd      = "cmd"
	subPathInternal = "internal"

	fileCICommitLintConfig = ".commitlintrc.cjs"
	fileWebI18NIndex       = "src/i18n/index.ts"
	fileWebPackage         = "package.json"

	fileDocsCommitMessageGuidelines = "docs/content/contributing/guidelines/commit-message.md"

	fileCodeConfigKeys  = "internal/configuration/schema/keys.go"
	fileServerGenerated = "internal/server/gen.go"
	fileScriptsGen      = "cmd/authelia-scripts/cmd/gen.go"

	dirDocs                  = "docs"
	dirDocsContent           = "content"
	dirDocsStatic            = "static"
	dirDocsStaticJSONSchemas = "schemas"
	dirDocsData              = "data"
	dirDocsCLIReference      = "reference/cli"

	fileDocsDataLanguages  = "languages.json"
	fileDocsDataMisc       = "misc.json"
	fileDocsDataConfigKeys = "configkeys.json"

	fileDocsStaticJSONSchemasConfiguration      = "configuration"
	fileDocsStaticJSONSchemasUserDatabase       = "user-database"
	fileDocsStaticJSONSchemasExportsTOTP        = "exports.totp"
	fileDocsStaticJSONSchemasExportsWebAuthn    = "exports.webauthn"
	fileDocsStaticJSONSchemasExportsIdentifiers = "exports.identifiers"

	fileGitHubIssueTemplateFR = ".github/ISSUE_TEMPLATE/feature-request.yml"
	fileGitHubIssueTemplateBR = ".github/ISSUE_TEMPLATE/bug-report.yml"
)

const (
	pathJSONSchema = "json-schema"
	extJSON        = ".json"
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
	cmdFlagRoot                                   = "dir.root"
	cmdFlagWeb                                    = "dir.web"
	cmdFlagFileWebI18N                            = "file.web.i18n"
	cmdFlagFileWebPackage                         = "file.web.package"
	cmdFlagDocs                                   = "dir.docs"
	cmdFlagDirLocales                             = "dir.locales"
	cmdFlagDirSchema                              = "dir.schema"
	cmdFlagDirAuthentication                      = "dir.authentication"
	cmdFlagDocsCLIReference                       = "dir.docs.cli-reference"
	cmdFlagDocsContent                            = "dir.docs.content"
	cmdFlagDocsStatic                             = "dir.docs.static"
	cmdFlagDocsStaticJSONSchemas                  = "dir.docs.static.json-schemas"
	cmdFlagDocsData                               = "dir.docs.data"
	cmdFlagDocsDataMisc                           = "file.docs.data.misc"
	cmdFlagDocsDataKeys                           = "file.docs.data.keys"
	cmdFlagDocsDataLanguages                      = "file.docs.data.languages"
	cmdFlagDocsStaticJSONSchemaConfiguration      = "file.docs.static.json-schemas.configuration"
	cmdFlagDocsStaticJSONSchemaUserDatabase       = "file.docs.static.json-schemas.user-database"
	cmdFlagDocsStaticJSONSchemaExportsTOTP        = "file.docs.static.json-schemas.exports.totp"
	cmdFlagDocsStaticJSONSchemaExportsWebAuthn    = "file.docs.static.json-schemas.exports.webauthn"
	cmdFlagDocsStaticJSONSchemaExportsIdentifiers = "file.docs.static.json-schemas.exports.identifiers"
	cmdFlagFileConfigKeys                         = "file.configuration-keys"
	cmdFlagFileScriptsGen                         = "file.scripts.gen"
	cmdFlagFileServerGenerated                    = "file.server.generated"
	cmdFlagFileConfigCommitLint                   = "file.commit-lint-config"
	cmdFlagFileDocsCommitMsgGuidelines            = "file.docs-commit-msg-guidelines"
	cmdFlagFeatureRequest                         = "file.feature-request"
	cmdFlagBugReport                              = "file.bug-report"
	cmdFlagVersions                               = "versions"

	cmdFlagExclude           = "exclude"
	cmdFlagVersionCount      = "version-count"
	cmdFlagCwd               = "cwd"
	cmdFlagPackageConfigKeys = "package.configuration.keys"
	cmdFlagPackageScriptsGen = "package.scripts.gen"
)

const (
	metaVersionNext    = "next"
	metaVersionLatest  = "latest"
	metaVersionCurrent = "current"
)

const (
	codeCSPProductionDefaultSrc  = "'self'"
	codeCSPDevelopmentDefaultSrc = "'self' 'unsafe-eval'"
	codeCSPNonce                 = "${NONCE}"
)

const (
	goModuleBase = "github.com/authelia/authelia/v4"
)

const (
	windows = "windows"
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
