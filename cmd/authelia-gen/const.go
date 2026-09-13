// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

const (
	dirCurrent = "./"
	dirEvents  = "internal/events"

	dirRepositoryRoot = "../.."

	dirLocales = "internal/server/locales"
	dirWeb     = "web"

	subPathCmd      = "cmd"
	subPathInternal = "internal"

	fileCICommitLintConfig = "commitlint.config.mjs"
	fileWebI18NIndex       = "src/i18n/index.ts"
	fileWebPackage         = "package.json"

	fileDocsCommitMessageGuidelines = "docs/content/contributing/guidelines/commit-message.md"

	fileCodeConfigKeys  = "internal/configuration/schema/keys.go"
	fileServerGenerated = "internal/server/gen.go"
	fileScriptsGen      = "cmd/authelia-scripts/cmd/gen.go"

	dirDocs                          = "docs"
	dirDocsContent                   = "content"
	dirDocsStatic                    = "static"
	dirDocsStaticJSONSchemas         = "schemas"
	dirDocsStaticJSONSchemasWebhooks = "webhooks"
	dirDocsData                      = "data"
	dirDocsADR                       = "reference/architecture-decision-log"
	dirDocsCLIReference              = "reference/cli"

	fileDocsDataLanguages  = "languages.json"
	fileDocsDataMisc       = "misc.json"
	fileDocsDataConfigKeys = "configkeys.json"

	fileDocsStaticJSONSchemasConfiguration      = "configuration"
	fileDocsStaticJSONSchemasUserDatabase       = "user-database"
	fileDocsStaticJSONSchemasExportsTOTP        = "exports.totp"
	fileDocsStaticJSONSchemasExportsWebAuthn    = "exports.webauthn"
	fileDocsStaticJSONSchemasExportsIdentifiers = "exports.identifiers"
	fileDocsStaticJSONSchemasWebhooksEnvelope   = "envelope"
	fileDocsStaticJSONSchemasWebhooksBatch      = "envelope-batch"

	dirDocsStaticImages             = "images"
	dirDocsStaticImagesContributors = "contributors"
	fileAllContributors             = ".all-contributorsrc"
	fileREADME                      = "README.md"

	extSVG = ".svg"

	fileGitHubIssueTemplateFR = ".github/ISSUE_TEMPLATE/feature-request.yml"
	fileGitHubIssueTemplateBR = ".github/ISSUE_TEMPLATE/bug-report.yml"
)

const (
	pathJSONSchema = "json-schema"
	extJSON        = ".json"
	extYAML        = ".yaml"
)

const urlFormatJSONSchemaWebhooks = "https://www.authelia.com/schemas/webhooks/v%s/%s.json"

const (
	dateFmtRFC2822 = "Mon, _2 Jan 2006 15:04:05 -0700"
	dateFmtYAML    = "2006-01-02T15:04:05-07:00"
)

const (
	delimiterLineFrontMatter = "---"
)

const (
	pkgConfigSchema = "schema"
	pkgScriptsGen   = "cmd"
)

const (
	cmdUseRoot                   = "authelia-gen"
	cmdUseCompletion             = "completion"
	cmdUseDocs                   = "docs"
	cmdUseManage                 = "manage"
	cmdUseMisc                   = "misc"
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
	cmdUseContributors           = "contributors"
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
	cmdFlagDirEvents                              = "dir.events"
	cmdFlagDocsCLIReference                       = "dir.docs.cli-reference"
	cmdFlagDocsContent                            = "dir.docs.content"
	cmdFlagDocsStatic                             = "dir.docs.static"
	cmdFlagDocsStaticJSONSchemas                  = "dir.docs.static.json-schemas"
	cmdFlagDocsStaticJSONSchemasWebhooks          = "dir.docs.static.json-schemas.webhooks"
	cmdFlagDocsData                               = "dir.docs.data"
	cmdFlagDocsADR                                = "dir.docs.adr"
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
	codeCSPDirectiveDefaultSrc   = "default-src"
	codeCSPSelf                  = "'self'"
	codeCSPDevelopmentDefaultSrc = "'self' 'unsafe-eval'"
	codeCSPNone                  = "'none'"
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
		{Name: codeCSPDirectiveDefaultSrc, Value: codeCSPSelf},
		{Name: "frame-src", Value: codeCSPNone},
		{Name: "object-src", Value: codeCSPNone},
		{Name: "style-src", Value: "'self' 'nonce-%s'"},
		{Name: "frame-ancestors", Value: codeCSPNone},
		{Name: "base-uri", Value: codeCSPSelf},
	}

	codeCSPValuesProduction = []CSPValue{
		{Name: "connect-src", Value: codeCSPSelf},
		{Name: "script-src", Value: codeCSPSelf},
	}

	codeCSPValuesDevelopment = []CSPValue{
		{Name: codeCSPDirectiveDefaultSrc, Value: codeCSPDevelopmentDefaultSrc},
	}
)

const (
	contributorsCellWidth         = 120
	contributorsCellHeight        = 100
	contributorsAvatar            = 56
	contributorsOverlordAvatar    = 40
	contributorsMaxEmoji          = 4
	contributorsMaxName           = 16
	contributorsAvatarSize        = 112
	contributorsAvatarQuality     = 68
	contributorsAvatarConcurrency = 12
	contributorsAvatarRedirects   = 3

	contributorsAvatarHost = "avatars.githubusercontent.com"
	contributorsAvatarPath = "/u/"

	contributorsContributionDoc = "doc"

	schemeHTTP  = "http"
	schemeHTTPS = "https"

	contributorsEmojiKey  = "https://allcontributors.org/en/reference/emoji-key/"
	contributorsImageURL  = "https://www.authelia.com/images/contributors"
	contributorsFont      = "'Inter', ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif"
	contributorsEmojiFont = "'Apple Color Emoji', 'Segoe UI Emoji', 'Noto Color Emoji', 'Twemoji Mozilla', sans-serif"

	contributorsMarkerStart = "<!-- ALL-CONTRIBUTORS-LIST:START -->"
	contributorsMarkerEnd   = "<!-- ALL-CONTRIBUTORS-LIST:END -->"
)

// contributorsEmoji maps the all-contributors contribution types to the emoji from their key.
var contributorsEmoji = map[string]string{
	"a11y":              "\u267F\uFE0F",               // Accessibility.
	"audio":             "\U0001F50A",                 // Audio.
	"blog":              "\U0001F4DD",                 // Blogposts.
	"bug":               "\U0001F41B",                 // Bug reports.
	"business":          "\U0001F4BC",                 // Business development.
	"code":              "\U0001F4BB",                 // Code.
	"content":           "\U0001F58B",                 // Content.
	"data":              "\U0001F523",                 // Data.
	"design":            "\U0001F3A8",                 // Design.
	"doc":               "\U0001F4D6",                 // Documentation.
	"eventOrganizing":   "\U0001F4CB",                 // Event Organizing.
	"example":           "\U0001F4A1",                 // Examples.
	"financial":         "\U0001F4B5",                 // Financial.
	"fundingFinding":    "\U0001F50D",                 // Funding Finding.
	"ideas":             "\U0001F914",                 // Ideas, Planning, & Feedback.
	"infra":             "\U0001F687",                 // Infrastructure (Hosting, Build-Tools, etc).
	"maintenance":       "\U0001F6A7",                 // Maintenance.
	"mentoring":         "\U0001F9D1\u200D\U0001F3EB", // Mentoring.
	"platform":          "\U0001F4E6",                 // Packaging/porting to new platform.
	"plugin":            "\U0001F50C",                 // Plugin/utility libraries.
	"projectManagement": "\U0001F4C6",                 // Project Management.
	"promotion":         "\U0001F4E3",                 // Promotion.
	"question":          "\U0001F4AC",                 // Answering Questions.
	"research":          "\U0001F52C",                 // Research.
	"review":            "\U0001F440",                 // Reviewed Pull Requests.
	"security":          "\U0001F6E1\uFE0F",           // Security.
	"talk":              "\U0001F4E2",                 // Talks.
	"test":              "\u26A0\uFE0F",               // Tests.
	"tool":              "\U0001F527",                 // Tools.
	"translation":       "\U0001F30D",                 // Translation.
	"tutorial":          "\u2705",                     // Tutorials.
	"userTesting":       "\U0001F4D3",                 // User Testing.
	"video":             "\U0001F4F9",                 // Videos.
}
