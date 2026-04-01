package suites

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserManagementOpenLDAPScenario struct {
	UserManagementAPIScenario
}

func NewUserManagementOpenLDAPScenario() *UserManagementOpenLDAPScenario {
	return &UserManagementOpenLDAPScenario{}
}

func (s *UserManagementOpenLDAPScenario) SetupSuite() {
	s.UserManagementAPIScenario.SetupSuite()
}

func (s *UserManagementOpenLDAPScenario) SetupTest() {
	s.UserManagementAPIScenario.SetupTest()
}

func TestUserManagementOpenLDAPScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewUserManagementOpenLDAPScenario())
}

//nolint:unparam
func (s *UserManagementOpenLDAPScenario) login(username, password string) {
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
func (s *UserManagementOpenLDAPScenario) storeCookies(newCookies []*http.Cookie) {
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

func (s *UserManagementOpenLDAPScenario) apiRequest(method, path string, body interface{}) (*http.Response, []byte) {
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

func (s *UserManagementOpenLDAPScenario) Test_NewUserPOST_OpenLDAP_ShouldCreateNewUser() {
	s.login(adminUsername, adminPassword)

	username := "testuser-create"

	_, _ = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)

	newUser := map[string]interface{}{
		"username":    username,
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"testuser@example.com"},
		"groups":      []string{"dev"},
		"password":    "password",
	}

	res, body := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user: %s", string(body)))

	s.apiRequest("DELETE", "/api/admin/users/testuser-create", nil)
}

func (s *UserManagementOpenLDAPScenario) Test_ChangeUserPATCH_OpenLDAP_ShouldUpdateUser() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"username":     username,
		"given_name":   "Bob",
		"family_name":  "Dylan",
		"display_name": "Updated Bob Dylan",
		"mail":         []string{"updated@example.com"},
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=given_name,family_name,display_name,mail", username), updateData)

	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to update user: %s", string(body)))
}

func (s *UserManagementOpenLDAPScenario) Test_DeleteUserDELETE_OpenLDAP_ShouldDeleteUser() {
	s.login(adminUsername, adminPassword)

	username := testUserUsername

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)

	newUser := map[string]interface{}{
		"username":    username,
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"testuser@example.com"},
		"groups":      []string{"dev"},
		"password":    "password",
	}

	res, body := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user: %s", string(body)))

	res, body = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)

	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to delete user: %s", string(body)))

	res, _ = s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", username), nil)
	s.Assert().Equal(http.StatusNotFound, res.StatusCode)
}

func (s *UserManagementOpenLDAPScenario) Test_NewUserPOST_OpenLDAP_ShouldFailWithMissingUsername() {
	s.login(adminUsername, adminPassword)

	newUser := map[string]interface{}{
		"username":    "",
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"testuser@example.com"},
		"groups":      []string{"dev"},
		"password":    "password",
	}

	res, _ := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *UserManagementOpenLDAPScenario) Test_NewUserPOST_OpenLDAP_ShouldFailWithMissingSurname() {
	s.login(adminUsername, adminPassword)

	newUser := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "test",
		"family_name": "",
		"mail":        []string{"testuser@example.com"},
		"groups":      []string{"dev"},
		"password":    "password",
	}

	res, _ := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}
