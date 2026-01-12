package suites

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/authentication"
)

const adminUsername = "john"
const adminPassword = "password"

const nonAdminUsername = "bob"
const nonAdminPassword = "password"

const nonExistentUsername = "nonexistentuser"

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

func (s *UserManagementAPIScenario) Test_NewUserPOST_OpenLDAP_ShouldCreateNewUser() {
	s.login(adminUsername, adminPassword)

	username := "testuser-create"

	_, _ = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)

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

	s.apiRequest("DELETE", "/api/admin/users/testuser-create", nil)
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

	username := "testuser"

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
		"username":   "testuser",
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
		"username":   "testuser",
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

	// First set a full address.
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

	// Now update only the region.
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
	// Region should be updated.
	s.Assert().Equal("CA", response.Data.Address.Region)
	// Other fields should remain unchanged.
	s.Assert().Equal("123 Main St", response.Data.Address.StreetAddress)
	s.Assert().Equal("Springfield", response.Data.Address.Locality)
	s.Assert().Equal("62701", response.Data.Address.PostalCode)
	s.Assert().Equal("USA", response.Data.Address.Country)
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldUpdateMultipleAddressFields() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername

	// First set a full address.
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

	// Update region and postal_code.
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
	// Updated fields.
	s.Assert().Equal("NY", response.Data.Address.Region)
	s.Assert().Equal("10001", response.Data.Address.PostalCode)

	// Unchanged fields.
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

	// First set some optional fields.
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

	// Request without update_mask should fail.
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

	// Request with invalid field in mask should fail.
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

	// Request with password in mask should fail.
	res, body := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=password", username), updateData)

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
	s.Assert().Contains(string(body), "password")
}

func (s *UserManagementAPIScenario) Test_ChangeUserPATCH_ShouldOnlyUpdateMaskedFields() {
	s.login(adminUsername, adminPassword)

	username := nonAdminUsername

	// First, set multiple fields.
	setupData := map[string]interface{}{
		"display_name": "Bob Dylan",
		"first_name":   "Bob",
		"last_name":    "Dylan",
		"phone_number": "+1234567890",
	}
	res, _ := s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=display_name,first_name,last_name,phone_number", username), setupData)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	// Now update only display_name, but send other fields too.
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
	// Only display_name should be updated.
	s.Assert().Equal("Robert Dylan", response.Data.DisplayName)
	// Other fields should remain unchanged.
	s.Assert().Equal("Bob", response.Data.GivenName)
	s.Assert().Equal("Dylan", response.Data.FamilyName)
	s.Assert().Equal("+1234567890", response.Data.PhoneNumber)
}

func (s *UserManagementAPIScenario) Test_NewGroupPOST_ShouldCreateGroup() {
	s.login(adminUsername, adminPassword)

	// Use a unique group name based on timestamp to avoid conflicts
	newGroup := fmt.Sprintf("test-group-%d", time.Now().UnixNano())

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

	// Clean up: delete the test group
	_, _ = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", newGroup), nil)
}

func (s *UserManagementAPIScenario) Test_GetGroupsGET_ShouldGetGroups() {}

// func (s *UserManagementAPIScenario) Test_NewUserPOST_ShouldCreateUserWithMultipleGroups() {
//	s.login(adminUsername, adminPassword)
//
//	username := "testuser"
//
//	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)
//
//	newUser := map[string]interface{}{
//		"username":   username,
//		"first_name": "test",
//		"last_name":  "user",
//		"emails":     []string{"testuser@example.com"},
//		"groups":     []string{"dev", "dev2", "dev3"},
//		"password":   "password",
//	}
//
//	res, body := s.apiRequest("POST", "/api/admin/users", newUser)
//	s.Assert().Equal(http.StatusCreated, res.StatusCode,
//		fmt.Sprintf("Failed to create user: %s", string(body)))
//
//	res, body = s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", username), nil)
//	s.Assert().Equal(http.StatusOK, res.StatusCode)
//
//	var response struct {
//		Status string                             `json:"status"`
//		Data   authentication.UserDetailsExtended `json:"data"`
//	}
//
//	err := json.Unmarshal(body, &response)
//	s.Assert().NoError(err)
//	s.Assert().Equal([]string{"dev", "dev2", "dev3"}, response.Data.Groups)
//
//	//s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", username), nil)
// }.
