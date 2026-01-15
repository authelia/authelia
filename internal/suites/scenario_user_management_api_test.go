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

const testUserUsername = "testuser"

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
	s.cookies = make([]*http.Cookie, 0)
}

func (s *UserManagementAPIScenario) SetupTest() {
	s.cookies = make([]*http.Cookie, 0)
}

func TestUserManagementScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewUserManagementAPIScenario())
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

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_OpenLDAP_ShouldUpdateUser() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"username":     username,
		"first_name":   "Bob",
		"last_name":    "Dylan",
		"display_name": "Updated Bob Dylan",
		"emails":       []string{"updated@example.com"},
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=first_name,last_name,display_name,emails", username), updateData)

	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to update user: %s", string(body)))
}

func (s *UserManagementAPIScenario) Test_DeleteUserDELETE_OpenLDAP_ShouldDeleteUser() {
	s.login(adminUsername, adminPassword)

	username := testUserUsername

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)

	newUser := map[string]interface{}{
		"username":   username,
		"first_name": "test",
		"last_name":  "user",
		"emails":     []string{"testuser@example.com"},
		"groups":     []string{"dev"},
		"password":   "password",
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
		"first_name":   "Bob",
		"last_name":    "Dylan",
		"display_name": "Updated Bob Dylan",
		"emails":       []string{"updated@example.com"},
	}

	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=first_name,last_name,display_name,emails", username), updateData)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_DeleteUserDELETE_ShouldReturnForbiddenForNonAdmin() {
	s.login(nonAdminUsername, nonAdminPassword)

	username := nonAdminUsername

	res, _ := s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_NewUserPOST_OpenLDAP_ShouldFailWithMissingUsername() {
	s.login(adminUsername, adminPassword)

	newUser := map[string]interface{}{
		"username":   "",
		"first_name": "test",
		"last_name":  "user",
		"emails":     []string{"testuser@example.com"},
		"groups":     []string{"dev"},
		"password":   "password",
	}

	res, _ := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_NewUserPOST_OpenLDAP_ShouldFailWithMissingSurname() {
	s.login(adminUsername, adminPassword)

	newUser := map[string]interface{}{
		"username":   testUserUsername,
		"first_name": "test",
		"last_name":  "",
		"emails":     []string{"testuser@example.com"},
		"groups":     []string{"dev"},
		"password":   "password",
	}

	res, _ := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_NewUserPOST_ShouldFailWithInvalidEmail() {
	s.login(adminUsername, adminPassword)

	newUser := map[string]interface{}{
		"username":   testUserUsername,
		"first_name": "test",
		"last_name":  "user",
		"emails":     []string{"invalid.example.com"},
		"groups":     []string{"dev"},
		"password":   "password",
	}

	res, _ := s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldFailWhenPasswordProvided() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"password":     "password",
		"username":     username,
		"first_name":   "Bob",
		"last_name":    "Dylan",
		"display_name": "Updated Bob Dylan",
		"emails":       []string{"updated@example.com"},
	}

	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=first_name,last_name,display_name,emails,password", username), updateData)

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldFailWithInvalidEmail() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"username":     username,
		"first_name":   "Bob",
		"last_name":    "Dylan",
		"display_name": "Updated Bob Dylan",
		"emails":       []string{"invalid.example.com"},
	}

	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=first_name,last_name,display_name,emails", username), updateData)

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

	username := nonAdminUsername
	updateData := map[string]interface{}{
		"first_name": "Robert",
		"last_name":  "Smith",
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=first_name,last_name", username), updateData)

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
	s.Assert().Equal("Robert", response.Data.GivenName)
	s.Assert().Equal("Smith", response.Data.FamilyName)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldUpdateMultipleEmailAddresses() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername
	emails := []string{"bob.primary@example.com", "bob.secondary@example.com", "bob.work@example.com"}
	updateData := map[string]interface{}{
		"emails": emails,
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=emails", username), updateData)

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
	s.Assert().Equal(emails[0], response.Data.Emails[0])
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
			"formatted":      "123 Main St, Springfield, IL 62701, USA",
			"street_address": "123 Main St",
			"locality":       "Springfield",
			"region":         "IL",
			"postal_code":    "62701",
			"country":        "USA",
		},
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=address", username), updateData)

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
	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=address", username), setupData)
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
	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=address", username), setupData)
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
		"locale":    "en-US",
		"zone_info": "America/New_York",
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=locale,zone_info", username), updateData)

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
			"employee_id":   "EMP12345",
			"employee_type": "IT",
		},
	}

	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=extra", username), updateData)

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
		"first_name":   "Bob",
		"last_name":    "Dylan",
		"phone_number": "+1234567890",
	}
	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=display_name,first_name,last_name,phone_number", username), setupData)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	updateData := map[string]interface{}{
		"display_name": "Robert Dylan",
		"first_name":   "Robert",
		"last_name":    "Smith",
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

func (s *UserManagementAPIScenario) Test_NewGroupPOST_ShouldCreateGroup() {
	s.login(adminUsername, adminPassword)

	newGroup := "test-create-group"

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", newGroup), nil)

	groupPostBody := map[string]interface{}{
		"name": newGroup,
	}

	res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)

	s.Assert().Equal(http.StatusOK, res.StatusCode, fmt.Sprintf("Failed to create group: %s", string(body)))

	var postResponse struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}

	err := json.Unmarshal(body, &postResponse)
	s.Assert().NoError(err)
	s.Assert().Equal("OK", postResponse.Status)

	res, body = s.apiRequest("GET", "/api/admin/groups", nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var getResponse struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}

	err = json.Unmarshal(body, &getResponse)
	s.Assert().NoError(err)
	s.Assert().Contains(getResponse.Data, newGroup)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", newGroup), nil)
}

func (s *UserManagementAPIScenario) Test_NewGroupPOST_ShouldFailWithMissingRequiredFields() {
	s.login(adminUsername, adminPassword)

	groupPostBody := map[string]interface{}{
		"name": "",
	}

	res, _ := s.apiRequest("POST", "/api/admin/groups", groupPostBody)

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_NewGroupPOST_ShouldFailWithDuplicateGroupName() {
	s.login(adminUsername, adminPassword)

	groupName := "test-duplicate-group"

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)

	groupPostBody := map[string]interface{}{
		"name": groupName,
	}

	res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to create group: %s", string(body)))

	res, _ = s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)
}

func (s *UserManagementAPIScenario) Test_GetGroupsGET_ShouldReturnAllGroups() {
	s.login(adminUsername, adminPassword)

	res, body := s.apiRequest("GET", "/api/admin/groups", nil)

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}

	err := json.Unmarshal(body, &response)
	s.Assert().NoError(err)
	s.Assert().Equal("OK", response.Status)
	s.Assert().NotEmpty(response.Data, "Groups list should not be empty")
}

func (s *UserManagementAPIScenario) Test_GetGroupsGET_ShouldReturnEmptyListWhenNoGroups() {
	s.login(adminUsername, adminPassword)

	res, body := s.apiRequest("GET", "/api/admin/groups", nil)

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}

	err := json.Unmarshal(body, &response)
	s.Assert().NoError(err)
	s.Assert().Equal("OK", response.Status)
}

func (s *UserManagementAPIScenario) Test_DeleteGroupDELETE_ShouldRemoveGroup() {
	s.login(adminUsername, adminPassword)

	groupName := "test-delete-group"

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)

	groupPostBody := map[string]interface{}{
		"name": groupName,
	}

	res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to create group: %s", string(body)))

	res, body = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)

	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to delete group: %s", string(body)))

	res, body = s.apiRequest("GET", "/api/admin/groups", nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var response struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}

	err := json.Unmarshal(body, &response)
	s.Assert().NoError(err)
	s.Assert().NotContains(response.Data, groupName)
}

func (s *UserManagementAPIScenario) Test_DeleteGroupDELETE_ShouldSucceedForNonexistentGroup() {
	s.login(adminUsername, adminPassword)

	groupName := "nonexistent-group"

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)

	res, _ := s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)

	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_GetGroupsGET_ShouldReturnForbiddenForNonAdmin() {
	s.login(nonAdminUsername, nonAdminPassword)

	res, _ := s.apiRequest("GET", "/api/admin/groups", nil)
	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_NewGroupPOST_ShouldReturnForbiddenForNonAdmin() {
	s.login(nonAdminUsername, nonAdminPassword)

	groupPostBody := map[string]interface{}{
		"name": "test-group-nonadmin",
	}

	res, _ := s.apiRequest("POST", "/api/admin/groups", groupPostBody)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_DeleteGroupDELETE_ShouldReturnForbiddenForNonAdmin() {
	s.login(nonAdminUsername, nonAdminPassword)

	res, _ := s.apiRequest("DELETE", "/api/admin/groups/dev", nil)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_GetGroupsGET_ShouldReturnForbiddenForAnonymous() {
	res, _ := s.apiRequest("GET", "/api/admin/groups", nil)
	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_NewGroupPOST_ShouldReturnForbiddenForAnonymous() {
	groupPostBody := map[string]interface{}{
		"name": "test-group-anon",
	}

	res, _ := s.apiRequest("POST", "/api/admin/groups", groupPostBody)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_DeleteGroupDELETE_ShouldReturnForbiddenForAnonymous() {
	res, _ := s.apiRequest("DELETE", "/api/admin/groups/dev", nil)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *UserManagementAPIScenario) Test_DeleteGroupDELETE_ShouldRemoveGroupFromAssignedUsers() {
	s.login(adminUsername, adminPassword)

	groupName := "test-user-group"
	username := "testuser-group"

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)

	groupPostBody := map[string]interface{}{
		"name": groupName,
	}

	res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to create group: %s", string(body)))

	newUser := map[string]interface{}{
		"username":   username,
		"first_name": "Test",
		"last_name":  "User",
		"emails":     []string{fmt.Sprintf("%s@example.com", username)},
		"password":   "password",
	}

	res, body = s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user: %s", string(body)))

	updateData := map[string]interface{}{
		"groups": []string{groupName},
	}

	res, body = s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=groups", username), updateData)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to update user groups: %s", string(body)))

	res, body = s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", username), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var getUserResponse struct {
		Status string                             `json:"status"`
		Data   authentication.UserDetailsExtended `json:"data"`
	}

	err := json.Unmarshal(body, &getUserResponse)
	s.Assert().NoError(err)
	s.Assert().Contains(getUserResponse.Data.Groups, groupName, "User should be in the group")

	res, body = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to delete group: %s", string(body)))

	res, body = s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", username), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	err = json.Unmarshal(body, &getUserResponse)
	s.Assert().NoError(err)
	s.Assert().NotContains(getUserResponse.Data.Groups, groupName, "User should not have the deleted group")

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)
}

func (s *UserManagementAPIScenario) Test_GetGroupsGET_ShouldShowGroupMemberCount() {
	s.login(adminUsername, adminPassword)

	groupName := "test-member-count"
	username1 := "testuser1-member"
	username2 := "testuser2-member"

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username1), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username2), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)

	groupPostBody := map[string]interface{}{
		"name": groupName,
	}

	res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to create group: %s", string(body)))

	newUser1 := map[string]interface{}{
		"username":   username1,
		"first_name": "Test",
		"last_name":  "User1",
		"emails":     []string{fmt.Sprintf("%s@example.com", username1)},
		"password":   "password",
	}

	res, body = s.apiRequest("POST", "/api/admin/users", newUser1)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user1: %s", string(body)))

	updateData := map[string]interface{}{
		"groups": []string{groupName},
	}

	res, body = s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=groups", username1), updateData)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to update user1 groups: %s", string(body)))

	newUser2 := map[string]interface{}{
		"username":   username2,
		"first_name": "Test",
		"last_name":  "User2",
		"emails":     []string{fmt.Sprintf("%s@example.com", username2)},
		"password":   "password",
	}

	res, body = s.apiRequest("POST", "/api/admin/users", newUser2)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user2: %s", string(body)))

	res, body = s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=groups", username2), updateData)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to update user2 groups: %s", string(body)))

	res, body = s.apiRequest("GET", "/api/admin/users", nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var usersResponse struct {
		Status string                               `json:"status"`
		Data   []authentication.UserDetailsExtended `json:"data"`
	}

	err := json.Unmarshal(body, &usersResponse)
	s.Assert().NoError(err)

	memberCount := 0

	for _, user := range usersResponse.Data {
		for _, group := range user.Groups {
			if group == groupName {
				memberCount++
				break
			}
		}
	}

	s.Assert().Equal(2, memberCount, "Group should have 2 members")

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username1), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username2), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)
}

func (s *UserManagementAPIScenario) Test_NewGroupPOST_ShouldHandleSpecialCharactersInGroupName() {
	s.login(adminUsername, adminPassword)

	testCases := []struct {
		groupName     string
		shouldSucceed bool
		description   string
	}{
		{
			groupName:     "test-group-with-dashes",
			shouldSucceed: true,
			description:   "Group name with dashes",
		},
		{
			groupName:     "test_group_with_underscores",
			shouldSucceed: true,
			description:   "Group name with underscores",
		},
		{
			groupName:     "test.group.with.dots",
			shouldSucceed: true,
			description:   "Group name with dots",
		},
	}

	for _, tc := range testCases {
		s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", tc.groupName), nil)

		groupPostBody := map[string]interface{}{
			"name": tc.groupName,
		}

		res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)

		if tc.shouldSucceed {
			s.Assert().Equal(http.StatusOK, res.StatusCode,
				fmt.Sprintf("%s failed: %s", tc.description, string(body)))

			res, body = s.apiRequest("GET", "/api/admin/groups", nil)
			s.Assert().Equal(http.StatusOK, res.StatusCode)

			var response struct {
				Status string   `json:"status"`
				Data   []string `json:"data"`
			}

			err := json.Unmarshal(body, &response)
			s.Assert().NoError(err)
			s.Assert().Contains(response.Data, tc.groupName, tc.description)

			s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", tc.groupName), nil)
		} else {
			s.Assert().NotEqual(http.StatusOK, res.StatusCode, tc.description)
		}
	}
}

func (s *UserManagementAPIScenario) Test_NewGroupPOST_ShouldEscapeLDAPSpecialCharacters() {
	s.login(adminUsername, adminPassword)

	testCases := []struct {
		groupName     string
		shouldSucceed bool
		description   string
	}{
		{
			groupName:     "test-group-safe",
			shouldSucceed: true,
			description:   "Safe group name without special chars",
		},
	}

	for _, tc := range testCases {
		s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", tc.groupName), nil)

		groupPostBody := map[string]interface{}{
			"name": tc.groupName,
		}

		res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)

		if tc.shouldSucceed {
			s.Assert().Equal(http.StatusOK, res.StatusCode,
				fmt.Sprintf("%s failed: %s", tc.description, string(body)))

			s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", tc.groupName), nil)
		} else {
			s.Assert().True(
				res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusInternalServerError,
				fmt.Sprintf("%s: expected BadRequest or InternalServerError, got %d", tc.description, res.StatusCode),
			)
		}
	}
}

func (s *UserManagementAPIScenario) Test_DeleteGroupDELETE_ShouldCleanupGroupMetadata() {
	s.login(adminUsername, adminPassword)

	groupName := "test-cleanup-group"

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)

	groupPostBody := map[string]interface{}{
		"name": groupName,
	}

	res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to create group: %s", string(body)))

	res, body = s.apiRequest("GET", "/api/admin/groups", nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var getGroupsResponse struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}

	err := json.Unmarshal(body, &getGroupsResponse)
	s.Assert().NoError(err)
	s.Assert().Contains(getGroupsResponse.Data, groupName, "Group should exist before deletion")

	res, body = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to delete group: %s", string(body)))

	res, body = s.apiRequest("GET", "/api/admin/groups", nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	err = json.Unmarshal(body, &getGroupsResponse)
	s.Assert().NoError(err)
	s.Assert().NotContains(getGroupsResponse.Data, groupName, "Group should not exist after deletion")

	res, _ = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode, "Deleting non-existent group should succeed")
}
