package schema

// Theme represents the configuration related to Styling.
type ThemeConfiguration struct {
        Name                 string `mapstructure:"name"`
        MainColor            string `mapstructure:"maincolor"`
        SecondaryColor       string `mapstructure:"secondarycolor"`
}

// DefaultThemeConfiguration represents default configuration parameters for Theming/Styling.
var DefaultThemeConfiguration = ThemeConfiguration{
	Name: "light",
	MainColor: "#1976d2", //Blue
  SecondaryColor: "#fff",
}
