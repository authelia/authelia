package validator

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type Theme struct {
	suite.Suite
	config    *schema.Configuration
	validator *schema.StructValidator
}

func (suite *Theme) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = &schema.Configuration{
		Theme: "light",
	}
}

func (suite *Theme) TestShouldValidateCompleteConfiguration() {
	ValidateTheme(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *Theme) TestShouldRaiseErrorWhenInvalidThemeProvided() {
	suite.config.Theme = testInvalid

	ValidateTheme(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "option 'theme' must be one of 'light', 'dark', 'grey', or 'auto' but it's configured as 'invalid'")
}

func TestThemes(t *testing.T) {
	suite.Run(t, new(Theme))
}
