package handlers

import (
	"fmt"
	"net"
	"testing"

	"github.com/avct/uasurfer"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestRedactEmail(t *testing.T) {
	testCases := []struct {
		testName string
		input    string
		expected string
	}{
		{"ShouldRedactEmail", "james.dean@authelia.com", "j********n@authelia.com"},
		{"ShouldRedactShortEmail", "me@authelia.com", "**@authelia.com"},
		{"ShouldRedactInvalidEmail", "invalidEmail.com", ""},
		{"ShouldNotErrorOnEmptyInput", "", ""},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			require.Equal(t, tc.expected, redactEmail(tc.input))
		})
	}
}

func TestIsIPTrusted(t *testing.T) {
	testCases := []struct {
		testName            string
		ip                  string
		notifyPrivateRanges bool
		trustedNetworks     []string
		expected            bool
	}{
		{"ShouldReturnFalseForNilIP", "", false, []string{}, false},
		{"ShouldTrustPrivateIPWhenNotNotifying", "192.168.1.1", false, []string{}, true},
		{"ShouldTrustLoopbackIPWhenNotNotifying", "127.0.0.1", false, []string{}, true},
		{"ShouldNotTrustPrivateIPWhenNotifying", "192.168.1.1", true, []string{}, false},
		{"ShouldNotTrustLoopbackIPWhenNotifying", "127.0.0.1", true, []string{}, false},
		{"ShouldTrustIPInTrustedNetwork", "203.0.113.1", true, []string{"203.0.113.0/24"}, true},
		{"ShouldNotTrustIPNotInTrustedNetwork", "203.0.113.1", true, []string{"198.51.100.0/24"}, false},
		{"ShouldTrustIPInMultipleTrustedNetworks", "203.0.113.1", true, []string{"198.51.100.0/24", "203.0.113.0/24"}, true},
		{"ShouldNotTrustPublicIPWithNoTrustedNetworks", "203.0.113.1", true, []string{}, false},
		{"ShouldTrustIPv6PrivateWhenNotNotifying", "fd12:3456:789a:1::1", false, []string{}, true},
		{"ShouldNotTrustIPv6PrivateWhenNotifying", "fd12:3456:789a:1::1", true, []string{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Configuration.AuthenticationBackend.KnownIP.NotifyPrivateRanges = tc.notifyPrivateRanges
			mock.Ctx.Configuration.AuthenticationBackend.KnownIP.TrustedNetworks = parseTrustedNetworks(tc.trustedNetworks)

			var ip net.IP
			if tc.ip != "" {
				ip = net.ParseIP(tc.ip)
			}

			result := IsIPTrusted(mock.Ctx, ip)
			require.Equal(t, tc.expected, result)
		})
	}
}
func TestHandleKnownIPTracking(t *testing.T) {
	testCases := []struct {
		testName             string
		knownIPEnabled       bool
		remoteIP             string
		isPrivateIP          bool
		notifyPrivateRanges  bool
		ipExists             bool
		storageError         error
		updateError          error
		saveError            error
		expectedStorageCalls string
	}{
		{"ShouldSkipWhenDisabled", false, "203.0.113.1", false, true, false, nil, nil, nil, "none"},
		{"ShouldSkipWhenIPTrusted", true, "192.168.1.1", true, false, false, nil, nil, nil, "none"},
		{"ShouldUpdateExistingIP", true, "203.0.113.1", false, true, true, nil, nil, nil, "update"},
		{"ShouldHandleNewIP", true, "203.0.113.1", false, true, false, nil, nil, nil, "save"},
		{"ShouldHandleCheckError", true, "203.0.113.1", false, true, false, fmt.Errorf("check failed"), nil, nil, "check_error"},
		{"ShouldHandleUpdateError", true, "203.0.113.1", false, true, true, nil, fmt.Errorf("update failed"), nil, "update_error"},
		{"ShouldHandleSaveError", true, "203.0.113.1", false, true, false, nil, nil, fmt.Errorf("save failed"), "save_error"},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Configuration.AuthenticationBackend.KnownIP.Enable = tc.knownIPEnabled
			mock.Ctx.Configuration.AuthenticationBackend.KnownIP.NotifyPrivateRanges = tc.notifyPrivateRanges

			if tc.remoteIP != "" {
				mock.Ctx.SetRemoteAddr(&net.TCPAddr{
					IP:   net.ParseIP(tc.remoteIP),
					Port: 12345,
				})
			}

			userSession := &session.UserSession{
				Username: "testuser",
				Emails:   []string{"test@example.com"},
			}

			switch tc.expectedStorageCalls {
			case "update":
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						IsIPKnownForUser(mock.Ctx, "testuser", gomock.Any()).
						Return(tc.ipExists, tc.storageError),
					mock.StorageMock.EXPECT().
						UpdateKnownIP(mock.Ctx, "testuser", gomock.Any()).
						Return(tc.updateError),
				)
			case "save":
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						IsIPKnownForUser(mock.Ctx, "testuser", gomock.Any()).
						Return(tc.ipExists, tc.storageError),
					mock.StorageMock.EXPECT().
						SaveNewIPForUser(mock.Ctx, "testuser", gomock.Any(), gomock.Any()).
						Return(tc.saveError),
				)

				if tc.saveError == nil {
					mock.NotifierMock.EXPECT().
						Send(mock.Ctx, gomock.Any(), "Login From New IP", gomock.Any(), gomock.Any()).
						Return(nil)
				}
			case "check_error":
				mock.StorageMock.EXPECT().
					IsIPKnownForUser(mock.Ctx, "testuser", gomock.Any()).
					Return(tc.ipExists, tc.storageError)
			case "update_error":
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						IsIPKnownForUser(mock.Ctx, "testuser", gomock.Any()).
						Return(tc.ipExists, tc.storageError),
					mock.StorageMock.EXPECT().
						UpdateKnownIP(mock.Ctx, "testuser", gomock.Any()).
						Return(tc.updateError),
				)
			case "save_error":
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						IsIPKnownForUser(mock.Ctx, "testuser", gomock.Any()).
						Return(tc.ipExists, tc.storageError),
					mock.StorageMock.EXPECT().
						SaveNewIPForUser(mock.Ctx, "testuser", gomock.Any(), gomock.Any()).
						Return(tc.saveError),
				)
			}

			HandleKnownIPTracking(mock.Ctx, userSession)
		})
	}
}

func TestHandleNewIP(t *testing.T) {
	testCases := []struct {
		testName              string
		username              string
		userEmails            []string
		storageError          error
		notifierError         error
		expectedNotifierCalls bool
	}{
		{"ShouldHandleNewIPSuccessfully", "testuser", []string{"test@example.com"}, nil, nil, true},
		{"ShouldHandleStorageError", "testuser", []string{"test@example.com"}, fmt.Errorf("storage failed"), nil, false},
		{"ShouldHandleNoEmailAddress", "testuser", []string{}, nil, nil, false},
		{"ShouldHandleNotifierError", "testuser", []string{"test@example.com"}, nil, fmt.Errorf("notifier failed"), true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

			userSession := &session.UserSession{
				Username:    tc.username,
				DisplayName: "Test User",
				Emails:      tc.userEmails,
			}

			ip := model.NewIP(net.ParseIP("203.0.113.1"))

			mock.StorageMock.EXPECT().
				SaveNewIPForUser(mock.Ctx, tc.username, gomock.Any(), gomock.Any()).
				Return(tc.storageError)

			if tc.storageError == nil && len(tc.userEmails) > 0 {
				mock.NotifierMock.EXPECT().
					Send(mock.Ctx, gomock.Any(), "Login From New IP", gomock.Any(), gomock.Any()).
					Return(tc.notifierError)
			}

			handleNewIP(mock.Ctx, userSession, ip)
		})
	}
}

func TestSendNewIPEmail(t *testing.T) {
	testCases := []struct {
		testName          string
		userEmails        []string
		notifierError     error
		expectedSendCalls bool
	}{
		{"ShouldSendEmailSuccessfully", []string{"test@example.com"}, nil, true},
		{"ShouldHandleNoEmailAddress", []string{}, nil, false},
		{"ShouldHandleNotifierError", []string{"test@example.com"}, fmt.Errorf("notifier failed"), true},
		{"ShouldUseFirstEmailWhenMultiple", []string{"test@example.com", "test2@example.com"}, nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			userSession := &session.UserSession{
				Username:    "testuser",
				DisplayName: "Test User",
				Emails:      tc.userEmails,
			}

			ip := model.NewIP(net.ParseIP("203.0.113.1"))

			userAgent := &uasurfer.UserAgent{
				Browser: uasurfer.Browser{Name: uasurfer.BrowserChrome},
				OS:      uasurfer.OS{Name: uasurfer.OSWindows},
			}

			rawUserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

			if len(tc.userEmails) > 0 {
				mock.NotifierMock.EXPECT().
					Send(mock.Ctx, gomock.Any(), "Login From New IP", gomock.Any(), gomock.Any()).
					Return(tc.notifierError)
			}

			sendNewIPEmail(mock.Ctx, userSession, ip, userAgent, rawUserAgent)
		})
	}
}

func parseTrustedNetworks(networks []string) []*net.IPNet {
	var result []*net.IPNet
	for _, network := range networks {
		_, ip, err := net.ParseCIDR(network)
		if err == nil && ip != nil {
			result = append(result, ip)
		}
	}

	return result
}
