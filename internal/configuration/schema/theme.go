package schema

// Theme represents the configuration related to Duo API.
type ThemeConfiguration struct {
        Theme      string `mapstructure:"theme"`
}

// DefaultTOTPConfiguration represents default configuration parameters for TOTP generation.
var DefaultThemeConfiguration = ThemeConfiguration{
        Theme: "light",
}
