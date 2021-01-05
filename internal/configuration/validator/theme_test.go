package validator

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
)

type Theme struct {
	suite.Suite
	configuration *schema.ThemeConfiguration
	validator     *schema.StructValidator
}

func (suite *Theme) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration = &schema.ThemeConfiguration{
		Name:           "light",
		PrimaryColor:   "#1976d2",
		SecondaryColor: "#ffffff",
	}
}

func (suite *Theme) TestShouldValidateCompleteConfiguration() {
	ValidateTheme(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *Theme) TestShouldRaiseErrorWhenInvalidThemeProvided() {
	suite.configuration.Name = "invalid"

	ValidateTheme(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Theme: invalid is not valid, valid themes are: \"light\", \"dark\", \"grey\" or \"custom\"")
}

func (suite *Theme) TestShouldRaiseWarningsWhenColorsNotProvided() {
	suite.configuration.Name = themeCustom
	suite.configuration.PrimaryColor = ""
	suite.configuration.SecondaryColor = ""

	ValidateTheme(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasErrors())
	suite.Require().Len(suite.validator.Warnings(), 2)

	suite.Assert().EqualError(suite.validator.Warnings()[0], "Theme primary color has not been specified, defaulting to: #1976d2")
	suite.Assert().EqualError(suite.validator.Warnings()[1], "Theme secondary color has not been specified, defaulting to: #ffffff")
}

func (suite *Theme) TestShouldRaiseErrorWhenInvalidPrimaryColor() {
	suite.configuration.Name = themeCustom
	suite.configuration.PrimaryColor = "abcd"

	ValidateTheme(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Theme primary color: abcd is not valid, valid color values are hex from: \"#000000\" to \"#FFFFFF\"")
}

func (suite *Theme) TestShouldRaiseErrorWhenInvalidSecondaryColor() {
	suite.configuration.Name = themeCustom
	suite.configuration.SecondaryColor = "dcbe"

	ValidateTheme(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Theme secondary color: dcbe is not valid, valid color values are hex from: \"#000000\" to \"#FFFFFF\"")
}

func TestThemes(t *testing.T) {
	suite.Run(t, new(Theme))
}
