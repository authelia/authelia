package suites

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/poy/onpar"
)

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

func DoGetWithAuth(t *testing.T, username, password string) int {
	is := is.New(t)
	client := NewHTTPClient()
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/secret.html", SingleFactorBaseURL), nil)
	is.NoErr(err)
	req.SetBasicAuth(username, password)

	res, err := client.Do(req)
	is.NoErr(err)

	return res.StatusCode
}

func TestHighAvailabilitySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	s := setupTest(t, "", true)
	teardownTest(s)

	o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
		s := setupTest(t, "", false)
		return t, s
	})

	o.AfterEach(func(t *testing.T, s RodSuite) {
		teardownTest(s)
	})

	o.Spec("TestBasicAuth", func(t *testing.T, s RodSuite) {
		is := is.New(t)
		is.Equal(DoGetWithAuth(t, testUsername, testPassword), 200)
		is.Equal(DoGetWithAuth(t, testUsername, badPassword), 302)
		is.Equal(DoGetWithAuth(t, "dontexist", testPassword), 302)
	})

	o.Spec("TestShouldVerifyAccessControl", func(t *testing.T, s RodSuite) {
		verifyUserIsAuthorized := func(ctx context.Context, t *testing.T, targetURL string, authorized bool) {
			s.doVisit(s.Context(ctx), targetURL)
			s.verifyURLIs(t, s.Context(ctx), targetURL)

			if authorized {
				s.verifySecretAuthorized(t, s.Context(ctx))
			} else {
				s.verifyBodyContains(t, s.Context(ctx), "403 Forbidden")
			}
		}

		verifyAuthorization := func(username string) func(t *testing.T) {
			return func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer func() {
					s.collectScreenshot(ctx.Err(), s.Page)
					cancel()
				}()

				if username == testUsername {
					s.doLoginTwoFactor(t, s.Context(ctx), username, testPassword, false, secret, "")
				} else {
					s.doRegisterAndLogin2FA(t, s.Context(ctx), username, testPassword, false, "")
				}

				for url, authorizations := range expectedAuthorizations {
					t.Run(url, func(t *testing.T) {
						verifyUserIsAuthorized(ctx, t, url, authorizations[username])
					})
				}

				s.doLogout(t, s.Context(ctx))
			}
		}

		for _, user := range Users {
			t.Run(user, verifyAuthorization(user))
		}
	})

	TestRun1FAScenario(t)
	TestRun2FAScenario(t)
	TestRunCustomHeadersScenario(t)
	TestRunRedirectionCheckScenario(t)
	TestRunRegulationScenario(t)
	t.Run("TestShouldKeepUserSessionActive", TestShouldKeepUserSessionActive)
	t.Run("TestShouldKeepUserSessionActiveWithPrimaryRedisNodeFailure", TestShouldKeepUserSessionActiveWithPrimaryRedisNodeFailure)
	t.Run("TestShouldKeepUserSessionActiveWithPrimaryRedisSentinelFailureAndSecondaryRedisNodeFailure", TestShouldKeepUserSessionActiveWithPrimaryRedisSentinelFailureAndSecondaryRedisNodeFailure)
	t.Run("TestShouldKeepUserDataInDB", TestShouldKeepUserDataInDB)
	t.Run("TestShouldKeepSessionAfterAutheliaRestart", TestShouldKeepSessionAfterAutheliaRestart)
}

func TestShouldKeepUserSessionActive(t *testing.T) {
	s := setupTest(t, "", false)
	is := is.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownTest(s)
	}()

	err := haDockerEnvironment.Restart("redis-node-0")
	is.NoErr(err)

	s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, "")
	s.verifyIsSecondFactorPage(t, s.Context(ctx))
}

func TestShouldKeepUserSessionActiveWithPrimaryRedisNodeFailure(t *testing.T) {
	s := setupTest(t, "", false)
	is := is.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownTest(s)
	}()

	s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, "")
	s.verifyIsSecondFactorPage(t, s.Context(ctx))

	err := haDockerEnvironment.Stop("redis-node-0")
	is.NoErr(err)

	defer func() {
		err = haDockerEnvironment.Start("redis-node-0")
		is.NoErr(err)
	}()

	s.doVisit(s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(t, s.Context(ctx))

	// Verify the user is still authenticated.
	s.doVisit(s.Context(ctx), GetLoginBaseURL())
	s.verifyIsSecondFactorPage(t, s.Context(ctx))

	// Then logout and login again to check we can see the secret.
	s.doLogout(t, s.Context(ctx))
	s.verifyIsFirstFactorPage(t, s.Context(ctx))

	s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, fmt.Sprintf("%s/secret.html", SecureBaseURL))
	s.verifySecretAuthorized(t, s.Context(ctx))
}

func TestShouldKeepUserSessionActiveWithPrimaryRedisSentinelFailureAndSecondaryRedisNodeFailure(t *testing.T) {
	s := setupTest(t, "", false)
	is := is.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownTest(s)
	}()

	s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, "")
	s.verifyIsSecondFactorPage(t, s.Context(ctx))

	err := haDockerEnvironment.Stop("redis-sentinel-0")
	is.NoErr(err)

	defer func() {
		err = haDockerEnvironment.Start("redis-sentinel-0")
		is.NoErr(err)
	}()

	err = haDockerEnvironment.Stop("redis-node-2")
	is.NoErr(err)

	defer func() {
		err = haDockerEnvironment.Start("redis-node-2")
		is.NoErr(err)
	}()

	s.doVisit(s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(t, s.Context(ctx))

	// Verify the user is still authenticated.
	s.doVisit(s.Context(ctx), GetLoginBaseURL())
	s.verifyIsSecondFactorPage(t, s.Context(ctx))
}

func TestShouldKeepUserDataInDB(t *testing.T) {
	s := setupTest(t, "", false)
	is := is.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownTest(s)
	}()

	err := haDockerEnvironment.Restart("mariadb")
	is.NoErr(err)

	s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, "")
	s.verifyIsSecondFactorPage(t, s.Context(ctx))
}

func TestShouldKeepSessionAfterAutheliaRestart(t *testing.T) {
	s := setupTest(t, "", false)
	is := is.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownTest(s)
	}()

	s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")
	s.verifyIsSecondFactorPage(t, s.Context(ctx))

	err := haDockerEnvironment.Restart("authelia-backend")
	is.NoErr(err)

	err = waitUntilAutheliaBackendIsReady(haDockerEnvironment)
	is.NoErr(err)

	s.doVisit(s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(t, s.Context(ctx))

	// Verify the user is still authenticated.
	s.doVisit(s.Context(ctx), GetLoginBaseURL())
	s.verifyIsSecondFactorPage(t, s.Context(ctx))

	// Then logout and login again to check the secret is still there.
	s.doLogout(t, s.Context(ctx))
	s.verifyIsFirstFactorPage(t, s.Context(ctx))

	s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, fmt.Sprintf("%s/secret.html", SecureBaseURL))
	s.verifySecretAuthorized(t, s.Context(ctx))
}
