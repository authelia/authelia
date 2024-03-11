package schema

// CustomLocales represents the configuration related to custom locales.
type CustomLocales struct {
	Enabled bool   `koanf:"enabled" json:"enabled" jsonschema:"title=Enabled" jsonschema_description:"Enable custom Locales."`
	Path    string `koanf:"path" json:"path" jsonschema:"title=Path" jsonschema_description:"path to the custom locales folder."`
}

// DefaultCustomLocalesConfiguration is the default custom locales configuration.
var DefaultCustomLocalesConfiguration = CustomLocales{
	Enabled: false,
	Path:    "",
}
