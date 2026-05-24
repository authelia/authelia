package suites

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ssh"

	"github.com/authelia/otp"
	"github.com/authelia/otp/totp"

	"github.com/authelia/authelia/v4/internal/storage"
)

type PAMSuite struct {
	*RodSuite
}

func NewPAMSuite() *PAMSuite {
	return &PAMSuite{
		RodSuite: NewRodSuite(pamSuiteName),
	}
}

func (s *PAMSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.seedUserTOTP("john")
	s.seedUserTOTP("jane")

	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *PAMSuite) TearDownSuite() {
	err := s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *PAMSuite) TestShouldAuthenticateWith1FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s.setPAMAuthLevel("1FA")

	since := time.Now()

	output, err := s.sshLogin(ctx, "john", []string{"password"})
	s.Require().NoError(err, "1FA SSH login should succeed: %s", output)

	logs := s.pamLogsSince(since)
	s.Contains(logs, "POST https://login.example.com:8080/api/firstfactor")
	s.Contains(logs, `response status=200 status_field="OK"`)
	s.NotContains(logs, "/api/secondfactor/")
	s.Contains(logs, "Accepted keyboard-interactive/pam for john")
}

func (s *PAMSuite) TestShouldReject1FAWithBadPassword() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s.setPAMAuthLevel("1FA")

	since := time.Now()

	output, err := s.sshLogin(ctx, "john", []string{"wrongpassword"})
	s.Error(err, "1FA SSH with bad password should fail: %s", output)

	logs := s.pamLogsSince(since)
	s.Contains(logs, "POST https://login.example.com:8080/api/firstfactor")
	s.Contains(logs, "response status=401")
	s.Contains(logs, "first factor authentication failed")
	s.Contains(logs, "PAM: Authentication failure for john")
	s.NotContains(logs, "Accepted keyboard-interactive/pam for john")
}

func (s *PAMSuite) TestShouldAuthenticateWith1FA2FAUsingTOTP() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s.setPAMAuthLevel("1FA+2FA")

	code, err := totp.GenerateCode(pamTOTPSecret, time.Now())
	s.Require().NoError(err)

	since := time.Now()

	output, err := s.sshLogin(ctx, "john", []string{"password", code})
	s.Require().NoError(err, "1FA+2FA SSH login should succeed: %s", output)

	logs := s.pamLogsSince(since)
	s.Contains(logs, "POST https://login.example.com:8080/api/firstfactor")
	s.Contains(logs, "user info method=")
	s.Contains(logs, `method="totp"`)
	s.Contains(logs, "has_totp=true")
	s.Contains(logs, "POST https://login.example.com:8080/api/secondfactor/totp")
	s.Contains(logs, "Accepted keyboard-interactive/pam for john")
}

func (s *PAMSuite) TestShouldReject1FA2FAWithInvalidTOTP() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s.setPAMAuthLevel("1FA+2FA")

	since := time.Now()

	output, err := s.sshLogin(ctx, "john", []string{"password", "000000"})
	s.Error(err, "1FA+2FA with invalid TOTP should fail: %s", output)

	logs := s.pamLogsSince(since)
	s.Contains(logs, "POST https://login.example.com:8080/api/secondfactor/totp")
	s.Contains(logs, "response status=403")
	s.Contains(logs, "TOTP authentication failed")
	s.Contains(logs, "PAM: Authentication failure for john")
	s.NotContains(logs, "Accepted keyboard-interactive/pam for john")
}

func (s *PAMSuite) TestShouldAuthenticateWith2FAOnly() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s.setPAMAuthLevel("2FA")

	code, err := totp.GenerateCode(pamTOTPSecret, time.Now())
	s.Require().NoError(err)

	since := time.Now()

	output, err := s.sshLogin(ctx, "john", []string{"password", code})
	s.Require().NoError(err, "2FA-only SSH login should succeed: %s", output)

	logs := s.pamLogsSince(since)
	s.Contains(logs, "POST https://login.example.com:8080/api/firstfactor")
	s.Contains(logs, "POST https://login.example.com:8080/api/secondfactor/totp")
	s.Contains(logs, "Accepted keyboard-interactive/pam for john")
}

func (s *PAMSuite) TestShouldInitiateDeviceAuthorizationFlow() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s.setPAMAuthLevel("device-auth")

	since := time.Now()

	var qrSeen bool

	cfg := &ssh.ClientConfig{
		User: "john",
		Auth: []ssh.AuthMethod{
			ssh.KeyboardInteractive(func(_, _ string, questions []string, _ []bool) ([]string, error) {
				answers := make([]string, len(questions))

				for i, q := range questions {
					if strings.Contains(q, "Scan the QR code") {
						qrSeen = true
					}

					answers[i] = ""
				}

				return answers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // Test environment only.
		Timeout:         contextDialTimeout(ctx, 10*time.Second),
	}

	client, err := dialSSHContext(ctx, "ssh.example.com:22", cfg)
	if client != nil {
		client.Close()
	}

	s.Require().Error(err, "device-auth SSH dial should fail when nobody approves the flow")
	s.Require().True(qrSeen, "did not observe QR code prompt from device flow")

	logs := s.pamLogsSince(since)

	s.Require().Contains(logs, "POST https://login.example.com:8080/api/oidc/device-authorization")
	s.Require().Contains(logs, "device authorization response status=200")

	s.Require().Contains(logs, "POST https://login.example.com:8080/api/oidc/token")
	s.Require().Contains(logs, "authorization_pending")

	s.NotContains(logs, "invalid_client")
	s.NotContains(logs, "device token error")
	s.NotContains(logs, "device authorization denied")
	s.NotContains(logs, "device authorization token expired")

	s.NotContains(logs, "/api/firstfactor")
}

func (s *PAMSuite) TestShouldAuthenticateWithDeviceAuthorizationMatchingUser() {
	s.Page = s.doCreateTab(s.T(), GetLoginBaseURL(BaseDomain))
	s.verifyIsFirstFactorPage(s.T(), s.Page)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		s.MustClose()
	}()

	s.setPAMAuthLevel("device-auth-bind")

	since := time.Now()

	err := s.doDeviceAuthSSHLogin(ctx, "john", func(verificationURL string) {
		s.driveDeviceAuthConsent(ctx, s.T(), verificationURL, "john", "password")
	})
	s.Require().NoError(err, "device-auth SSH with matching user should succeed")

	logs := s.pamLogsSince(since)

	s.Contains(logs, "POST https://login.example.com:8080/api/oidc/device-authorization")
	s.Contains(logs, "device authorization response status=200")
	s.Contains(logs, "POST https://login.example.com:8080/api/oidc/token")
	s.Contains(logs, `device identity verified: claim "authelia.pam.username" == pam username "john"`)
	s.Contains(logs, "Accepted keyboard-interactive/pam for john")
	s.NotContains(logs, "does not match pam username")
}

func (s *PAMSuite) TestShouldRejectDeviceAuthorizationWithMismatchedUser() {
	s.Page = s.doCreateTab(s.T(), GetLoginBaseURL(BaseDomain))
	s.verifyIsFirstFactorPage(s.T(), s.Page)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		s.MustClose()
	}()

	s.setPAMAuthLevel("device-auth-bind")

	since := time.Now()

	err := s.doDeviceAuthSSHLogin(ctx, "john", func(verificationURL string) {
		s.driveDeviceAuthConsent(ctx, s.T(), verificationURL, "jane", "password")
	})
	s.Require().Error(err, "device-auth SSH with mismatched approver must be rejected")

	logs := s.pamLogsSince(since)

	s.Contains(logs, "POST https://login.example.com:8080/api/oidc/device-authorization")
	s.Contains(logs, "POST https://login.example.com:8080/api/oidc/token")

	s.Contains(logs, `authelia identity "jane" does not match pam username "john"`)
	s.Contains(logs, "PAM: Authentication failure for john")
	s.NotContains(logs, "Accepted keyboard-interactive/pam for john")
	s.NotContains(logs, `device identity verified: claim "authelia.pam.username" == pam username "john"`)
}

func (s *PAMSuite) seedUserTOTP(username string) {
	output, err := pamDockerEnvironment.Exec("authelia-backend", []string{
		"authelia", "storage", "user", "totp", "generate", username,
		"--force",
		"--secret", pamTOTPSecret,
		"--config=/config/configuration.yml",
	})
	s.Require().NoError(err, "failed to seed TOTP for %s: %s", username, output)

	ctx := context.Background()
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	s.Require().NoError(provider.SavePreferred2FAMethod(ctx, username, "totp"))

	s.SetOneTimePassword(username, RodSuiteCredentialOneTimePassword{
		Secret: pamTOTPSecret,
		ValidationOptions: totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	})
}

func (s *PAMSuite) setPAMAuthLevel(authLevel string) {
	var source string

	switch authLevel {
	case "1FA":
		source = "authelia-1fa"
	case "2FA":
		source = "authelia-2fa"
	case "1FA+2FA":
		source = "authelia-1fa2fa"
	case "device-auth":
		source = "authelia-device-auth"
	case "device-auth-bind":
		source = "authelia-device-auth-bind"
	default:
		s.T().Fatalf("unknown auth level: %s", authLevel)
	}

	out, err := pamDockerEnvironment.Exec("pam", []string{"cp", "/etc/pam.d/" + source, "/etc/pam.d/sshd"})
	s.Require().NoError(err, "failed to set PAM auth-level=%s: %s", authLevel, out)
}

func (s *PAMSuite) pamLogsSince(since time.Time) string {
	logs, err := pamDockerEnvironment.Logs("pam", []string{"--since", since.UTC().Format(time.RFC3339Nano)})
	s.Require().NoError(err, "failed to fetch pam container logs")

	return logs
}

//nolint:unparam // user is part of the helper's interface, reserved for future tests.
func (s *PAMSuite) sshLogin(ctx context.Context, user string, responses []string) (string, error) {
	next := 0

	cfg := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.KeyboardInteractive(func(_, _ string, questions []string, _ []bool) ([]string, error) {
				answers := make([]string, len(questions))

				for i := range questions {
					if next >= len(responses) {
						return nil, errors.New("ran out of responses for prompts")
					}

					answers[i] = responses[next]
					next++
				}

				return answers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // Test environment only.
		Timeout:         contextDialTimeout(ctx, 10*time.Second),
	}

	client, err := dialSSHContext(ctx, "ssh.example.com:22", cfg)
	if err != nil {
		return err.Error(), err
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err.Error(), err
	}

	defer session.Close()

	out, err := session.CombinedOutput(`echo "AUTHELIA_PAM_LOGIN_SUCCESS"`)

	return string(out), err
}

func contextDialTimeout(ctx context.Context, fallback time.Duration) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return fallback
	}

	return time.Until(deadline)
}

func dialSSHContext(ctx context.Context, addr string, cfg *ssh.ClientConfig) (*ssh.Client, error) {
	d := &net.Dialer{Timeout: cfg.Timeout}

	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, cfg)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return ssh.NewClient(c, chans, reqs), nil
}

func extractVerificationURL(prompt string) string {
	for _, line := range strings.Split(prompt, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "https://") {
			return line
		}
	}

	return ""
}

func (s *PAMSuite) driveDeviceAuthConsent(ctx context.Context, t *testing.T, verificationURL, username, password string) {
	s.doVisit(t, s.Context(ctx), verificationURL)
	s.verifyIsFirstFactorPage(t, s.Context(ctx))
	s.doFillLoginPageAndClick(t, s.Context(ctx), username, password, false)
	s.verifyIsSecondFactorPage(t, s.Context(ctx))
	s.doValidateTOTP(t, s.Context(ctx), username)
	s.verifyIsOpenIDConsentDecisionStage(t, s.Context(ctx))

	require := s.Require()
	require.NoError(s.WaitElementLocatedByID(t, s.Context(ctx), "openid-consent-accept").Click("left", 1))

	s.verifyBodyContains(t, s.Context(ctx), "Consent has been accepted and processed")
}

func (s *PAMSuite) doDeviceAuthSSHLogin(ctx context.Context, linuxUser string, approveFn func(verificationURL string)) error {
	var qrSeen bool

	cfg := &ssh.ClientConfig{
		User: linuxUser,
		Auth: []ssh.AuthMethod{
			ssh.KeyboardInteractive(func(_, _ string, questions []string, _ []bool) ([]string, error) {
				answers := make([]string, len(questions))

				for i, q := range questions {
					if !strings.Contains(q, "Scan the QR code") {
						continue
					}

					qrSeen = true

					verificationURL := extractVerificationURL(q)
					if verificationURL == "" {
						return nil, fmt.Errorf("could not extract verification URL from prompt")
					}

					approveFn(verificationURL)

					answers[i] = ""
				}

				return answers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // Test environment only.
		Timeout:         contextDialTimeout(ctx, 30*time.Second),
	}

	client, dialErr := dialSSHContext(ctx, "ssh.example.com:22", cfg)
	if client != nil {
		defer client.Close()
	}

	s.Require().True(qrSeen, "did not observe QR code prompt from device flow")

	if dialErr != nil {
		return dialErr
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}

	defer session.Close()

	_, err = session.CombinedOutput(`echo "AUTHELIA_PAM_LOGIN_SUCCESS"`)

	return err
}

func TestPAMSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewPAMSuite())
}
