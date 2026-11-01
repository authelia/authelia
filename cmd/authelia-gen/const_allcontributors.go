package main

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

	contributorsEmojiKey  = "https://allcontributors.org/en/reference/emoji-key/"
	contributorsImageURL  = "https://www.authelia.com/images/contributors"
	contributorsFont      = "'Inter', ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif"
	contributorsEmojiFont = "'Apple Color Emoji', 'Segoe UI Emoji', 'Noto Color Emoji', 'Twemoji Mozilla', sans-serif"

	contributorsMarkerStart = "<!-- ALL-CONTRIBUTORS-LIST:START -->"
	contributorsMarkerEnd   = "<!-- ALL-CONTRIBUTORS-LIST:END -->"
)

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
