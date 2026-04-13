package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	methodTOTP       = "totp"
	methodWebAuthn   = "webauthn"
	methodMobilePush = "mobile_push"
	methodDeviceAuth = "device_authorization"
	// methodUser resolves to the user's preferred Authelia method at runtime.
	methodUser = "user"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "authelia-pam: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := ParseConfig(os.Args[1:])
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	client, err := NewAutheliaClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)
	writer := os.Stdout

	username, err := ReadLine(reader)
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}

	if username == "" {
		return writeFailure(writer, "empty username")
	}

	password, err := ReadLine(reader)
	if err != nil {
		return writeFailure(writer, "failed to read password")
	}

	// Device flow is a self-contained authentication; if it's the first priority entry
	// skip 1FA and the user info lookup entirely.
	if len(cfg.MethodPriority) > 0 && cfg.MethodPriority[0] == methodDeviceAuth && cfg.OAuth2ClientID != "" {
		if err = performDeviceAuth(cfg, client, writer); err != nil {
			return writeFailure(writer, err.Error())
		}

		return WriteSuccess(writer)
	}

	if err = client.FirstFactor(username, password); err != nil {
		return writeFailure(writer, fmt.Sprintf("authentication failed: %v", err))
	}

	if cfg.AuthLevel == AuthLevel1FA {
		return WriteSuccess(writer)
	}

	userInfo, err := client.UserInfo()
	if err != nil {
		return writeFailure(writer, "failed to retrieve user information")
	}

	if err = performSecondFactor(cfg, client, userInfo, reader, writer); err != nil {
		return writeFailure(writer, err.Error())
	}

	return WriteSuccess(writer)
}

// pickSecondFactorMethod returns the first method in cfg.MethodPriority that is usable
// for the current user, defaulting to the user's Authelia preference when unset.
func pickSecondFactorMethod(cfg *Config, client *AutheliaClient, userInfo *UserInfoResponse) (string, error) {
	priority := cfg.MethodPriority
	if len(priority) == 0 {
		priority = []string{methodUser}
	}

	for _, m := range priority {
		resolved := resolveMethod(m, cfg, userInfo)
		if resolved != "" && methodUsable(resolved, cfg, userInfo) {
			client.debugf("selected %q (from priority entry %q)", resolved, m)

			return resolved, nil
		}

		client.debugf("method %q not usable for user, trying next", m)
	}

	return "", fmt.Errorf("no usable 2FA method for this user")
}

// resolveMethod maps a priority list entry to a concrete 2FA method. The special
// "user" entry resolves to the user's preferred Authelia method; webauthn falls back
// through TOTP, Duo, and device authorization since it cannot respond over SSH.
func resolveMethod(entry string, cfg *Config, userInfo *UserInfoResponse) string {
	if entry != methodUser {
		return entry
	}

	pref := userInfo.Method
	if pref == methodWebAuthn || pref == "" {
		switch {
		case userInfo.HasTOTP:
			return methodTOTP
		case userInfo.HasDuo:
			return methodMobilePush
		case cfg.OAuth2ClientID != "":
			return methodDeviceAuth
		default:
			return ""
		}
	}

	return pref
}

// methodUsable reports whether the given method can be used for the current user.
func methodUsable(method string, cfg *Config, userInfo *UserInfoResponse) bool {
	switch method {
	case methodTOTP:
		return userInfo.HasTOTP
	case methodMobilePush:
		return userInfo.HasDuo
	case methodDeviceAuth:
		return cfg.OAuth2ClientID != ""
	default:
		return false
	}
}

func performSecondFactor(cfg *Config, client *AutheliaClient, userInfo *UserInfoResponse, reader *bufio.Reader, writer *os.File) error {
	method, err := pickSecondFactorMethod(cfg, client, userInfo)
	if err != nil {
		return err
	}

	switch method {
	case methodTOTP:
		return performTOTP(client, reader, writer)
	case methodMobilePush:
		return performDuoPush(client, writer)
	case methodDeviceAuth:
		return performDeviceAuth(cfg, client, writer)
	default:
		return fmt.Errorf("unsupported 2FA method: %s", method)
	}
}

func performTOTP(client *AutheliaClient, reader *bufio.Reader, writer *os.File) error {
	if err := WritePromptVisible(writer, "TOTP Code: "); err != nil {
		return err
	}

	token, err := ReadLine(reader)
	if err != nil {
		return fmt.Errorf("failed to read TOTP code")
	}

	token = strings.TrimSpace(token)

	n := len(token)
	if n != 6 && n != 8 {
		return fmt.Errorf("TOTP code must be 6 or 8 digits")
	}

	return client.SecondFactorTOTP(token)
}

func performDeviceAuth(cfg *Config, client *AutheliaClient, writer *os.File) error {
	if cfg.OAuth2ClientID == "" {
		return fmt.Errorf("device authorization requires --oauth2-client-id")
	}

	resp, err := client.DeviceAuthorize(cfg.OAuth2ClientID, cfg.OAuth2ClientSecret, cfg.OAuth2Scope)
	if err != nil {
		return fmt.Errorf("failed to initiate device authorization: %w", err)
	}

	verification := resp.VerificationURIComplete
	if verification == "" {
		verification = resp.VerificationURI
	}

	lines := []string{
		"Scan the QR code below or visit the URL to approve, then enter code: " + resp.UserCode,
		verification,
	}

	if qrCode, err := renderQRCode(verification); err == nil {
		lines = append(lines, "")
		lines = append(lines, strings.Split(strings.TrimRight(qrCode, "\n"), "\n")...)
	}

	for _, line := range lines {
		if err := WriteInfo(writer, line); err != nil {
			return err
		}
	}

	return client.PollDeviceToken(cfg.OAuth2ClientID, cfg.OAuth2ClientSecret, resp.DeviceCode, resp.ExpiresIn, resp.Interval)
}

func performDuoPush(client *AutheliaClient, writer *os.File) error {
	if err := WriteInfo(writer, "Duo push sent. Approve on your device..."); err != nil {
		return err
	}

	return client.SecondFactorDuoPush()
}

func writeFailure(writer *os.File, msg string) error {
	fmt.Fprintf(os.Stderr, "authelia-pam: %s\n", msg)
	_ = WriteFailure(writer, msg)

	return fmt.Errorf("%s", msg)
}
