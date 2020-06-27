package schema

// Theme represents the configuration related to Duo API.
type ThemeConfiguration struct {
        Name                 string `mapstructure:"name"`
        MainColor            string `mapstructure:"maincolor"`
        SecondaryColor       string `mapstructure:"secondarycolor"`
}

// DefaultTOTPConfiguration represents default configuration parameters for TOTP generation.
var DefaultThemeConfiguration = ThemeConfiguration{
	Name: "light",
	MainColor: "#000",
  SecondaryColor: "#fff",
}
