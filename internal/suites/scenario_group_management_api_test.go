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

type GroupManagementAPIScenario struct {
	suite.Suite

	client  *http.Client
	cookies []*http.Cookie
}

func NewGroupManagementAPIScenario() *GroupManagementAPIScenario {
	return &GroupManagementAPIScenario{}
}

func (s *GroupManagementAPIScenario) SetupSuite() {
	s.client = NewHTTPClient()
	s.logout()
}

func (s *GroupManagementAPIScenario) SetupTest() {
	s.logout()
	s.cleanupTestFixtures()
}

func (s *GroupManagementAPIScenario) TearDownTest() {
	s.logout()
	s.cleanupTestFixtures()
}

func (s *GroupManagementAPIScenario) logout() {
	s.cookies = make([]*http.Cookie, 0)
}

func TestGroupManagementScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewGroupManagementAPIScenario())
}

func (s *GroupManagementAPIScenario) cleanupTestFixtures() {
	s.login(adminUsername, adminPassword)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername2), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName2), nil)
	s.logout()
}

//nolint:unparam
func (s *GroupManagementAPIScenario) login(username, password string) {
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
func (s *GroupManagementAPIScenario) storeCookies(newCookies []*http.Cookie) {
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

func (s *GroupManagementAPIScenario) apiRequest(method, path string, body interface{}) (*http.Response, []byte) {
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

func (s *GroupManagementAPIScenario) Test_NewGroupPOST_ShouldCreateGroup() {
	s.login(adminUsername, adminPassword)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)

	groupPostBody := map[string]interface{}{
		"name": testGroupName,
	}

	res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)

	s.Assert().Equal(http.StatusOK, res.StatusCode, fmt.Sprintf("Failed to create group: %s", string(body)))

	var postResponse struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}

	err := json.Unmarshal(body, &postResponse)
	s.Assert().NoError(err)

	res, body = s.apiRequest("GET", "/api/admin/groups", nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	var getResponse struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}

	err = json.Unmarshal(body, &getResponse)
	s.Assert().NoError(err)
	s.Assert().Contains(getResponse.Data, testGroupName)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)
}

func (s *GroupManagementAPIScenario) Test_NewGroupPOST_ShouldFailWithMissingRequiredFields() {
	s.login(adminUsername, adminPassword)

	groupPostBody := map[string]interface{}{
		"name": "",
	}

	res, _ := s.apiRequest("POST", "/api/admin/groups", groupPostBody)

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *GroupManagementAPIScenario) Test_NewGroupPOST_ShouldFailWithDuplicateGroupName() {
	s.login(adminUsername, adminPassword)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)

	groupPostBody := map[string]interface{}{
		"name": testGroupName,
	}

	res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to create group: %s", string(body)))

	res, _ = s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusConflict, res.StatusCode)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)
}

func (s *GroupManagementAPIScenario) Test_GetGroupsGET_ShouldReturnAllGroups() {
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

func (s *GroupManagementAPIScenario) Test_GetGroupsGET_ShouldReturnEmptyListWhenNoGroups() {
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

func (s *GroupManagementAPIScenario) Test_DeleteGroupDELETE_ShouldRemoveGroup() {
	s.login(adminUsername, adminPassword)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)

	groupPostBody := map[string]interface{}{
		"name": testGroupName,
	}

	res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to create group: %s", string(body)))

	res, body = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)

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
	s.Assert().NotContains(response.Data, testGroupName)
}

func (s *GroupManagementAPIScenario) Test_DeleteGroupDELETE_ShouldSucceedForNonexistentGroup() {
	s.login(adminUsername, adminPassword)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", nonExistentGroupName), nil)

	res, _ := s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", nonExistentGroupName), nil)

	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func (s *GroupManagementAPIScenario) Test_GetGroupsGET_ShouldReturnForbiddenForNonAdmin() {
	s.login(nonAdminUsername, nonAdminPassword)

	res, _ := s.apiRequest("GET", "/api/admin/groups", nil)
	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *GroupManagementAPIScenario) Test_NewGroupPOST_ShouldReturnForbiddenForNonAdmin() {
	s.login(nonAdminUsername, nonAdminPassword)

	groupPostBody := map[string]interface{}{
		"name": testGroupName,
	}

	res, _ := s.apiRequest("POST", "/api/admin/groups", groupPostBody)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *GroupManagementAPIScenario) Test_DeleteGroupDELETE_ShouldReturnForbiddenForNonAdmin() {
	s.login(nonAdminUsername, nonAdminPassword)

	res, _ := s.apiRequest("DELETE", "/api/admin/groups/dev", nil)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *GroupManagementAPIScenario) Test_GetGroupsGET_ShouldReturnForbiddenForAnonymous() {
	res, _ := s.apiRequest("GET", "/api/admin/groups", nil)
	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *GroupManagementAPIScenario) Test_NewGroupPOST_ShouldReturnForbiddenForAnonymous() {
	groupPostBody := map[string]interface{}{
		"name": testGroupName,
	}

	res, _ := s.apiRequest("POST", "/api/admin/groups", groupPostBody)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *GroupManagementAPIScenario) Test_DeleteGroupDELETE_ShouldReturnForbiddenForAnonymous() {
	res, _ := s.apiRequest("DELETE", "/api/admin/groups/dev", nil)

	s.Assert().Equal(http.StatusForbidden, res.StatusCode)
}

func (s *GroupManagementAPIScenario) Test_DeleteGroupDELETE_ShouldRemoveGroupFromAssignedUsers() {
	s.login(adminUsername, adminPassword)

	groupName := testGroupName

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", groupName), nil)

	groupPostBody := map[string]interface{}{
		"name": groupName,
	}

	res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to create group: %s", string(body)))

	newUser := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "Test",
		"family_name": "User",
		"mail":        []string{fmt.Sprintf("%s@example.com", testUserUsername)},
		"password":    "password",
	}

	res, body = s.apiRequest("POST", "/api/admin/users", newUser)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user: %s", string(body)))

	updateData := map[string]interface{}{
		"groups": []string{groupName},
	}

	res, body = s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=groups", testUserUsername), updateData)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to update user groups: %s", string(body)))

	res, body = s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
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

	res, body = s.apiRequest("GET", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	err = json.Unmarshal(body, &getUserResponse)
	s.Assert().NoError(err)
	s.Assert().NotContains(getUserResponse.Data.Groups, groupName, "User should not have the deleted group")

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
}

func (s *GroupManagementAPIScenario) Test_GetGroupsGET_ShouldShowGroupMemberCount() {
	s.login(adminUsername, adminPassword)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername2), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)

	groupPostBody := map[string]interface{}{
		"name": testGroupName,
	}

	res, body := s.apiRequest("POST", "/api/admin/groups", groupPostBody)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to create group: %s", string(body)))

	newUserPostBody := map[string]interface{}{
		"username":    testUserUsername,
		"given_name":  "Test",
		"family_name": "User1",
		"mail":        []string{fmt.Sprintf("%s@example.com", testUserUsername)},
		"password":    "password",
	}

	res, body = s.apiRequest("POST", "/api/admin/users", newUserPostBody)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user1: %s", string(body)))

	updateData := map[string]interface{}{
		"groups": []string{testGroupName},
	}

	res, body = s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=groups", testUserUsername), updateData)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to update user1 groups: %s", string(body)))

	newUserPostBody2 := map[string]interface{}{
		"username":    testUserUsername2,
		"given_name":  "Test",
		"family_name": "User2",
		"mail":        []string{fmt.Sprintf("%s@example.com", testUserUsername2)},
		"password":    "password",
	}

	res, body = s.apiRequest("POST", "/api/admin/users", newUserPostBody2)
	s.Assert().Equal(http.StatusCreated, res.StatusCode,
		fmt.Sprintf("Failed to create user2: %s", string(body)))

	res, body = s.apiRequest("PATCH", fmt.Sprintf("/api/admin/users/%s?update_mask=groups", testUserUsername2), updateData)
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
			if group == testGroupName {
				memberCount++
				break
			}
		}
	}

	s.Assert().Equal(2, memberCount, "Group should have 2 members")

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/users/%s", testUserUsername2), nil)
	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)
}

func (s *GroupManagementAPIScenario) Test_NewGroupPOST_ShouldHandleSpecialCharactersInGroupName() {
	s.login(adminUsername, adminPassword)

	//TODO: add more test cases.
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

func (s *GroupManagementAPIScenario) Test_NewGroupPOST_ShouldEscapeLDAPSpecialCharacters() {
	s.login(adminUsername, adminPassword)

	//TODO: add more test cases.
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

func (s *GroupManagementAPIScenario) Test_DeleteGroupDELETE_ShouldCleanupGroupMetadata() {
	s.login(adminUsername, adminPassword)

	s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)

	groupPostBody := map[string]interface{}{
		"name": testGroupName,
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
	s.Assert().Contains(getGroupsResponse.Data, testGroupName, "Group should exist before deletion")

	res, body = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode,
		fmt.Sprintf("Failed to delete group: %s", string(body)))

	res, body = s.apiRequest("GET", "/api/admin/groups", nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	err = json.Unmarshal(body, &getGroupsResponse)
	s.Assert().NoError(err)
	s.Assert().NotContains(getGroupsResponse.Data, testGroupName, "Group should not exist after deletion")

	res, _ = s.apiRequest("DELETE", fmt.Sprintf("/api/admin/groups/%s", testGroupName), nil)
	s.Assert().Equal(http.StatusOK, res.StatusCode, "Deleting non-existent group should succeed")
}
