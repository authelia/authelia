package schema

// Theme represents the configuration related to Styling.
type ThemeConfiguration struct {
        Name                 string `mapstructure:"name"`
        PrimaryColor        string `mapstructure:"primary_color"`
        SecondaryColor      string `mapstructure:"secondary_color"`
}

// DefaultThemeConfiguration represents default configuration parameters for Theming/Styling.
var DefaultThemeConfiguration = ThemeConfiguration{
	Name: "light",
	PrimaryColor: "#1976d2", //Blue
	SecondaryColor: "#fff",
}
