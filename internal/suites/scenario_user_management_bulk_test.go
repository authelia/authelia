package suites

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// bulkTestUserCount is the number of users inserted/removed by the bulk load tests below.
const bulkTestUserCount = 1000

// bulkTestUsernamePrefix identifies users created by TestUserManagementBulkInsert so
// TestUserManagementBulkRemove can find and delete them again.
const bulkTestUsernamePrefix = "loadtestuser"

// TestUserManagementBulkInsert is a manual, opt-in test that inserts ~1000 users via the
// admin API and leaves them in place (e.g. to exercise pagination/filtering in the UI
// against a large dataset). It is skipped unless AUTHELIA_BULK_TEST_INSERT=1 is set.
//
// Run against a running Standalone suite with:
//
//	AUTHELIA_BULK_TEST_INSERT=1 go test ./internal/suites/... -run TestUserManagementBulkInsert -v
func TestUserManagementBulkInsert(t *testing.T) {
	if os.Getenv("AUTHELIA_BULK_TEST_INSERT") != "1" {
		t.Skip("skipping bulk user insert: set AUTHELIA_BULK_TEST_INSERT=1 to run")
	}

	client := NewHTTPClient()
	cookies := bulkLogin(t, client, adminUsername, adminPassword)

	for i := 0; i < bulkTestUserCount; i++ {
		username := fmt.Sprintf("%s%04d", bulkTestUsernamePrefix, i)

		newUser := map[string]interface{}{
			"username":    username,
			"given_name":  "Load",
			"family_name": fmt.Sprintf("Test %04d", i),
			"mail":        []string{fmt.Sprintf("%s@example.com", username)},
			"groups":      []string{},
			"password":    "password",
		}

		res, body := bulkAPIRequest(t, client, cookies, "POST", "/api/admin/users", newUser)
		require.Equal(t, http.StatusCreated, res.StatusCode, "failed to create user %s: %s", username, string(body))
	}

	t.Logf("created %d users with prefix %q", bulkTestUserCount, bulkTestUsernamePrefix)
}

// TestUserManagementBulkRemove is the counterpart to TestUserManagementBulkInsert: it
// deletes every user matching bulkTestUsernamePrefix. It is skipped unless
// AUTHELIA_BULK_TEST_REMOVE=1 is set.
//
// Run against a running Standalone suite with:
//
//	AUTHELIA_BULK_TEST_REMOVE=1 go test ./internal/suites/... -run TestUserManagementBulkRemove -v
func TestUserManagementBulkRemove(t *testing.T) {
	if os.Getenv("AUTHELIA_BULK_TEST_REMOVE") != "1" {
		t.Skip("skipping bulk user removal: set AUTHELIA_BULK_TEST_REMOVE=1 to run")
	}

	client := NewHTTPClient()
	cookies := bulkLogin(t, client, adminUsername, adminPassword)

	deleted := 0

	for i := 0; i < bulkTestUserCount; i++ {
		username := fmt.Sprintf("%s%04d", bulkTestUsernamePrefix, i)

		res, body := bulkAPIRequest(t, client, cookies, "DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)
		if res.StatusCode != http.StatusOK {
			t.Logf("failed to delete user %s: %s", username, string(body))

			continue
		}

		deleted++
	}

	t.Logf("deleted %d users with prefix %q", deleted, bulkTestUsernamePrefix)
}

func bulkLogin(t *testing.T, client *http.Client, username, password string) []*http.Cookie {
	t.Helper()

	loginURL := fmt.Sprintf("%s/api/firstfactor", AutheliaBaseURL)

	loginData := map[string]interface{}{
		"username":       username,
		"password":       password,
		"KeepMeLoggedIn": false,
	}

	body, err := json.Marshal(loginData)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	require.NoError(t, err)

	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "login failed")

	cookies := res.Cookies()
	require.NotEmpty(t, cookies, "no cookies received")

	return cookies
}

func bulkAPIRequest(t *testing.T, client *http.Client, cookies []*http.Cookie, method, path string, body interface{}) (*http.Response, []byte) {
	t.Helper()

	url := fmt.Sprintf("%s%s", AutheliaBaseURL, path)

	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)

		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	require.NoError(t, err)

	responseBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()

	return res, responseBody
}
