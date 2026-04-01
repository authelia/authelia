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

const adminUsername = "john"
const adminPassword = "password"

const nonAdminUsername = "bob"
const nonAdminPassword = "password"

const nonExistentUsername = "nonexistentuser"
const nonExistentGroupName = "nonexistentgroup"

const testUserUsername = "testuser"
const testUserUsername2 = "testuser2"

const testGroupName = "testgroup"
const testGroupName2 = "testgroup2"

type UserManagementAPIScenario struct {
	suite.Suite

	client  *http.Client
	cookies []*http.Cookie
}

func NewUserManagementAPIScenario() *UserManagementAPIScenario {
	return &UserManagementAPIScenario{}
}

func (s *UserManagementAPIScenario) SetupSuite() {
	s.client = NewHTTPClient()
	s.logout()
}

func (s *UserManagementAPIScenario) SetupTest() {
	s.logout()
	s.cleanupTestFixtures()
}

func (s *UserManagementAPIScenario) TearDownTest() {
	s.logout()
	s.cleanupTestFixtures()
}

func (s *UserManagementAPIScenario) logout() {
	s.cookies = make([]*http.Cookie, 0)
}

func TestUserManagementScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewUserManagementAPIScenario())
}

func (s *UserManagementAPIScenario) cleanupTestFixtures() {
	s.login(adminUsername, adminPassword)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername2), nil)
	s.logout()
}

//nolint:unparam
func (s *UserManagementAPIScenario) login(username, password string) {
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
func (s *UserManagementAPIScenario) storeCookies(newCookies []*http.Cookie) {
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

func (s *UserManagementAPIScenario) apiRequest(method, path string, body interface{}) (*http.Response, []byte) {
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

func (s *UserManagementAPIScenario) Test_AllUsersInfoGET_ShouldReturnUnauthorizedForAnonymous() {
	res, _ := s.apiRequest("GET", "/api/admin/users", nil)
	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_AllUsersInfoGET_ShouldReturnSuccessForAdmin() {
	s.login(adminUsername, adminPassword)

	res, body := s.apiRequest("GET", "/api/admin/users", nil)

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response struct {
		Status string                               `json:"status"`
		Data   []authentication.UserDetailsExtended `json:"data"`
	}

	err := json.Unmarshal(body, &response)
	s.Assert().NoError(err, "Response should match UserDetailsExtended structure")

	s.Assert().NotEmpty(response.Data, "Users list should not be empty")
}

func (s *UserManagementAPIScenario) Test_GetUserGET_ShouldReturnUser() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	res, body := s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", username), nil)

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response struct {
		Status string                             `json:"status"`
		Data   authentication.UserDetailsExtended `json:"data"`
	}

	err := json.Unmarshal(body, &response)
	s.Assert().NoError(err)

	s.Assert().Equal("OK", response.Status)
	s.Assert().Equal(username, response.Data.Username)
}

func (s *UserManagementAPIScenario) Test_GetUserGET_ShouldReturnUserGroups() {
	s.login(adminUsername, adminPassword)

	res, body := s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", nonAdminUsername), nil)

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response struct {
		Status string                             `json:"status"`
		Data   authentication.UserDetailsExtended `json:"data"`
	}

	err := json.Unmarshal(body, &response)
	s.Assert().NoError(err)

	s.Assert().Equal("OK", response.Status)
	s.Assert().Equal([]string{"dev"}, response.Data.Groups)
}

func (s *UserManagementAPIScenario) Test_AllUsersInfoGET_ShouldReturnForbiddenForNonAdmin() {
	s.login(nonAdminUsername, nonAdminPassword)

	res, _ := s.apiRequest("GET", "/api/admin/users", nil)
	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_GetUserGET_ShouldReturnForbiddenForNonAdmin() {
	s.login(nonAdminUsername, nonAdminPassword)

	username := nonAdminUsername
	res, _ := s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", username), nil)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldReturnForbiddenForNonAdmin() {
	s.login(nonAdminUsername, nonAdminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"username":     username,
		"given_name":   "Bob",
		"family_name":  "Dylan",
		"display_name": "Updated Bob Dylan",
		"mail":         []string{"updated@example.com"},
	}

	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=given_name,family_name,display_name,emails", username), updateData)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_DeleteUserDELETE_ShouldReturnForbiddenForNonAdmin() {
	s.login(nonAdminUsername, nonAdminPassword)

	username := nonAdminUsername

	res, _ := s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_NewUserPOST_ShouldCreateUserWithNoGroups() {
	s.login(adminUsername, adminPassword)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)

	newUser := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"test-user@example.com"},
		"groups":      []string{},
		"password":    "password",
	}

	res, body := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user: %s", string(body)))

	res, body = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to delete user: %s", string(body)))
}

func (s *UserManagementAPIScenario) Test_NewUserPOST_ShouldCreateUserWithMultipleGroups() {
	s.login(adminUsername, adminPassword)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)

	newUser := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"test-user@example.com"},
		"groups":      []string{"dev", "admins"},
		"password":    "password",
	}

	res, body := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user: %s", string(body)))

	res, body = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to delete user: %s", string(body)))
}

func (s *UserManagementAPIScenario) Test_NewUserPOST_ShouldErrorWhenCreatingUserWithNonexistentGroups() {
	s.login(adminUsername, adminPassword)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)

	newUser := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"test-user@example.com"},
		"groups":      []string{"dev", nonExistentGroupName},
		"password":    "password",
	}

	res, _ := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	res, body := s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to delete user: %s", string(body)))
}

func (s *UserManagementAPIScenario) Test_NewUserPOST_ShouldFailWithInvalidEmail() {
	s.login(adminUsername, adminPassword)

	newUser := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"invalid.example.com"},
		"groups":      []string{"dev"},
		"password":    "password",
	}

	res, _ := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_NewUserPOST_ShouldCreateUserWithExtraAttributes() {
	s.login(adminUsername, adminPassword)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)

	updateData := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"test@example.com"},
		"groups":      []string{"dev"},
		"password":    "password",
		"extra": map[string]interface{}{
			"employee_id":     "EMP12345",
			"employee_type":   "IT",
			"employee_number": 42,
			"test_flag":       "TRUE",
			"tags":            []string{"tag1", "tag2"},
		},
	}

	res, _ := s.apiRequest("POST", "/api/admin/users", updateData)
	s.Assert().Equal(http.StatusCreated, res.StatusCode)

	res, body := s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response struct {
		Status string                             `json:"status"`
		Data   authentication.UserDetailsExtended `json:"data"`
	}

	err := json.Unmarshal(body, &response)
	s.Assert().NoError(err)
	s.Assert().NotNil(response.Data.Extra)
	s.Assert().Equal("EMP12345", response.Data.Extra["employee_id"])
	s.Assert().Equal("IT", response.Data.Extra["employee_type"])
	s.Assert().Equal(int64(42), response.Data.Extra["employee_number"])
	s.Assert().Equal(true, response.Data.Extra["test_flag"])
	s.Assert().Equal("tag1", response.Data.Extra["tags"].([]interface{})[0])

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldFailWhenPasswordProvided() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"password":     "password",
		"username":     username,
		"given_name":   "Bob",
		"family_name":  "Dylan",
		"display_name": "Updated Bob Dylan",
		"mail":         []string{"updated@example.com"},
	}

	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=given_name,family_name,display_name,mail,password", username), updateData)

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldFailWithInvalidEmail() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"username":     username,
		"given_name":   "Bob",
		"family_name":  "Dylan",
		"display_name": "Updated Bob Dylan",
		"mail":         []string{"invalid.example.com"},
	}

	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=given_name,family_name,display_name,mail", username), updateData)

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_GetUserGET_ShouldReturnNotFoundForNonexistentUser() {
	s.login(adminUsername, adminPassword)

	username := nonExistentUsername
	res, _ := s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", username), nil)

	s.Assert().Equal(http.StatusNotFound, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_DeleteUserDELETE_ShouldSucceedForNonNonexistentUser() {
	s.login(adminUsername, adminPassword)

	username := nonExistentUsername
	res, _ := s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)

	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldAddUserToExistingGroups() {
	s.login(adminUsername, adminPassword)

	res, _ := s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	newUser := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "test",
		"family_name": "user",
		"mail":        []string{"test-user@example.com"},
		"groups":      []string{"dev"},
		"password":    "password",
	}

	res, body := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user: %s", string(body)))

	username := testUserUsername
	updateData := map[string]interface{}{
		"groups": []string{"dev", "admins"},
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
	s.Assert().Equal([]string{"dev", "admins"}, response.Data.Groups)

	res, _ = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldRemoveUserFromGroups() {
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
		"groups": []string{"dev"},
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
	s.Assert().Equal([]string{"dev"}, response.Data.Groups)

	res, _ = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldReplaceAllUserGroups() {
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

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)

	groupPostBody := map[string]interface{}{
		"name": testGroupName,
	}

	res, body = s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusOK, res.StatusCode, fmt.Sprintf("Failed to create group: %s", string(body)))

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

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldErrorWhenAddingNonexistentGroups() {
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
		"groups": []string{"dev", "admins", nonExistentGroupName},
	}

	res, body = s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=groups", username), updateData)

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode,
		fmt.Sprintf("Failed to update user: %s", string(body)))

	res, body = s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", username), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response struct {
		Status string                             `json:"status"`
		Data   authentication.UserDetailsExtended `json:"data"`
	}

	err := json.Unmarshal(body, &response)
	s.Assert().NoError(err)
	s.Assert().Equal([]string{"dev", "admins"}, response.Data.Groups)

	res, _ = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldUpdateDisplayName() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"display_name": "Robert Dylan",
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=display_name", username), updateData)

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
	s.Assert().Equal("Robert Dylan", response.Data.DisplayName)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldUpdateGivenNameAndFamilyName() {
	s.login(adminUsername, adminPassword)

	updateData := map[string]interface{}{
		"given_name":  "Robert",
		"family_name": "Smith",
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=given_name,family_name", nonAdminUsername), updateData)

	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to update user: %s", string(body)))

	res, body = s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", nonAdminUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response struct {
		Status string                             `json:"status"`
		Data   authentication.UserDetailsExtended `json:"data"`
	}

	err := json.Unmarshal(body, &response)
	s.Assert().NoError(err)
	s.Assert().Equal("Robert", response.Data.GivenName)
	s.Assert().Equal("Smith", response.Data.FamilyName)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldUpdatePhoneNumber() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"phone_number": "+1234567890",
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=phone_number,phone_extension", username), updateData)

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
	s.Assert().Equal("+1234567890", response.Data.PhoneNumber)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldUpdateAddressFields() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"address": map[string]interface{}{
			"street_address": "123 Main St",
			"locality":       "Springfield",
			"region":         "IL",
			"postal_code":    "62701",
			"country":        "USA",
		},
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=address.street_address,address.locality,address.region,address.postal_code,address.country", username), updateData)

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
	s.Assert().NotNil(response.Data.Address)
	s.Assert().Equal("123 Main St", response.Data.Address.StreetAddress)
	s.Assert().Equal("Springfield", response.Data.Address.Locality)
	s.Assert().Equal("IL", response.Data.Address.Region)
	s.Assert().Equal("62701", response.Data.Address.PostalCode)
	s.Assert().Equal("USA", response.Data.Address.Country)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldUpdateSingleAddressField() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername

	setupData := map[string]interface{}{
		"address": map[string]interface{}{
			"street_address": "123 Main St",
			"locality":       "Springfield",
			"region":         "IL",
			"postal_code":    "62701",
			"country":        "USA",
		},
	}
	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=address.street_address,address.locality,address.region,address.postal_code,address.country", username), setupData)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	updateData := map[string]interface{}{
		"address": map[string]interface{}{
			"region": "CA",
		},
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=address.region", username), updateData)

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
	s.Assert().NotNil(response.Data.Address)
	s.Assert().Equal("CA", response.Data.Address.Region)
	s.Assert().Equal("123 Main St", response.Data.Address.StreetAddress)
	s.Assert().Equal("Springfield", response.Data.Address.Locality)
	s.Assert().Equal("62701", response.Data.Address.PostalCode)
	s.Assert().Equal("USA", response.Data.Address.Country)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldUpdateMultipleAddressFields() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername

	setupData := map[string]interface{}{
		"address": map[string]interface{}{
			"street_address": "123 Main St",
			"locality":       "Springfield",
			"region":         "IL",
			"postal_code":    "62701",
			"country":        "USA",
		},
	}
	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=address.street_address,address.locality,address.region,address.postal_code,address.country", username), setupData)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	updateData := map[string]interface{}{
		"address": map[string]interface{}{
			"region":      "NY",
			"postal_code": "10001",
		},
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=address.region,address.postal_code", username), updateData)

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
	s.Assert().NotNil(response.Data.Address)
	s.Assert().Equal("NY", response.Data.Address.Region)
	s.Assert().Equal("10001", response.Data.Address.PostalCode)
	s.Assert().Equal("123 Main St", response.Data.Address.StreetAddress)
	s.Assert().Equal("Springfield", response.Data.Address.Locality)
	s.Assert().Equal("USA", response.Data.Address.Country)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldUpdateProfileURLs() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"profile": "https://example.com/bob",
		"picture": "https://example.com/bob.jpg",
		"website": "https://bobdylan.com",
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=profile,picture,website", username), updateData)

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
	s.Assert().NotNil(response.Data.Profile)
	s.Assert().Equal("https://example.com/bob", response.Data.Profile.String())
	s.Assert().NotNil(response.Data.Picture)
	s.Assert().Equal("https://example.com/bob.jpg", response.Data.Picture.String())
	s.Assert().NotNil(response.Data.Website)
	s.Assert().Equal("https://bobdylan.com", response.Data.Website.String())
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldUpdateLocaleAndZoneInfo() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"locale":   "en-US",
		"zoneinfo": "America/New_York",
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=locale,zoneinfo", username), updateData)

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
	s.Assert().NotNil(response.Data.Locale)
	s.Assert().Equal("en-US", response.Data.Locale.String())
	s.Assert().Equal("America/New_York", response.Data.ZoneInfo)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldUpdateExtraFields() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"extra": map[string]interface{}{
			"employee_id":     "EMP12345",
			"employee_type":   "IT",
			"employee_number": 42,
			"test_flag":       "TRUE",
			"tags":            []string{"tag1", "tag2"},
		},
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=extra.employee_id,extra.employee_type,extra.employee_number,extra.test_flag,extra.tags", username), updateData)

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
	s.Assert().NotNil(response.Data.Extra)
	s.Assert().Equal("EMP12345", response.Data.Extra["employee_id"])
	s.Assert().Equal("IT", response.Data.Extra["employee_type"])
	s.Assert().Equal(int64(42), response.Data.Extra["employee_number"])
	s.Assert().Equal(true, response.Data.Extra["test_flag"])
	s.Assert().Equal("tag1", response.Data.Extra["tags"].([]interface{})[0])
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldClearOptionalFields() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername

	setupData := map[string]interface{}{
		"middle_name":     "James",
		"nickname":        "Bobby",
		"gender":          "male",
		"birthdate":       "1990-01-01",
		"phone_extension": "456",
	}
	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=middle_name,nickname,gender,birthdate,phone_extension", username), setupData)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	updateData := map[string]interface{}{}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=middle_name,nickname,gender,birthdate,phone_extension", username), updateData)

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
	s.Assert().Equal("", response.Data.MiddleName)
	s.Assert().Equal("", response.Data.Nickname)
	s.Assert().Equal("", response.Data.Gender)
	s.Assert().Equal("", response.Data.Birthdate)
	s.Assert().Equal("", response.Data.PhoneExtension)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldRequireUpdateMask() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"display_name": "Robert Dylan",
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s", username), updateData)

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
	s.Assert().Contains(string(body), "update_mask")
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldRejectInvalidFieldInMask() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"display_name": "Robert Dylan",
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=invalid_field", username), updateData)

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
	s.Assert().Contains(string(body), "invalid_field")
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldRejectPasswordInMask() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"password": "newpassword",
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=password", username), updateData)

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
	s.Assert().Contains(string(body), "password")
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldOnlyUpdateMaskedFields() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername

	setupData := map[string]interface{}{
		"display_name": "Bob Dylan",
		"given_name":   "Bob",
		"family_name":  "Dylan",
		"phone_number": "+1234567890",
	}
	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=display_name,given_name,family_name,phone_number", username), setupData)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	updateData := map[string]interface{}{
		"display_name": "Robert Dylan",
		"given_name":   "Robert",
		"family_name":  "Smith",
		"phone_number": "+9999999999",
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=display_name", username), updateData)

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
	s.Assert().Equal("Robert Dylan", response.Data.DisplayName)
	s.Assert().Equal("Bob", response.Data.GivenName)
	s.Assert().Equal("Dylan", response.Data.FamilyName)
	s.Assert().Equal("+1234567890", response.Data.PhoneNumber)
}
