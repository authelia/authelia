package templates

const (
	extText = ".txt"
	extHTML = ".html"
)

// Template File Names.
const (
	TemplateNameEmailIdentityVerificationJWT = "IdentityVerificationJWT"
	TemplateNameEmailIdentityVerificationOTC = "IdentityVerificationOTC"
	TemplateNameEmailEvent                   = "Event"

	TemplateNameOIDCAuthorizeFormPost = "AuthorizeResponseFormPost.html"
)

// Templated Asset Paths, relative to the root of the embedded asset filesystem.
const (
	AssetPathAPIIndex = "public_html/api/index.html"
	AssetPathAPISpec  = "public_html/api/openapi.yml"
	AssetPathIndex    = "public_html/index.html"

	assetPathPrefix = "assets/"
)

// AssetPathsTemplated are the embedded assets rendered per request rather than served as they are embedded.
var AssetPathsTemplated = []string{
	AssetPathAPIIndex,
	AssetPathAPISpec,
	AssetPathIndex,
}

// Template Category Names.
const (
	TemplateCategoryNotifications = "notification"
	TemplateCategoryOpenIDConnect = "oidc"
)
