package suites

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/authentication"
)

type UserManagementFileScenario struct {
	UserManagementAPIScenario
}

func NewUserManagementFileScenario() *UserManagementFileScenario {
	return &UserManagementFileScenario{}
}

func (s *UserManagementFileScenario) SetupSuite() {
	s.UserManagementAPIScenario.SetupSuite()
}

func (s *UserManagementFileScenario) SetupTest() {
	s.UserManagementAPIScenario.SetupTest()
}

func TestUserManagementFileScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewUserManagementFileScenario())
}

func (s *UserManagementFileScenario) login(username, password string) {
	loginURL := fmt.Sprintf("%s/api/firstfactor", AutheliaBaseURL)

	loginData := map[string]interface{}{
		"username":       username,
		"password":       password,
		"KeepMeLoggedIn": false,
	}

	body, err := json.Marshal(loginData)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	res, err := s.client.Do(req)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode, "Login failed")

	s.storeCookies(res.Cookies())

	s.Require().NotEmpty(s.cookies, "No cookies received")
}

// storeCookies updates the stored cookies, replacing existing ones with the same name.
func (s *UserManagementFileScenario) storeCookies(newCookies []*http.Cookie) {
	for _, newCookie := range newCookies {
		found := false

		for i, existingCookie := range s.cookies {
			if existingCookie.Name == newCookie.Name {
				s.cookies[i] = newCookie
				found = true

				break
			}
		}

		if !found {
			s.cookies = append(s.cookies, newCookie)
		}
	}
}

func (s *UserManagementFileScenario) apiRequest(method, path string, body interface{}) (*http.Response, []byte) {
	url := fmt.Sprintf("%s%s", AutheliaBaseURL, path)

	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		s.Require().NoError(err)

		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	s.Require().NoError(err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for _, cookie := range s.cookies {
		req.AddCookie(cookie)
	}

	res, err := s.client.Do(req)
	s.Require().NoError(err)

	s.storeCookies(res.Cookies())

	responseBody, err := io.ReadAll(res.Body)
	s.Require().NoError(err)
	res.Body.Close()

	return res, responseBody
}

func (s *UserManagementFileScenario) Test_NewUserPOST_ShouldErrorWhenCreatingUserWithNonexistentGroups() {
	s.T().Skip("File provider allows arbitrary group names")
}

func (s *UserManagementFileScenario) Test_ChangeUserPATCH_ShouldReplaceAllUserGroups() {
	s.login(adminUsername, adminPassword)

	res, _ := s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	newUser := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"test-user@example.com"},
		"groups":      []string{"dev", "admins"},
		"password":    testPassword,
	}

	res, body := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user: %s", string(body)))

	username := testUserUsername
	updateData := map[string]interface{}{
		"groups": []string{testGroupName},
	}

	res, body = s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=groups", username), updateData)

	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to update user: %s", string(body)))

	res, body = s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", username), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response struct {
		Status string                             `json:"status"`
		Data   authentication.UserDetailsExtended `json:"data"`
	}

	err := json.Unmarshal(body, &response)
	s.Assert().NoError(err)
	s.Assert().Equal([]string{testGroupName}, response.Data.Groups)

	res, _ = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func (s *UserManagementFileScenario) Test_ChangeUserPATCH_ShouldErrorWhenAddingNonexistentGroups() {
	s.T().Skip("File provider allows arbitrary group names")
}

func (s *UserManagementFileScenario) Test_NewUserPOST_File_ShouldCreateUserWithArbitraryGroups() {
	s.login(adminUsername, adminPassword)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)

	newUser := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"test-user@example.com"},
		"groups":      []string{"newgroup1", "newgroup2", "arbitrary-group"},
		"password":    "password",
	}

	res, body := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user with arbitrary groups: %s", string(body)))

	res, body = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to delete user: %s", string(body)))
}

func (s *UserManagementFileScenario) Test_ChangeUserPATCH_File_ShouldUpdateToArbitraryGroups() {
	s.login(adminUsername, adminPassword)

	res, _ := s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	newUser := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"test-user@example.com"},
		"groups":      []string{"group1"},
		"password":    testPassword,
	}

	res, body := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user: %s", string(body)))

	username := testUserUsername
	updateData := map[string]interface{}{
		"groups": []string{"arbitrary-group-1", "arbitrary-group-2", "new-group"},
	}

	res, body = s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=groups", username), updateData)

	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to update user with arbitrary groups: %s", string(body)))

	res, _ = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)
}
