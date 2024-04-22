package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TwoFactorOneTimePasswordSuite struct {
	*RodSuite
}

func NewTwoFactorOneTimePasswordScenario() *TwoFactorOneTimePasswordSuite {
	return &TwoFactorOneTimePasswordSuite{
		RodSuite: NewRodSuite(""),
	}
}

func (s *TwoFactorOneTimePasswordSuite) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *TwoFactorOneTimePasswordSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *TwoFactorOneTimePasswordSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *TwoFactorOneTimePasswordSuite) TestShouldRegisterAllAdvancedOptions() {
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.Page = s.doCreateTab(s.T(), LoginBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", true, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.doOpenSettingsMenuClickTwoFactor(s.T(), s.Context(ctx))

	algorithms := []string{SHA1, SHA256, SHA512}
	lengths := []int{6, 8}
	periods := []int{30, 60, 90, 120}

	for a, algorithm := range algorithms {
		for l, length := range lengths {
			for p, period := range periods {
				s.T().Run(fmt.Sprintf("%s-%d-%d", algorithm, length, period), func(t *testing.T) {
					s.doRegisterTOTPAdvanced(t, s.Context(ctx), a+l+p == 0, "john", algorithm, length, period)
				})
			}
		}
	}
}

func TestRunTwoFactorOneTimePassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTwoFactorOneTimePasswordScenario())
}
