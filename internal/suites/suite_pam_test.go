package suites

import (
	"context"
	"errors"
	"fmt"
	"log"
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

	*DockerEnvironment
}

func NewPAMSuite() *PAMSuite {
	return &PAMSuite{
		RodSuite: NewRodSuite(pamSuiteName),
	}
}

func (s *PAMSuite) SetupSuite() {
	s.DockerEnvironment = NewDockerEnvironment([]string{
		"internal/suites/compose.yml",
		"internal/suites/PAM/compose.yml",
		"internal/suites/example/compose/authelia/compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/portal/compose.yml",
		"internal/suites/example/compose/pam/compose.yml",
	})

	s.seedUserTOTP("john")
	s.seedUserTOTP("jane")

	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *PAMSuite) TearDownSuite() {
	if s.RodSession == nil {
		return
	}

	if err := s.RodSession.Stop(); err != nil {
		log.Fatal(err)
	}
}

// seedUserTOTP provisions TOTP for a user via the authelia CLI (which writes the
// encrypted secret to Authelia's storage) and simultaneously records the same
// secret in the suite's in-memory credential store so the Rod helpers can
// generate matching TOTP codes at validation time. Both users also get their
// preferred 2FA method set to totp so LoadUserInfo's subquery-driven fields
// report correctly.
func (s *PAMSuite) seedUserTOTP(username string) {
	output, err := s.Exec("authelia-backend", []string{
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

// setPAMAuthLevel switches /etc/pam.d/sshd in the pam container to one of the pre-seeded
// PAM configs. sshd re-reads the PAM config on each login, so no restart is required.
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

	out, err := s.Exec("pam", []string{"cp", "/etc/pam.d/" + source, "/etc/pam.d/sshd"})
	s.Require().NoError(err, "failed to set PAM auth-level=%s: %s", authLevel, out)
}

// pamLogsSince returns the pam container log output emitted since the given timestamp.
// Used by tests to assert that specific pam_authelia debug lines were produced for the
// attempted authentication, giving richer validation than just checking the ssh exit code.
func (s *PAMSuite) pamLogsSince(since time.Time) string {
	logs, err := s.Logs("pam", []string{"--since", since.UTC().Format(time.RFC3339Nano)})
	s.Require().NoError(err, "failed to fetch pam container logs")

	return logs
}

// sshLogin connects to the pam container's sshd and replies to each keyboard-interactive
// prompt with the corresponding entry from responses. It returns an error if the server
// rejects the authentication or if there are more prompts than responses.
//
//nolint:unparam // user is part of the helper's interface, reserved for future tests.
func (s *PAMSuite) sshLogin(user string, responses []string) (string, error) {
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
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", "ssh.example.com:22", cfg)
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

// extractVerificationURL scans a multi-line PROMPT_MULTI_VISIBLE payload for the first
// https:// URL. That line is the verification_uri_complete emitted by pam_authelia's
// performDeviceAuth (see cmd/pam_authelia/main.go in the pam repo) — the helper builds
// a prompt body of the form "Scan the QR code below or visit the URL to approve.\n
// <verification_uri_complete>\n\n<ASCII QR art>\n\nApprove on your device, then press
// Enter." so the first https:// line is always the verification URL the browser drive
// needs to hit.
func extractVerificationURL(prompt string) string {
	for _, line := range strings.Split(prompt, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "https://") {
			return line
		}
	}

	return ""
}

// driveDeviceAuthConsent opens a fresh browser tab against the Device Authorization
// verification URL, logs in to Authelia as the given user with password + TOTP, and
// accepts the OpenID Connect consent prompt. This is the browser half of the device
// flow — pam_authelia is parked waiting for the user to press Enter in the SSH session
// while this runs, then resumes polling the token endpoint once control returns.
func (s *PAMSuite) driveDeviceAuthConsent(t *testing.T, verificationURL, username, password string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	page := s.doCreateTab(t, verificationURL)

	defer func() {
		_ = page.Close()
	}()

	s.verifyIsFirstFactorPage(t, page.Context(ctx))
	s.doFillLoginPageAndClick(t, page.Context(ctx), username, password, false)
	s.verifyIsSecondFactorPage(t, page.Context(ctx))
	s.doValidateTOTP(t, page.Context(ctx), username)
	s.verifyIsOpenIDConsentDecisionStage(t, page.Context(ctx))

	require := s.Require()
	require.NoError(s.WaitElementLocatedByID(t, page.Context(ctx), "openid-consent-accept").Click("left", 1))

	s.verifyBodyContains(t, page.Context(ctx), "Consent has been accepted and processed")
}

// doDeviceAuthSSHLogin runs the SSH half of a Device Authorization flow: dials the
// pam container's sshd, waits for the QR-code keyboard-interactive prompt, parses the
// verification URL out of it, invokes approveFn to complete the browser-side consent
// (which blocks until Authelia records the grant), and then returns an empty response
// to unblock pam_authelia's polling. Returns the ssh.Dial error (nil on success, a
// PAM_AUTH_ERR-derived handshake failure on identity mismatch).
func (s *PAMSuite) doDeviceAuthSSHLogin(linuxUser string, approveFn func(verificationURL string)) error {
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
		Timeout:         30 * time.Second,
	}

	client, dialErr := ssh.Dial("tcp", "ssh.example.com:22", cfg)
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

func (s *PAMSuite) TestShouldAuthenticateWith1FA() {
	s.setPAMAuthLevel("1FA")

	since := time.Now()

	output, err := s.sshLogin("john", []string{"password"})
	s.Require().NoError(err, "1FA SSH login should succeed: %s", output)

	logs := s.pamLogsSince(since)
	s.Contains(logs, "POST https://login.example.com:8080/api/firstfactor")
	s.Contains(logs, `response status=200 status_field="OK"`)
	s.NotContains(logs, "/api/secondfactor/")
	s.Contains(logs, "Accepted keyboard-interactive/pam for john")
}

func (s *PAMSuite) TestShouldReject1FAWithBadPassword() {
	s.setPAMAuthLevel("1FA")

	since := time.Now()

	output, err := s.sshLogin("john", []string{"wrongpassword"})
	s.Error(err, "1FA SSH with bad password should fail: %s", output)

	logs := s.pamLogsSince(since)
	s.Contains(logs, "POST https://login.example.com:8080/api/firstfactor")
	s.Contains(logs, "response status=401")
	s.Contains(logs, "first factor authentication failed")
	s.Contains(logs, "PAM: Authentication failure for john")
	s.NotContains(logs, "Accepted keyboard-interactive/pam for john")
}

func (s *PAMSuite) TestShouldAuthenticateWith1FA2FAUsingTOTP() {
	s.setPAMAuthLevel("1FA+2FA")

	code, err := totp.GenerateCode(pamTOTPSecret, time.Now())
	s.Require().NoError(err)

	since := time.Now()

	output, err := s.sshLogin("john", []string{"password", code})
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
	s.setPAMAuthLevel("1FA+2FA")

	since := time.Now()

	output, err := s.sshLogin("john", []string{"password", "000000"})
	s.Error(err, "1FA+2FA with invalid TOTP should fail: %s", output)

	logs := s.pamLogsSince(since)
	s.Contains(logs, "POST https://login.example.com:8080/api/secondfactor/totp")
	s.Contains(logs, "response status=403")
	s.Contains(logs, "TOTP authentication failed")
	s.Contains(logs, "PAM: Authentication failure for john")
	s.NotContains(logs, "Accepted keyboard-interactive/pam for john")
}

// TestShouldAuthenticateWith2FAOnly tests the 2FA-only mode where pam_unix validates
// the password first, then pam_authelia picks up the password from the PAM stack via
// PAM_AUTHTOK, performs silent 1FA against Authelia, then prompts for TOTP.
func (s *PAMSuite) TestShouldAuthenticateWith2FAOnly() {
	s.setPAMAuthLevel("2FA")

	code, err := totp.GenerateCode(pamTOTPSecret, time.Now())
	s.Require().NoError(err)

	since := time.Now()

	output, err := s.sshLogin("john", []string{"password", code})
	s.Require().NoError(err, "2FA-only SSH login should succeed: %s", output)

	logs := s.pamLogsSince(since)
	s.Contains(logs, "POST https://login.example.com:8080/api/firstfactor")
	s.Contains(logs, "POST https://login.example.com:8080/api/secondfactor/totp")
	s.Contains(logs, "Accepted keyboard-interactive/pam for john")
}

// TestShouldInitiateDeviceAuthorizationFlow verifies that when method-priority is set
// to device_authorization, pam_authelia initiates the OAuth2 Device Authorization Grant
// flow against Authelia, begins polling, and enters the authorization_pending state.
//
// The test does NOT complete the browser-side consent step — that is covered by the
// identity-binding tests below. This one exists to regress the shim's POLLRDHUP / timeout
// cleanup path: the device-auth PAM config sets timeout=3 so the shim gives up reading
// from the Go helper after 3 seconds and returns PAM_AUTH_ERR, synchronizing the test on
// the real flow rather than an arbitrary sleep.
func (s *PAMSuite) TestShouldInitiateDeviceAuthorizationFlow() {
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
		Timeout:         10 * time.Second,
	}

	// Dial blocks until the C shim's 3-second deadline fires, pam_authelia returns
	// PAM_AUTH_ERR, and sshd terminates the session — synchronizing the test on the
	// real flow rather than an arbitrary sleep. An unexpected success would mean the
	// device flow authenticated without anyone approving, which must fail the test.
	client, err := ssh.Dial("tcp", "ssh.example.com:22", cfg)
	if client != nil {
		client.Close()
	}

	s.Require().Error(err, "device-auth SSH dial should fail when nobody approves the flow")
	s.Require().True(qrSeen, "did not observe QR code prompt from device flow")

	logs := s.pamLogsSince(since)

	// Flow initiation: device authorization endpoint was hit and returned 200. The
	// response body is no longer logged, so we can't assert on the RFC 8628 fields
	// directly — the fact that polling starts below implicitly proves they parsed OK.
	s.Require().Contains(logs, "POST https://login.example.com:8080/api/oidc/device-authorization")
	s.Require().Contains(logs, "device authorization response status=200")

	// Polling reached the healthy authorization_pending state — the only way this line
	// appears is if the token endpoint was reached AND the OAuth2 client authenticated
	// successfully. An invalid_client / invalid config would never get past the first poll.
	s.Require().Contains(logs, "POST https://login.example.com:8080/api/oidc/token")
	s.Require().Contains(logs, "authorization_pending")

	// Forbidden conditions — any of these indicate a broken device flow setup.
	s.NotContains(logs, "invalid_client")
	s.NotContains(logs, "device token error")
	s.NotContains(logs, "device authorization denied")
	s.NotContains(logs, "device authorization token expired")

	// Device flow short-circuits 1FA entirely; the password endpoint should never be hit.
	s.NotContains(logs, "/api/firstfactor")
}

// TestShouldAuthenticateWithDeviceAuthorizationMatchingUser exercises the happy path
// of the Device Authorization flow end to end: john initiates via SSH, john approves
// in the browser. pam_authelia's VerifyDeviceIdentity then matches the issued token's
// `authelia.pam.username` claim against the Linux username and returns PAM_SUCCESS.
func (s *PAMSuite) TestShouldAuthenticateWithDeviceAuthorizationMatchingUser() {
	s.setPAMAuthLevel("device-auth-bind")

	since := time.Now()

	err := s.doDeviceAuthSSHLogin("john", func(verificationURL string) {
		s.driveDeviceAuthConsent(s.T(), verificationURL, "john", "password")
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

// TestShouldRejectDeviceAuthorizationWithMismatchedUser exercises the confused-deputy
// defense: john initiates the Device Authorization flow via SSH but jane approves it in
// the browser. pam_authelia's VerifyDeviceIdentity then observes userinfo.authelia.pam
// .username == "jane" != "john" and refuses to return PAM_SUCCESS, so sshd rejects the
// login even though Authelia happily issued a valid token to jane.
//
// Without this check any Authelia account holder could approve another user's QR code
// and end up logged in as them. This test locks that behavior in.
func (s *PAMSuite) TestShouldRejectDeviceAuthorizationWithMismatchedUser() {
	s.setPAMAuthLevel("device-auth-bind")

	since := time.Now()

	err := s.doDeviceAuthSSHLogin("john", func(verificationURL string) {
		s.driveDeviceAuthConsent(s.T(), verificationURL, "jane", "password")
	})
	s.Require().Error(err, "device-auth SSH with mismatched approver must be rejected")

	logs := s.pamLogsSince(since)

	// The flow reached the token endpoint and Authelia granted the token to jane ….
	s.Contains(logs, "POST https://login.example.com:8080/api/oidc/device-authorization")
	s.Contains(logs, "POST https://login.example.com:8080/api/oidc/token")

	// … but pam_authelia's VerifyDeviceIdentity rejected the mismatch.
	s.Contains(logs, `authelia identity "jane" does not match pam username "john"`)
	s.Contains(logs, "PAM: Authentication failure for john")
	s.NotContains(logs, "Accepted keyboard-interactive/pam for john")
	s.NotContains(logs, `device identity verified: claim "authelia.pam.username" == pam username "john"`)
}

func TestPAMSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewPAMSuite())
}
