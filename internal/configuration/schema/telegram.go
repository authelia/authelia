package schema

// Telegram represents the configuration for the Telegram authentication provider.
type Telegram struct {
	// BotToken is the Telegram Bot API token.
	BotToken string `koanf:"bot_token" yaml:"bot_token" json:"bot_token" jsonschema:"title=Bot Token" jsonschema_description:"The Telegram Bot API token."`

	// BotUsername is the Telegram bot username (without @).
	BotUsername string `koanf:"bot_username" yaml:"bot_username" json:"bot_username" jsonschema:"title=Bot Username" jsonschema_description:"The Telegram bot username."`

	// TokenTTL is how long a verification token is valid (in seconds).
	TokenTTL int `koanf:"token_ttl" yaml:"token_ttl" json:"token_ttl" jsonschema:"title=Token TTL" jsonschema_description:"How long a verification token is valid in seconds." default:"300"`
}
