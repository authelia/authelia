package suites

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/authelia/otp/totp"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ssh"

	"github.com/authelia/authelia/v4/internal/storage"
)

type PAMSuite struct {
	*CommandSuite
}

func NewPAMSuite() *PAMSuite {
	return &PAMSuite{
		CommandSuite: &CommandSuite{
			BaseSuite: &BaseSuite{
				Name: pamSuiteName,
			},
		},
	}
}

func (s *PAMSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/compose.yml",
		"internal/suites/PAM/compose.yml",
		"internal/suites/example/compose/authelia/compose.backend.{}.yml",
		"internal/suites/example/compose/nginx/portal/compose.yml",
		"internal/suites/example/compose/pam/compose.yml",
	})
	s.DockerEnvironment = dockerEnvironment

	output, err := s.Exec("authelia-backend", []string{
		"authelia", "storage", "user", "totp", "generate", "john",
		"--force",
		"--secret", pamTOTPSecret,
		"--config=/config/configuration.yml",
	})
	s.Require().NoError(err, "failed to seed TOTP for john: %s", output)

	ctx := context.Background()
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	s.Require().NoError(provider.SavePreferred2FAMethod(ctx, "john", "totp"))
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
// The test does NOT complete the browser-side consent step — that end-to-end flow is
// covered by Authelia's OIDC suite. Our scope is "the PAM module drives the RFC 8628
// client correctly and reaches the polling state".
//
// Flow: the QR code arrives as a PROMPT_MULTI_VISIBLE keyboard-interactive question.
// We answer with an empty response (as if the user pressed Enter after approving) so
// the Go binary starts polling the token endpoint. Nobody ever approves, so the poll
// sits in authorization_pending; the C shim's configured timeout=3 then fires and
// kills the Go process, returning PAM_AUTH_ERR to sshd which terminates the session.
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

	_, _ = ssh.Dial("tcp", "ssh.example.com:22", cfg)

	// Give the C shim time to fire its 3-second deadline, kill the Go binary, and
	// return PAM_AUTH_ERR to sshd so any polling stops before the next test runs.
	time.Sleep(4 * time.Second)

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

func TestPAMSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewPAMSuite())
}
