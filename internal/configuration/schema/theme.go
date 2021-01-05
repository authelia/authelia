package schema

// ThemeConfiguration represents the configuration related to styling.
type ThemeConfiguration struct {
	Name           string `mapstructure:"name"`
	PrimaryColor   string `mapstructure:"primary_color"`
	SecondaryColor string `mapstructure:"secondary_color"`
}

// DefaultThemeConfiguration represents default configuration parameters for styling.
var DefaultThemeConfiguration = ThemeConfiguration{
	Name:           "light",
	PrimaryColor:   "#1976d2",
	SecondaryColor: "#ffffff",
}
