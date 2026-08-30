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
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/utils"
)

type HighAvailabilityWebDriverSuite struct {
	*RodSuite
}

func NewHighAvailabilityWebDriverSuite() *HighAvailabilityWebDriverSuite {
	return &HighAvailabilityWebDriverSuite{
		RodSuite: NewRodSuite(""),
	}
}

func (s *HighAvailabilityWebDriverSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *HighAvailabilityWebDriverSuite) TearDownSuite() {
	err := s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *HighAvailabilityWebDriverSuite) SetupTest() {
	s.doSetupTest(HomeBaseURL)
}

func (s *HighAvailabilityWebDriverSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

var redisNodeServices = map[string]string{
	"redis-node-0": SuiteAddress(110),
	"redis-node-1": SuiteAddress(111),
	"redis-node-2": SuiteAddress(112),
}

func (s *HighAvailabilityWebDriverSuite) redisMaster(sentinel string) (master string) {
	// Which node holds the master role depends on whichever failovers earlier tests provoked, so it
	// cannot be assumed from the initial configuration. The node is also polled until it answers,
	// because sentinel reports an address before the node behind it is accepting connections again.
	err := utils.CheckUntil(time.Second, redisMasterTimeout, func() (bool, error) {
		output, err := haDockerEnvironment.Exec(sentinel, []string{
			"redis-cli", "-p", "26379", "-a", "sentinel-server-password", "--no-auth-warning",
			"sentinel", "get-master-addr-by-name", "authelia",
		})
		if err != nil {
			return false, nil
		}

		for service, address := range redisNodeServices {
			if !strings.Contains(output, address) {
				continue
			}

			// Ask the node itself rather than trusting the address alone. A single sentinel can still be
			// serving its pre-failover view, and a former master restarted by an earlier test answers PING
			// perfectly well while it is a replica, so accepting either would hand back the wrong node.
			role, rerr := haDockerEnvironment.Exec(service, []string{"redis-cli", "role"})
			if rerr != nil || !strings.HasPrefix(strings.TrimSpace(role), "master") {
				return false, nil
			}

			master = service

			return true, nil
		}

		return false, nil
	})

	s.Require().NoError(err, "Could not determine an available redis master")

	return master
}

func (s *HighAvailabilityWebDriverSuite) redisReplica(sentinel string) string {
	master := s.redisMaster(sentinel)

	for service := range redisNodeServices {
		if service != master {
			return service
		}
	}

	s.Require().FailNow("Could not determine a redis replica")

	return ""
}

func (s *HighAvailabilityWebDriverSuite) doStartAndSettle(service, observer string) {
	since := time.Now()

	s.Require().NoError(haDockerEnvironment.Start(service))

	// Reported for anything the sentinel had marked down, so it covers a returning node and a returning
	// sentinel alike.
	s.Require().NoError(waitUntilServiceLog(haDockerEnvironment, observer, "-sdown", since))

	// Resolving the master confirms the sentinels agree on one and that the node behind it answers as one.
	s.redisMaster(observer)
}

func (s *HighAvailabilityWebDriverSuite) TestShouldKeepUserSessionActive() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginAndRegisterTOTPThenLogout(s.T(), s.Context(ctx), "john", "password")

	err := haDockerEnvironment.Restart(s.redisMaster("redis-sentinel-0"))
	s.Require().NoError(err)

	// Restarting the master takes the session backend away with it, and a login attempted before it
	// returns fails at the first factor.
	s.redisMaster("redis-sentinel-0")

	doWithDisruptedDatastore(func() {
		s.doLoginSecondFactorTOTP(s.T(), s.Context(ctx), "john", "password", false, "")
		s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	})
}

func (s *HighAvailabilityWebDriverSuite) TestShouldKeepUserSessionActiveWithPrimaryRedisNodeFailure() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginAndRegisterTOTPThenLogout(s.T(), s.Context(ctx), "john", "password")

	s.doLoginSecondFactorTOTP(s.T(), s.Context(ctx), "john", "password", false, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))

	master := s.redisMaster("redis-sentinel-0")

	since := time.Now()

	err := haDockerEnvironment.Stop(master)
	s.Require().NoError(err)

	defer func() {
		s.doStartAndSettle(master, "redis-sentinel-0")
	}()

	s.Require().NoError(waitUntilServiceLog(haDockerEnvironment, "redis-sentinel-0", "+switch-master authelia", since))

	doWithDisruptedDatastore(func() {
		s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
		s.verifyIsHome(s.T(), s.Context(ctx))

		// Verify the user is still authenticated.
		s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURL(BaseDomain))
		s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))

		s.doLogout(s.T(), s.Context(ctx))
		s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))

		s.doLoginSecondFactorTOTP(s.T(), s.Context(ctx), "john", "password", false, fmt.Sprintf("%s/secret.html", SecureBaseURL))
		s.verifySecretAuthorized(s.T(), s.Context(ctx))
	})
}

func (s *HighAvailabilityWebDriverSuite) TestShouldKeepUserSessionActiveWithPrimaryRedisSentinelFailureAndSecondaryRedisNodeFailure() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginAndRegisterTOTPThenLogout(s.T(), s.Context(ctx), "john", "password")

	s.doLoginSecondFactorTOTP(s.T(), s.Context(ctx), "john", "password", false, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))

	replica := s.redisReplica("redis-sentinel-0")

	since := time.Now()

	err := haDockerEnvironment.Stop("redis-sentinel-0")
	s.Require().NoError(err)

	defer func() {
		s.doStartAndSettle("redis-sentinel-0", "redis-sentinel-1")
	}()

	err = haDockerEnvironment.Stop(replica)
	s.Require().NoError(err)

	defer func() {
		s.doStartAndSettle(replica, "redis-sentinel-1")
	}()

	s.Require().NoError(waitUntilServiceLog(haDockerEnvironment, "redis-sentinel-1", "+sdown sentinel", since))
	s.Require().NoError(waitUntilServiceLog(haDockerEnvironment, "redis-sentinel-1", "+sdown slave", since))

	doWithDisruptedDatastore(func() {
		s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
		s.verifyIsHome(s.T(), s.Context(ctx))

		// Verify the user is still authenticated.
		s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURL(BaseDomain))
		s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	})
}

func (s *HighAvailabilityWebDriverSuite) TestShouldKeepUserDataInDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginAndRegisterTOTPThenLogout(s.T(), s.Context(ctx), "john", "password")

	since := time.Now()

	err := haDockerEnvironment.Restart("mariadb")
	s.Require().NoError(err)

	s.Require().NoError(waitUntilServiceLog(haDockerEnvironment, "mariadb", "mariadbd: ready for connections", since))

	doWithDisruptedDatastore(func() {
		s.doLoginSecondFactorTOTP(s.T(), s.Context(ctx), "john", "password", false, "")
		s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	})
}

func (s *HighAvailabilityWebDriverSuite) TestShouldKeepSessionAfterAutheliaRestart() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doRegisterTOTPAndLogin2FA(s.T(), s.Context(ctx), "john", "password", false, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))

	since := time.Now()

	err := haDockerEnvironment.Restart("authelia-backend")
	s.Require().NoError(err)

	err = waitUntilAutheliaBackendIsReady(haDockerEnvironment, since)
	s.Require().NoError(err)

	doWithDisruptedDatastore(func() {
		s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
		s.verifyIsHome(s.T(), s.Context(ctx))

		// Verify the user is still authenticated.
		s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURL(BaseDomain))
		s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))

		s.doLogout(s.T(), s.Context(ctx))
		s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))

		s.doLoginSecondFactorTOTP(s.T(), s.Context(ctx), "john", "password", false, fmt.Sprintf("%s/secret.html", SecureBaseURL))
		s.verifySecretAuthorized(s.T(), s.Context(ctx))
	})
}

var UserJohn = "john"

var UserBob = "bob"

var UserHarry = "harry"

var Users = []string{UserJohn, UserBob, UserHarry}

var expectedAuthorizations = map[string](map[string]bool){
	fmt.Sprintf("%s/secret.html", PublicBaseURL): {
		UserJohn: true, UserBob: true, UserHarry: true,
	},
	fmt.Sprintf("%s/secret.html", SecureBaseURL): {
		UserJohn: true, UserBob: true, UserHarry: true,
	},
	fmt.Sprintf("%s/secret.html", AdminBaseURL): {
		UserJohn: true, UserBob: false, UserHarry: false,
	},
	fmt.Sprintf("%s/secret.html", SingleFactorBaseURL): {
		UserJohn: true, UserBob: true, UserHarry: true,
	},
	fmt.Sprintf("%s/secret.html", MX1MailBaseURL): {
		UserJohn: true, UserBob: true, UserHarry: false,
	},
	fmt.Sprintf("%s/secret.html", MX2MailBaseURL): {
		UserJohn: false, UserBob: true, UserHarry: false,
	},

	fmt.Sprintf("%s/groups/admin/secret.html", DevBaseURL): {
		UserJohn: true, UserBob: false, UserHarry: false,
	},
	fmt.Sprintf("%s/groups/dev/secret.html", DevBaseURL): {
		UserJohn: true, UserBob: true, UserHarry: false,
	},
	fmt.Sprintf("%s/users/john/secret.html", DevBaseURL): {
		UserJohn: true, UserBob: false, UserHarry: false,
	},
	fmt.Sprintf("%s/users/harry/secret.html", DevBaseURL): {
		UserJohn: true, UserBob: false, UserHarry: true,
	},
	fmt.Sprintf("%s/users/bob/secret.html", DevBaseURL): {
		UserJohn: true, UserBob: true, UserHarry: false,
	},
}

func (s *HighAvailabilityWebDriverSuite) TestShouldVerifyAccessControl() {
	verifyUserIsAuthorized := func(ctx context.Context, t *testing.T, targetURL string, authorized bool) {
		s.doVisit(t, s.Context(ctx), targetURL)
		s.verifyURLIs(t, s.Context(ctx), targetURL)

		if authorized {
			s.verifySecretAuthorized(t, s.Context(ctx))
		} else {
			s.verifyBodyContains(t, s.Context(ctx), "403 Forbidden")
		}
	}

	verifyAuthorization := func(username string) func(t *testing.T) {
		return func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			defer func() {
				s.collectScreenshot(ctx.Err(), s.Page)
				cancel()
			}()

			s.doRegisterTOTPAndLogin2FA(t, s.Context(ctx), username, "password", false, "")

			for url, authorizations := range expectedAuthorizations {
				t.Run(url, func(t *testing.T) {
					verifyUserIsAuthorized(ctx, t, url, authorizations[username])
				})
			}

			s.doLogout(t, s.Context(ctx))
		}
	}

	for _, user := range Users {
		s.T().Run(user, verifyAuthorization(user))
	}
}

type HighAvailabilitySuite struct {
	*BaseSuite
}

func NewHighAvailabilitySuite() *HighAvailabilitySuite {
	return &HighAvailabilitySuite{
		BaseSuite: &BaseSuite{
			Name: highAvailabilitySuiteName,
		},
	}
}

func DoGetWithAuth(t *testing.T, username, password string) int {
	t.Helper()

	client := NewHTTPClient()
	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("%s/secret.html", SingleFactorBaseURL), nil)
	assert.NoError(t, err)
	req.SetBasicAuth(username, password)

	res, err := client.Do(req)
	assert.NoError(t, err)

	return res.StatusCode
}

func (s *HighAvailabilitySuite) TestBasicAuth() {
	s.Assert().Equal(fasthttp.StatusOK, DoGetWithAuth(s.T(), "john", "password"))
	s.Assert().Equal(fasthttp.StatusFound, DoGetWithAuth(s.T(), "john", "bad-password"))
	s.Assert().Equal(fasthttp.StatusFound, DoGetWithAuth(s.T(), "dontexist", "password"))
}

func (s *HighAvailabilitySuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *HighAvailabilitySuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *HighAvailabilitySuite) TestRegulationScenario() {
	suite.Run(s.T(), NewRegulationScenario())
}

func (s *HighAvailabilitySuite) TestCustomHeadersScenario() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *HighAvailabilitySuite) TestRedirectionCheckScenario() {
	suite.Run(s.T(), NewRedirectionCheckScenario())
}

func (s *HighAvailabilitySuite) TestHighAvailabilityWebDriverSuite() {
	suite.Run(s.T(), NewHighAvailabilityWebDriverSuite())
}

func TestHighAvailabilityWebDriverSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewHighAvailabilityWebDriverSuite())
}

func TestHighAvailabilitySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewHighAvailabilitySuite())
}
