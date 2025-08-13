package authentication

import (
	"errors"
	"fmt"

	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/go-ldap/ldap/v3"
)

type RFC2307bisUserManagement struct {
	provider *LDAPUserProvider
}

func (r *RFC2307bisUserManagement) GetRequiredFields() []string {
	return []string{
		"Username",
		"Password",
		"CommonName",
		"FamilyName",
	}
}

func (r *RFC2307bisUserManagement) GetSupportedFields() []string {
	return []string{
		"Username",
		"Password",
		"CommonName",
		"GivenName",
		"FamilyName",
		"Email",
		"Emails",
		"Groups",
		"ObjectClass",
		"Extra",
	}
}

func (r *RFC2307bisUserManagement) GetDefaultObjectClasses() []string {
	return []string{
		"top",
		"person",
		"organizationalPerson",
		"inetOrgPerson",
	}
}

// GetDefaultGroupObjectClasses returns the default object classes for groups
func (r *RFC2307bisUserManagement) GetDefaultGroupObjectClasses() []string {
	return []string{
		"top",
		"groupOfNames",
	}
}

// GetFieldMetadata describes the fields that are required to create new users for the RFC2307bis Backend.
func (r *RFC2307bisUserManagement) GetFieldMetadata() map[string]FieldMetadata {
	return map[string]FieldMetadata{
		"Username": {
			Required:    true,
			DisplayName: "Username",
			Description: "Unique identifier for the user (maps to uid attribute)",
			Type:        "string",
			MaxLength:   100,
		},
		"Password": {
			Required:    true,
			DisplayName: "Password",
			Description: "User's password",
			Type:        "password",
		},
		"CommonName": {
			Required:    true,
			DisplayName: "Common Name",
			Description: "Full name or display name (maps to cn attribute)",
			Type:        "string",
		},
		"GivenName": {
			Required:    false,
			DisplayName: "First Name",
			Description: "User's first/given name",
			Type:        "string",
		},
		"FamilyName": {
			Required:    true,
			DisplayName: "Last Name",
			Description: "User's last/family name (maps to sn attribute)",
			Type:        "string",
		},
		"Email": {
			Required:    false,
			DisplayName: "Email Address",
			Description: "Primary email address",
			Type:        "email",
		},
		"Groups": {
			Required:    false,
			DisplayName: "Groups",
			Description: "Groups the user should be added to",
			Type:        "array",
		},
	}
}

//ValidateUserData validates the userDetails struct contains all the required fields for new users exist.
/*
	New Users in RFC2307bis are required to have the following attributes
	- sn (surname/lastname)
	- uid (username)
	- cn (common name, full name)
	- objectClasses:
	  - top
	  - person
	  - organizationalPerson
	  - inetOrgPerson

	cn is built from required last name and optional first name
*/
func (r *RFC2307bisUserManagement) ValidateUserData(userData *UserDetailsExtended) error {
	if userData == nil {
		return fmt.Errorf("user data cannot be nil")
	}

	if userData.GetUsername() == "" {
		return fmt.Errorf("username is required for RFC2307bis")
	}

	if userData.GetFamilyName() == "" {
		return fmt.Errorf("last name is required for RFC2307bis")
	}

	if userData.CommonName == "" {
		if userData.GetGivenName() != "" {
			userData.CommonName = fmt.Sprintf("%s %s", userData.GetGivenName(), userData.GetFamilyName())
		} else {
			userData.CommonName = userData.GetFamilyName()
		}
	}

	if userData.UserDetails != nil && len(userData.GetEmails()) > 0 {
		for _, email := range userData.GetEmails() {
			if !utils.ValidateEmailString(email) {
				return fmt.Errorf("invalid email address: %s", email)
			}
		}
	}

	return nil
}

func (r *RFC2307bisUserManagement) UpdateUser(username string, userData *UserDetailsExtended) (err error) {
	if userData == nil || userData.UserDetails == nil {
		return fmt.Errorf("userData and userData.UserDetails cannot be nil")
	}

	if userData.Password != "" {
		return fmt.Errorf("cannot modify user passwords via update user, please use the password reset or password change endpoints")
	}

	if err := r.ValidateUserData(userData); err != nil {
		return fmt.Errorf("validation failed for user '%s': %w", username, err)
	}

	var client ldap.Client
	if client, err = r.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to update user '%s': %w", username, err)
	}

	defer func() {
		if err := r.provider.factory.ReleaseClient(client); err != nil {
			r.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	profile, err := r.provider.getUserProfile(client, username)
	if err != nil {
		return fmt.Errorf("unable to retrieve user profile for update of user '%s': %w", username, err)
	}

	modifyRequest := ldap.NewModifyRequest(profile.DN, nil)

	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.GivenName, userData.GivenName)
	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.FamilyName, userData.FamilyName)

	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.MiddleName, userData.MiddleName)
	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Nickname, userData.Nickname)
	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Gender, userData.Gender)
	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Birthdate, userData.Birthdate)
	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.ZoneInfo, userData.ZoneInfo)
	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.PhoneNumber, userData.PhoneNumber)
	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.PhoneExtension, userData.PhoneExtension)

	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Locale, userData.GetLocale())
	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Profile, userData.GetProfile())
	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Picture, userData.GetPicture())
	r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Website, userData.GetWebsite())

	if userData.Address != nil {
		r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.StreetAddress, userData.Address.StreetAddress)
		r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Locality, userData.Address.Locality)
		r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Region, userData.Address.Region)
		r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.PostalCode, userData.Address.PostalCode)
		r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Country, userData.Address.Country)
	}

	if userData.GetGroups() != nil {
		err := r.UpdateGroups(username, userData.GetGroups())
		if err != nil {
			return err
		}
	}

	if userData.GetDisplayName() != "" {
		r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.DisplayName, userData.GetDisplayName())
	}

	if len(userData.UserDetails.Emails) > 0 {
		r.addAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Mail, userData.UserDetails.Emails[0])
	}

	for ldapAttr, value := range userData.Extra {
		if value != nil {
			r.addAttributeIfPresent(modifyRequest, ldapAttr, fmt.Sprintf("%v", value))
		}
	}

	if len(modifyRequest.Changes) == 0 {
		r.provider.log.Debugf("No changes detected for user '%s', skipping update", username)
		return nil
	}

	r.provider.log.Debugf("Sending modify request for user '%s' with %d changes:", username, len(modifyRequest.Changes))
	for i, change := range modifyRequest.Changes {
		r.provider.log.Debugf("  Change %d: %s %s = %v", i+1, change.Operation, change.Modification.Type, change.Modification.Vals)
	}

	if err = r.provider.modify(client, modifyRequest); err != nil {
		return fmt.Errorf("unable to update user '%s': %w", username, err)
	}

	r.provider.log.Infof("Successfully updated user '%s'", username)
	return nil
}

func (r *RFC2307bisUserManagement) AddUser(userData *UserDetailsExtended) (err error) {
	if userData == nil || userData.UserDetails == nil {
		return fmt.Errorf("userData and userData.UserDetails cannot be nil")
	}

	if err = r.ValidateUserData(userData); err != nil {
		return fmt.Errorf("validation failed for user '%s': %w", userData.Username, err)
	}

	if userData.Password == "" {
		return fmt.Errorf("password is required to create user '%s'", userData.Username)
	}

	var client ldap.Client
	if client, err = r.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to create user '%s': %w", userData.Username, err)
	}

	defer func() {
		if err := r.provider.factory.ReleaseClient(client); err != nil {
			r.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	userDN := fmt.Sprintf("%s=%s,%s", r.provider.config.Attributes.Username, ldap.EscapeFilter(userData.Username), r.provider.usersBaseDN)

	addRequest := ldap.NewAddRequest(userDN, nil)

	addRequest.Attribute("objectClass", r.GetDefaultObjectClasses())

	addRequest.Attribute(r.provider.config.Attributes.Username, []string{userData.UserDetails.Username})
	addRequest.Attribute("cn", []string{userData.CommonName})
	addRequest.Attribute("sn", []string{userData.FamilyName})
	addRequest.Attribute("userPassword", []string{userData.Password})

	// Optional attributes
	if userData.GivenName != "" {
		addRequest.Attribute(r.provider.config.Attributes.GivenName, []string{userData.GivenName})
	}
	if len(userData.UserDetails.Emails) > 0 {
		addRequest.Attribute(r.provider.config.Attributes.Mail, []string{userData.UserDetails.Emails[0]})
	}

	//Removed the microsoft server controls because this implementation doesn't use microsoft server :)

	if err = client.Add(addRequest); err != nil {
		return fmt.Errorf("failed to add user '%s': %w", userData.Username, err)
	}

	return nil
}

func (r *RFC2307bisUserManagement) DeleteUser(username string) (err error) {
	var client ldap.Client
	if client, err = r.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to delete user '%s': %w", username, err)
	}

	defer func() {
		if err := r.provider.factory.ReleaseClient(client); err != nil {
			r.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	profile, err := r.provider.getUserProfile(client, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			r.provider.log.Debugf("User '%s' not found for deletion.", username)
		}
		return fmt.Errorf("unable to retrieve user profile for deletion of user '%s': %w", username, err)
	}

	deleteRequest := ldap.NewDelRequest(profile.DN, nil)
	if err = client.Del(deleteRequest); err != nil {
		if referral, ok := r.provider.getReferral(err); ok {
			return r.handleReferralDelete(referral, deleteRequest)
		}
		return fmt.Errorf("unable to delete user '%s': %w", username, err)
	}

	r.provider.log.Debugf("User '%s' was successfully deleted.", username)
	return nil
}

func (r *RFC2307bisUserManagement) handleReferralDelete(referral string, deleteRequest *ldap.DelRequest) error {
	r.provider.log.Debugf("Attempting Delete on referred URL %s", referral)

	clientRef, err := r.provider.factory.GetClient(WithAddress(referral))
	if err != nil {
		return fmt.Errorf("error occurred connecting to referred LDAP server '%s': %w", referral, err)
	}

	defer func() {
		if err := r.provider.factory.ReleaseClient(clientRef); err != nil {
			r.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	if err = clientRef.Del(deleteRequest); err != nil {
		return fmt.Errorf("error occurred performing delete on referred LDAP server '%s': %w", referral, err)
	}

	return nil
}

func (r *RFC2307bisUserManagement) addAttributeIfPresent(req *ldap.ModifyRequest, ldapAttr, value string) {
	if ldapAttr == "" {
		r.provider.log.Debugf("Skipping attribute update - LDAP attribute name is empty")
		return
	}

	req.Replace(ldapAttr, []string{value})
}

func (r *RFC2307bisUserManagement) addPointerAttributeIfPresent(modifyRequest *ldap.ModifyRequest, attributeName string, value fmt.Stringer) {
	if value != nil {
		r.addAttributeIfPresent(modifyRequest, attributeName, value.String())
	} else {
		modifyRequest.Delete(attributeName, []string{})
	}
}

func (r *RFC2307bisUserManagement) UpdateGroups(username string, groups []string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	var client ldap.Client
	var err error
	if client, err = r.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to get LDAP client for group update: %w", err)
	}

	defer func() {
		if err := r.provider.factory.ReleaseClient(client); err != nil {
			r.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	profile, err := r.provider.getUserProfile(client, username)
	if err != nil {
		return fmt.Errorf("unable to retrieve user profile for group update of user '%s': %w", username, err)
	}

	userDN := profile.DN

	currentUserGroups, err := r.getCurrentUserGroups(client, userDN)
	if err != nil {
		return fmt.Errorf("unable to retrieve current groups for user '%s': %w", username, err)
	}

	currentGroupsMap := make(map[string]bool)
	for _, group := range currentUserGroups {
		currentGroupsMap[group] = true
	}

	targetGroupsMap := make(map[string]bool)
	for _, group := range groups {
		if group != "" {
			targetGroupsMap[group] = true
		}
	}

	var groupsToAdd, groupsToRemove []string

	for group := range targetGroupsMap {
		if !currentGroupsMap[group] {
			groupsToAdd = append(groupsToAdd, group)
		}
	}

	for group := range currentGroupsMap {
		if !targetGroupsMap[group] {
			groupsToRemove = append(groupsToRemove, group)
		}
	}

	r.provider.log.Debugf("Group update for user '%s': adding %d groups, removing %d groups",
		username, len(groupsToAdd), len(groupsToRemove))

	for _, groupName := range groupsToRemove {
		if err := r.removeUserFromGroup(client, userDN, groupName); err != nil {
			r.provider.log.WithError(err).Errorf("Failed to remove user '%s' from group '%s'", username, groupName)
		}
	}

	for _, groupName := range groupsToAdd {
		if err := r.addUserToGroup(client, userDN, username, groupName); err != nil {
			r.provider.log.WithError(err).Errorf("Failed to add user '%s' to group '%s'", username, groupName)
		}
	}

	r.provider.log.Infof("Successfully updated groups for user '%s'", username)
	return nil
}

// createGroup creates a new group in LDAP
func (r *RFC2307bisUserManagement) createGroup(client ldap.Client, groupName, groupDN string) error {
	addRequest := ldap.NewAddRequest(groupDN, nil)

	// RFC2307bis group object classes
	addRequest.Attribute("objectClass", []string{
		"top",
		"groupOfNames",
	})

	addRequest.Attribute(r.provider.config.Attributes.GroupName, []string{groupName})
	//placeholderDN := fmt.Sprintf("cn=placeholder,%s", r.provider.groupsBaseDN)
	addRequest.Attribute(r.provider.config.Attributes.GroupMember, []string{groupDN})

	if err := client.Add(addRequest); err != nil {
		return err
	}

	return nil
}

// deleteGroup deletes a group in LDAP
func (r *RFC2307bisUserManagement) deleteGroup(client ldap.Client, groupName, groupDN string) error {
	// Check if group exists first
	exists, err := r.groupExists(client, groupDN)
	if err != nil {
		return fmt.Errorf("failed to check if group '%s' exists: %w", groupName, err)
	}

	if !exists {
		r.provider.log.Debugf("Group '%s' doesn't exist, nothing to delete", groupName)
		return nil
	}

	deleteRequest := ldap.NewDelRequest(groupDN, nil)

	if err := client.Del(deleteRequest); err != nil {
		return fmt.Errorf("failed to delete group '%s': %w", groupName, err)
	}

	r.provider.log.Debugf("Successfully deleted group '%s'", groupName)
	return nil
}

func (r *RFC2307bisUserManagement) getGroupObject(client ldap.Client, groupDN string) (*ldap.Entry, error) {
	searchRequest := ldap.NewSearchRequest(
		groupDN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		1, 0, false,
		"(objectClass=*)",
		[]string{"*"},
		nil,
	)

	searchResult, err := r.provider.search(client, searchRequest)
	if err != nil {
		var ldapErr *ldap.Error
		if errors.As(err, &ldapErr) && ldapErr.ResultCode == ldap.LDAPResultNoSuchObject {
			return nil, nil
		}

		return nil, fmt.Errorf("error occurred searching for group '%s': %w", groupDN, err)
	}

	if len(searchResult.Entries) == 0 {
		return nil, nil
	}

	return searchResult.Entries[0], nil
}

// groupExists checks if a group exists in LDAP
func (r *RFC2307bisUserManagement) groupExists(client ldap.Client, groupDN string) (bool, error) {
	searchRequest := ldap.NewSearchRequest(
		groupDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		1, 0, false,
		"(objectClass=*)",
		[]string{"dn"},
		nil,
	)

	searchResult, err := client.Search(searchRequest)
	if err != nil {
		// Check if it's a "No Such Object" error
		if ldapErr, ok := err.(*ldap.Error); ok && ldapErr.ResultCode == ldap.LDAPResultNoSuchObject {
			return false, nil
		}
		return false, err
	}

	return len(searchResult.Entries) > 0, nil
}

// isUserMemberOfGroup checks if a user is already a member of a group
func (r *RFC2307bisUserManagement) isUserMemberOfGroup(client ldap.Client, userDN, groupDN string) (bool, error) {
	group, err := r.getGroupObject(client, groupDN)
	if err != nil {
		return false, err
	}

	if group == nil {
		return false, nil
	}

	members := group.GetAttributeValues(r.provider.config.Attributes.GroupMember)
	for _, member := range members {
		if member == userDN {
			return true, nil
		}
	}

	return false, nil
}

// removeUserFromGroup removes a user from a group
func (r *RFC2307bisUserManagement) removeUserFromGroup(client ldap.Client, userDN, groupName string) error {
	groupDN := fmt.Sprintf("%s=%s,%s",
		r.provider.config.Attributes.GroupName,
		ldap.EscapeFilter(groupName),
		r.provider.groupsBaseDN)

	exists, err := r.groupExists(client, groupDN)
	if err != nil {
		return fmt.Errorf("failed to check if group '%s' exists: %w", groupName, err)
	}

	if !exists {
		r.provider.log.Debugf("Group '%s' doesn't exist, nothing to remove", groupName)
		return nil
	}

	isMember, err := r.isUserMemberOfGroup(client, userDN, groupDN)
	if err != nil {
		return fmt.Errorf("failed to check group membership for group '%s' on user '%s': %w", groupDN, userDN, err)
	}

	if !isMember {
		r.provider.log.Debugf("User is not a member of group '%s', nothing to remove", groupName)
		return nil
	}

	modifyRequest := ldap.NewModifyRequest(groupDN, nil)

	modifyRequest.Delete(r.provider.config.Attributes.GroupMember, []string{userDN})

	if err := client.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to remove user '%s' from group '%s': %w", userDN, groupName, err)
	}

	r.provider.log.Debugf("Removed user from group '%s'", groupName)
	return nil
}

// addUserToGroup adds a user to a group, creating the group if it doesn't exist
func (r *RFC2307bisUserManagement) addUserToGroup(client ldap.Client, userDN, username, groupName string) error {
	groupDN := fmt.Sprintf("%s=%s,%s",
		r.provider.config.Attributes.GroupName,
		ldap.EscapeFilter(groupName),
		r.provider.groupsBaseDN)

	exists, err := r.groupExists(client, groupDN)
	if err != nil {
		return fmt.Errorf("failed to check if group '%s' exists: %w", groupName, err)
	}

	//TODO: allow conditional requirement for groups to be created prior to users being added -- `requireGroupsToExistPriorToUserMembership` -- or create group automatically
	if !exists {
		if err := r.createGroup(client, groupName, groupDN); err != nil {
			return fmt.Errorf("failed to create group '%s': %w", groupName, err)
		}
		r.provider.log.Infof("Created new group '%s'", groupName)
	}

	isMember, err := r.isUserMemberOfGroup(client, userDN, groupDN)
	if err != nil {
		return fmt.Errorf("failed to check group membership for group '%s' on user '%s' : %w", groupName, username, err)
	}

	if isMember {
		r.provider.log.Debugf("User '%s' is already a member of group '%s'", username, groupName)
		return nil
	}

	modifyRequest := ldap.NewModifyRequest(groupDN, nil)

	modifyRequest.Add(r.provider.config.Attributes.GroupMember, []string{userDN})

	if err := client.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to add user '%s' to group '%s': %w", username, groupName, err)
	}

	r.provider.log.Debugf("Added user '%s' to group '%s'", username, groupName)
	return nil
}

// getCurrentUserGroups retrieves all groups that the user currently belongs to
func (r *RFC2307bisUserManagement) getCurrentUserGroups(client ldap.Client, userDN string) ([]string, error) {
	searchRequest := ldap.NewSearchRequest(
		r.provider.groupsBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf("(%s=%s)", r.provider.config.Attributes.GroupMember, ldap.EscapeFilter(userDN)),
		[]string{r.provider.config.Attributes.GroupName},
		nil,
	)

	searchResult, err := client.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user's current groups: %w", err)
	}

	var groups []string
	for _, entry := range searchResult.Entries {
		groupName := entry.GetAttributeValue(r.provider.config.Attributes.GroupName)
		if groupName != "" {
			groups = append(groups, groupName)
		}
	}

	return groups, nil
}
