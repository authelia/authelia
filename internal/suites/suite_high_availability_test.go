package suites

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HighAvailabilityWebDriverSuite struct {
	*SeleniumSuite
}

func NewHighAvailabilityWebDriverSuite() *HighAvailabilityWebDriverSuite {
	return &HighAvailabilityWebDriverSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *HighAvailabilityWebDriverSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *HighAvailabilityWebDriverSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *HighAvailabilityWebDriverSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *HighAvailabilityWebDriverSuite) TestShouldKeepUserDataInDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	secret := s.doRegisterThenLogout(ctx, s.T(), "john", "password")

	err := haDockerEnvironment.Restart("mariadb")
	s.Assert().NoError(err)

	time.Sleep(2 * time.Second)

	s.doLoginTwoFactor(ctx, s.T(), "john", "password", false, secret, "")
	s.verifyIsSecondFactorPage(ctx, s.T())
}

func (s *HighAvailabilityWebDriverSuite) TestShouldKeepSessionAfterAutheliaRestart() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	secret := s.doRegisterAndLogin2FA(ctx, s.T(), "john", "password", false, "")

	err := haDockerEnvironment.Restart("authelia-backend")
	s.Assert().NoError(err)

	loop := true
	for loop {
		logs, err := haDockerEnvironment.Logs("authelia-backend", []string{"--tail", "10"})
		s.Assert().NoError(err)

		select {
		case <-time.After(1 * time.Second):
			if strings.Contains(logs, "Authelia is listening on :9091") {
				loop = false
			}
			break
		case <-ctx.Done():
			loop = false
			break
		}
	}

	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())

	// Verify the user is still authenticated
	s.doVisit(s.T(), LoginBaseURL)
	s.verifyIsSecondFactorPage(ctx, s.T())

	// Then logout and login again to check the secret is still there
	s.doLogout(ctx, s.T())
	s.verifyIsFirstFactorPage(ctx, s.T())

	s.doLoginTwoFactor(ctx, s.T(), "john", "password", false, secret, fmt.Sprintf("%s/secret.html", SecureBaseURL))
	s.verifySecretAuthorized(ctx, s.T())
}

var UserJohn = "john"
var UserBob = "bob"
var UserHarry = "harry"

var Users = []string{UserJohn, UserBob, UserHarry}

var expectedAuthorizations = map[string](map[string]bool){
	fmt.Sprintf("%s/secret.html", PublicBaseURL): map[string]bool{
		UserJohn: true, UserBob: true, UserHarry: true,
	},
	fmt.Sprintf("%s/secret.html", SecureBaseURL): map[string]bool{
		UserJohn: true, UserBob: true, UserHarry: true,
	},
	fmt.Sprintf("%s/secret.html", AdminBaseURL): map[string]bool{
		UserJohn: true, UserBob: false, UserHarry: false,
	},
	fmt.Sprintf("%s/secret.html", SingleFactorBaseURL): map[string]bool{
		UserJohn: true, UserBob: true, UserHarry: true,
	},
	fmt.Sprintf("%s/secret.html", MX1MailBaseURL): map[string]bool{
		UserJohn: true, UserBob: true, UserHarry: false,
	},
	fmt.Sprintf("%s/secret.html", MX2MailBaseURL): map[string]bool{
		UserJohn: false, UserBob: true, UserHarry: false,
	},

	fmt.Sprintf("%s/groups/admin/secret.html", DevBaseURL): map[string]bool{
		UserJohn: true, UserBob: false, UserHarry: false,
	},
	fmt.Sprintf("%s/groups/dev/secret.html", DevBaseURL): map[string]bool{
		UserJohn: true, UserBob: true, UserHarry: false,
	},
	fmt.Sprintf("%s/users/john/secret.html", DevBaseURL): map[string]bool{
		UserJohn: true, UserBob: false, UserHarry: false,
	},
	fmt.Sprintf("%s/users/harry/secret.html", DevBaseURL): map[string]bool{
		UserJohn: true, UserBob: false, UserHarry: true,
	},
	fmt.Sprintf("%s/users/bob/secret.html", DevBaseURL): map[string]bool{
		UserJohn: true, UserBob: true, UserHarry: false,
	},
}

func (s *HighAvailabilityWebDriverSuite) TestShouldVerifyAccessControl() {
	verifyUserIsAuthorized := func(ctx context.Context, t *testing.T, username, targetURL string, authorized bool) {
		s.doVisit(t, targetURL)
		s.verifyURLIs(ctx, t, targetURL)
		if authorized {
			s.verifySecretAuthorized(ctx, t)
		} else {
			s.verifyBodyContains(ctx, t, "403 Forbidden")
		}
	}

	verifyAuthorization := func(username string) func(t *testing.T) {
		return func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			s.doRegisterAndLogin2FA(ctx, t, username, "password", false, "")

			for url, authorizations := range expectedAuthorizations {
				t.Run(url, func(t *testing.T) {
					verifyUserIsAuthorized(ctx, t, username, url, authorizations[username])
				})
			}

			s.doLogout(ctx, t)
		}
	}

	for _, user := range []string{UserJohn, UserBob, UserHarry} {
		s.T().Run(fmt.Sprintf("%s", user), verifyAuthorization(user))
	}
}

type HighAvailabilitySuite struct {
	suite.Suite
}

func NewHighAvailabilitySuite() *HighAvailabilitySuite {
	return &HighAvailabilitySuite{}
}

func DoGetWithAuth(t *testing.T, username, password string) int {
	client := NewHTTPClient()
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/secret.html", SingleFactorBaseURL), nil)
	req.SetBasicAuth(username, password)
	assert.NoError(t, err)

	res, err := client.Do(req)
	assert.NoError(t, err)
	return res.StatusCode
}

func (s *HighAvailabilitySuite) TestBasicAuth() {
	s.Assert().Equal(DoGetWithAuth(s.T(), "john", "password"), 200)
	s.Assert().Equal(DoGetWithAuth(s.T(), "john", "bad-password"), 302)
	s.Assert().Equal(DoGetWithAuth(s.T(), "dontexist", "password"), 302)

}

func TestHighAvailabilitySuite(t *testing.T) {
	suite.Run(t, NewOneFactorScenario())
	suite.Run(t, NewTwoFactorScenario())
	suite.Run(t, NewRegulationScenario())
	suite.Run(t, NewCustomHeadersScenario())
	suite.Run(t, NewRedirectionCheckScenario())
	suite.Run(t, NewHighAvailabilityWebDriverSuite())
	suite.Run(t, NewHighAvailabilitySuite())
}
