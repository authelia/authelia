package authentication

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

type ActiveDirectoryUserManagement struct {
	provider *LDAPUserProvider
}

//nolint:gocyclo
func (a *ActiveDirectoryUserManagement) AddUser(userData *UserDetailsExtended) (err error) {
	if userData == nil || userData.UserDetails == nil {
		return fmt.Errorf("userData and userData.UserDetails cannot be nil")
	}

	if err = a.ValidateUserData(userData); err != nil {
		return fmt.Errorf("validation failed for user '%s': %w", userData.Username, err)
	}

	if userData.Password == "" {
		return fmt.Errorf("password is required to create user '%s'", userData.Username)
	}

	var client LDAPExtendedClient
	if client, err = a.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to create user '%s': %w", userData.Username, err)
	}

	defer func() {
		if err := a.provider.factory.ReleaseClient(client); err != nil {
			a.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	userDN := fmt.Sprintf("cn=%s,%s", ldap.EscapeFilter(userData.CommonName), a.provider.usersBaseDN)

	addRequest := ldap.NewAddRequest(userDN, nil)

	addRequest.Attribute(ldapAttrObjectClass, a.GetDefaultUserObjectClasses())

	addRequest.Attribute(ldapAttrCommonName, []string{userData.CommonName})
	addRequest.Attribute(a.provider.config.Attributes.Username, []string{userData.Username})
	addRequest.Attribute(a.provider.config.Attributes.FamilyName, []string{userData.FamilyName})

	var pwdEncoded string
	if pwdEncoded, err = encodingUTF16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", userData.Password)); err != nil {
		return fmt.Errorf("error occurred encoding password for user '%s': %w", userData.Username, err)
	}

	addRequest.Attribute(ldapAttributeUnicodePwd, []string{pwdEncoded})

	// Set userAccountControl to 512 (normal enabled account).
	// 512 = ADS_UF_NORMAL_ACCOUNT.
	addRequest.Attribute("userAccountControl", []string{"512"})

	if userData.GivenName != "" {
		addRequest.Attribute(a.provider.config.Attributes.GivenName, []string{userData.GivenName})
	}

	if userData.GetDisplayName() != "" {
		addRequest.Attribute(a.provider.config.Attributes.DisplayName, []string{userData.GetDisplayName()})
	}

	if len(userData.Emails) > 0 {
		addRequest.Attribute(a.provider.config.Attributes.Mail, []string{userData.Emails[0]})
	}

	a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.Nickname, userData.Nickname)
	a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.MiddleName, userData.MiddleName)
	a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.Gender, userData.Gender)
	a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.Birthdate, userData.Birthdate)
	a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.ZoneInfo, userData.ZoneInfo)
	a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.PhoneNumber, userData.PhoneNumber)
	a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.PhoneExtension, userData.PhoneExtension)

	a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.Locale, userData.GetLocale())
	a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.Profile, userData.GetProfile())
	a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.Picture, userData.GetPicture())
	a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.Website, userData.GetWebsite())

	if userData.Address != nil {
		a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.StreetAddress, userData.Address.StreetAddress)
		a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.Locality, userData.Address.Locality)
		a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.Region, userData.Address.Region)
		a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.PostalCode, userData.Address.PostalCode)
		a.addAttributeIfPresent(addRequest, a.provider.config.Attributes.Country, userData.Address.Country)
	}

	if userData.Extra != nil {
		for jsonKey, value := range userData.Extra {
			if value == nil {
				continue
			}

			ldapAttr := a.getLDAPAttributeForExtraField(jsonKey)
			if ldapAttr == "" {
				a.provider.log.Warnf("No LDAP attribute mapping found for extra field '%s', skipping", jsonKey)
				continue
			}

			a.addAttributeIfPresent(addRequest, ldapAttr, a.normalizeExtraAttributes(value))
		}
	}

	var controls []ldap.Control

	switch {
	case client.Discovery().Controls.MsftPwdPolHints:
		controls = append(controls, &controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints})
	case client.Discovery().Controls.MsftPwdPolHintsDeprecated:
		controls = append(controls, &controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHintsDeprecated})
	}

	if len(controls) > 0 {
		addRequest.Controls = controls
	}

	if err = client.Add(addRequest); err != nil {
		return fmt.Errorf("unable to add user '%s': %w", userData.Username, err)
	}

	if len(userData.Groups) > 0 {
		if err = a.UpdateUserGroups(userData.Username, userData.Groups); err != nil {
			return fmt.Errorf("failed to assign groups for user '%s': %w", userData.Username, err)
		}
	}

	return nil
}

func (a *ActiveDirectoryUserManagement) ModifyUser(username string, userData *UserDetailsExtended) error {
	return nil
}

func (a *ActiveDirectoryUserManagement) UpdateUser(username string, userData *UserDetailsExtended) (err error) {
	panic("implement me")
}

//nolint:gocyclo
func (a *ActiveDirectoryUserManagement) UpdateUserWithMask(username string, userData *UserDetailsExtended, updateMask []string) error {
	if userData == nil || userData.UserDetails == nil {
		return fmt.Errorf("userData and userData.UserDetails cannot be nil")
	}

	var (
		client LDAPExtendedClient
		err    error
	)
	if client, err = a.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to update user '%s': %w", username, err)
	}

	defer func() {
		if err := a.provider.factory.ReleaseClient(client); err != nil {
			a.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	profile, err := a.provider.getUserProfile(client, username)
	if err != nil {
		return fmt.Errorf("unable to retrieve user profile for update of user '%s': %w", username, err)
	}

	modifyRequest := ldap.NewModifyRequest(profile.DN, nil)

	for _, field := range updateMask {
		switch {
		case field == AttributeGivenName:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.GivenName, userData.GivenName)
		case field == AttributeFamilyName:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.FamilyName, userData.FamilyName)
		case field == AttributeMiddleName:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.MiddleName, userData.MiddleName)
		case field == AttributeNickname:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Nickname, userData.Nickname)
		case field == AttributeGender:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Gender, userData.Gender)
		case field == AttributeBirthdate:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Birthdate, userData.Birthdate)
		case field == AttributeZoneInfo:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.ZoneInfo, userData.ZoneInfo)
		case field == AttributePhoneNumber:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.PhoneNumber, userData.PhoneNumber)
		case field == AttributePhoneExtension:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.PhoneExtension, userData.PhoneExtension)
		case field == AttributeLocale:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Locale, userData.GetLocale())
		case field == AttributeProfile:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Profile, userData.GetProfile())
		case field == AttributePicture:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Picture, userData.GetPicture())
		case field == AttributeWebsite:
			a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Website, userData.GetWebsite())
		case field == AttributeDisplayName:
			if userData.GetDisplayName() != "" {
				a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.DisplayName, userData.GetDisplayName())
			}
		case field == AttributeMail:
			if len(userData.Emails) > 0 {
				a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Mail, userData.Emails[0])
			}
		case field == AttributeGroups:
			if userData.GetGroups() != nil {
				if err := a.UpdateUserGroups(username, userData.GetGroups()); err != nil {
					return err
				}
			}
		case strings.HasPrefix(field, PrefixAttributeExtra):
			extraField := strings.TrimPrefix(field, PrefixAttributeExtra)

			if userData.Extra != nil {
				if value, exists := userData.Extra[extraField]; exists && value != nil {
					ldapAttr := a.getLDAPAttributeForExtraField(extraField)
					if ldapAttr == "" {
						a.provider.log.Warnf("No LDAP attribute mapping found for extra field '%s', skipping", extraField)
						continue
					}

					a.replaceAttributeIfPresent(modifyRequest, ldapAttr, a.normalizeExtraAttributes(value))
				}
			}
		case field == AttributeAddress:
			if userData.Address != nil {
				a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.StreetAddress, userData.Address.StreetAddress)
				a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Locality, userData.Address.Locality)
				a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Region, userData.Address.Region)
				a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.PostalCode, userData.Address.PostalCode)
				a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Country, userData.Address.Country)
			}
		case strings.HasPrefix(field, PrefixAttributeAddress):
			if userData.Address != nil {
				subField := strings.TrimPrefix(field, PrefixAttributeAddress)
				switch subField {
				case AttributeAddressStreetAddress:
					a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.StreetAddress, userData.Address.StreetAddress)
				case AttributeAddressLocality:
					a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Locality, userData.Address.Locality)
				case AttributeAddressRegion:
					a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Region, userData.Address.Region)
				case AttributeAddressPostalCode:
					a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.PostalCode, userData.Address.PostalCode)
				case AttributeAddressCountry:
					a.replaceAttributeIfPresent(modifyRequest, a.provider.config.Attributes.Country, userData.Address.Country)
				}
			}
		}
	}

	if len(modifyRequest.Changes) == 0 {
		a.provider.log.Debugf("No changes detected for user '%s', skipping update", username)
		return nil
	}

	if err = a.provider.modify(client, modifyRequest); err != nil {
		if errorCode := getLDAPResultCode(err); errorCode != -1 {
			switch errorCode {
			case ldap.LDAPResultNoSuchAttribute,
				ldap.LDAPResultNoSuchObject:
				return nil
			case ldap.LDAPResultInvalidAttributeSyntax:
				return fmt.Errorf("invalid attribute syntax: %v", err)
			}
		}

		return fmt.Errorf("unable to update user '%s': %w", username, err)
	}

	return nil
}

func (a *ActiveDirectoryUserManagement) DeleteUser(username string) (err error) {
	var client LDAPExtendedClient
	if client, err = a.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to delete user '%s': %w", username, err)
	}

	defer func() {
		if err := a.provider.factory.ReleaseClient(client); err != nil {
			a.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	profile, err := a.provider.getUserProfile(client, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			a.provider.log.Debugf("User '%s' not found for deletion.", username)
		}

		return fmt.Errorf("unable to retrieve user profile for deletion of user '%s': %w", username, err)
	}

	deleteRequest := ldap.NewDelRequest(profile.DN, nil)
	if err = client.Del(deleteRequest); err != nil {
		return fmt.Errorf("unable to delete user '%s': %w", username, err)
	}

	a.provider.log.Debugf("User '%s' was successfully deleted.", username)

	return nil
}

// AddGroup creates a new group in Active Directory.
func (a *ActiveDirectoryUserManagement) AddGroup(groupName string) error {
	var (
		client LDAPExtendedClient
		err    error
	)
	if client, err = a.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to get LDAP client for group creation: %w", err)
	}

	defer func() {
		if err := a.provider.factory.ReleaseClient(client); err != nil {
			a.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	groupDN := a.provider.BuildGroupDN(groupName)

	exists, err := a.groupExists(client, groupName)
	if err != nil {
		return fmt.Errorf("failed to check if group '%s' exists: %w", groupName, err)
	}

	if exists {
		return fmt.Errorf("error creating group '%s': %w", groupName, ErrGroupExists)
	}

	addRequest := ldap.NewAddRequest(groupDN, nil)

	addRequest.Attribute(ldapAttrObjectClass, a.GetDefaultGroupObjectClasses())
	addRequest.Attribute(a.provider.config.Attributes.GroupName, []string{groupName})

	addRequest.Attribute(a.provider.config.Attributes.Username, []string{groupName})

	if err := client.Add(addRequest); err != nil {
		if getLDAPResultCode(err) == ldap.LDAPResultEntryAlreadyExists {
			return fmt.Errorf("error creating group '%s': %w", groupName, ErrGroupExists)
		}

		return fmt.Errorf("failed to create group '%s': %w", groupName, err)
	}

	a.provider.log.Infof("Successfully created group '%s'", groupName)

	return nil
}

// DeleteGroup deletes a group in Active Directory.
func (a *ActiveDirectoryUserManagement) DeleteGroup(groupName string) error {
	var (
		client LDAPExtendedClient
		err    error
	)
	if client, err = a.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to get LDAP client for group deletion: %w", err)
	}

	defer func() {
		if err := a.provider.factory.ReleaseClient(client); err != nil {
			a.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	groupDN := fmt.Sprintf("%s=%s,%s", a.provider.config.Attributes.GroupName, ldap.EscapeFilter(groupName), a.provider.groupsBaseDN)

	exists, err := a.groupExists(client, groupName)
	if err != nil {
		return fmt.Errorf("failed to check if group '%s' exists: %w", groupName, err)
	}

	if !exists {
		a.provider.log.Debugf("Group '%s' doesn't exist, nothing to delete", groupName)
		return ErrGroupNotFound
	}

	deleteRequest := ldap.NewDelRequest(groupDN, nil)

	if err := client.Del(deleteRequest); err != nil {
		return fmt.Errorf("failed to delete group '%s': %w", groupName, err)
	}

	a.provider.log.Infof("Successfully deleted group '%s'", groupName)

	return nil
}

// ListGroups returns a list of all group names.
func (a *ActiveDirectoryUserManagement) ListGroups() ([]string, error) {
	var (
		client LDAPExtendedClient
		err    error
	)
	if client, err = a.provider.factory.GetClient(); err != nil {
		return nil, fmt.Errorf("unable to get LDAP client for listing groups: %w", err)
	}

	defer func() {
		if err := a.provider.factory.ReleaseClient(client); err != nil {
			a.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	searchRequest := ldap.NewSearchRequest(
		a.provider.groupsBaseDN,
		ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=group)",
		[]string{a.provider.config.Attributes.GroupName},
		nil,
	)

	searchResult, err := a.provider.search(client, searchRequest)
	if err != nil {
		var ldapErr *ldap.Error
		if errors.As(err, &ldapErr) && ldapErr.ResultCode == ldap.LDAPResultNoSuchObject {
			return []string{}, nil
		}

		return nil, fmt.Errorf("error occurred searching for all groups: %w", err)
	}

	groups := make([]string, 0, len(searchResult.Entries))
	for _, entry := range searchResult.Entries {
		groupName := entry.GetAttributeValue(a.provider.config.Attributes.GroupName)
		if groupName != "" {
			groups = append(groups, groupName)
		}
	}

	return groups, nil
}

// replaceAttributeIfPresent replaces or deletes an LDAP attribute in a modify request.
func (a *ActiveDirectoryUserManagement) replaceAttributeIfPresent(req *ldap.ModifyRequest, ldapAttr string, value interface{}) {
	if ldapAttr == "" {
		return
	}

	if value == nil {
		req.Delete(ldapAttr, []string{})
		return
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			req.Delete(ldapAttr, []string{})
		} else {
			req.Replace(ldapAttr, []string{v})
		}
	case []string:
		if len(v) == 0 {
			req.Delete(ldapAttr, []string{})
		} else {
			req.Replace(ldapAttr, v)
		}
	}
}

func (a *ActiveDirectoryUserManagement) addAttributeIfPresent(req *ldap.AddRequest, ldapAttr string, value interface{}) {
	if ldapAttr == "" {
		return
	}

	if value == nil {
		return
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return
		}

		req.Attribute(ldapAttr, []string{v})
	case []string:
		if len(v) == 0 {
			return
		}

		req.Attribute(ldapAttr, v)
	}
}

// getLDAPAttributeForExtraField returns the LDAP attribute name for an extra field.
func (a *ActiveDirectoryUserManagement) getLDAPAttributeForExtraField(jsonKey string) string {
	for ldapAttr, extraAttr := range a.provider.config.Attributes.Extra {
		attrName := extraAttr.Name
		if attrName == "" {
			attrName = ldapAttr
		}

		if attrName == jsonKey {
			return ldapAttr
		}
	}

	return ""
}

// normalizeExtraAttributes normalizes extra attribute values for LDAP storage.
func (a *ActiveDirectoryUserManagement) normalizeExtraAttributes(value interface{}) interface{} {
	switch v := value.(type) {
	case []interface{}:
		strSlice := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				strSlice = append(strSlice, str)
			} else {
				strSlice = append(strSlice, fmt.Sprintf("%v", item))
			}
		}

		return strSlice
	case bool:
		if v {
			return BooleanValueTrue
		}

		return BooleanValueFalse
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	default:
		return value
	}
}

//nolint:gocyclo
func (a *ActiveDirectoryUserManagement) UpdateUserGroups(username string, groups []string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	var (
		client LDAPExtendedClient
		err    error
	)
	if client, err = a.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to get LDAP client for group update: %w", err)
	}

	defer func() {
		if err := a.provider.factory.ReleaseClient(client); err != nil {
			a.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	profile, err := a.provider.getUserProfile(client, username)
	if err != nil {
		return fmt.Errorf("unable to retrieve user profile for group update of user '%s': %w", username, err)
	}

	userDN := profile.DN

	currentUserGroups, err := a.getCurrentUserGroups(client, userDN)
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

	for _, groupName := range groupsToAdd {
		exists, err := a.groupExists(client, groupName)
		if err != nil {
			return fmt.Errorf("failed to check group existence for group '%s'", groupName)
		}

		if !exists {
			return fmt.Errorf("group '%s' does not exist: %w", groupName, ErrGroupNotFound)
		}
	}

	a.provider.log.Debugf("Group update for user '%s': adding %d groups, removing %d groups",
		username, len(groupsToAdd), len(groupsToRemove))

	var errs []error

	for _, groupName := range groupsToRemove {
		if err := a.removeUserFromGroup(client, userDN, groupName); err != nil {
			a.provider.log.WithError(err).Errorf("Failed to remove user '%s' from group '%s'", username, groupName)
			errs = append(errs, err)
		}
	}

	for _, groupName := range groupsToAdd {
		if err := a.addUserToGroup(client, groupName, userDN); err != nil {
			a.provider.log.WithError(err).Errorf("Failed to add user '%s' in group '%s'", username, groupName)
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// addUserToGroup adds a user to a group.
func (a *ActiveDirectoryUserManagement) addUserToGroup(client LDAPExtendedClient, groupName, userDN string) error {
	groupDN := fmt.Sprintf("%s=%s,%s",
		a.provider.config.Attributes.GroupName,
		ldap.EscapeFilter(groupName),
		a.provider.groupsBaseDN)

	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	modifyRequest.Add(a.provider.config.Attributes.GroupMember, []string{userDN})

	if err := client.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to add user '%s' to group '%s': %w", userDN, groupName, err)
	}

	return nil
}

// removeUserFromGroup removes a user from a group.
func (a *ActiveDirectoryUserManagement) removeUserFromGroup(client LDAPExtendedClient, userDN, groupName string) error {
	groupDN := fmt.Sprintf("%s=%s,%s",
		a.provider.config.Attributes.GroupName,
		ldap.EscapeFilter(groupName),
		a.provider.groupsBaseDN)

	exists, err := a.groupExists(client, groupName)
	if err != nil {
		return fmt.Errorf("failed to check if group '%s' exists: %w", groupName, err)
	}

	if !exists {
		a.provider.log.Debugf("Group '%s' doesn't exist, nothing to remove", groupName)
		return nil
	}

	isMember, err := a.isUserMemberOfGroup(client, userDN, groupDN)
	if err != nil {
		return fmt.Errorf("failed to check group membership for group '%s' on user '%s': %w", groupDN, userDN, err)
	}

	if !isMember {
		a.provider.log.Debugf("User is not a member of group '%s', nothing to remove", groupName)
		return nil
	}

	modifyRequest := ldap.NewModifyRequest(groupDN, nil)

	modifyRequest.Delete(a.provider.config.Attributes.GroupMember, []string{userDN})

	if err := client.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to remove user from group '%s': %w", groupName, err)
	}

	return nil
}

// getCurrentUserGroups retrieves all groups that the user currently belongs to.
func (a *ActiveDirectoryUserManagement) getCurrentUserGroups(client LDAPExtendedClient, userDN string) ([]string, error) {
	searchRequest := ldap.NewSearchRequest(
		a.provider.groupsBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf("(%s=%s)", a.provider.config.Attributes.GroupMember, ldap.EscapeFilter(userDN)),
		[]string{a.provider.config.Attributes.GroupName},
		nil,
	)

	searchResult, err := client.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user's current groups: %w", err)
	}

	var groups []string

	for _, entry := range searchResult.Entries {
		groupName := entry.GetAttributeValue(a.provider.config.Attributes.GroupName)
		if groupName != "" {
			groups = append(groups, groupName)
		}
	}

	return groups, nil
}

// isUserMemberOfGroup checks if a user is a member of a group.
func (a *ActiveDirectoryUserManagement) isUserMemberOfGroup(client LDAPExtendedClient, userDN, groupDN string) (bool, error) {
	searchRequest := ldap.NewSearchRequest(
		groupDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf("(%s=%s)", a.provider.config.Attributes.GroupMember, ldap.EscapeFilter(userDN)),
		[]string{"dn"},
		nil,
	)

	searchResult, err := client.Search(searchRequest)
	if err != nil {
		var ldapErr *ldap.Error
		if errors.As(err, &ldapErr) && ldapErr.ResultCode == ldap.LDAPResultNoSuchObject {
			return false, nil
		}

		return false, err
	}

	return len(searchResult.Entries) > 0, nil
}

// groupExists checks if a group exists in Active Directory.
func (a *ActiveDirectoryUserManagement) groupExists(client LDAPExtendedClient, groupName string) (bool, error) {
	groupDN := fmt.Sprintf("%s=%s,%s",
		a.provider.config.Attributes.GroupName,
		ldap.EscapeFilter(groupName),
		a.provider.groupsBaseDN)

	searchRequest := ldap.NewSearchRequest(
		groupDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		1, 0, false,
		"(objectClass=group)",
		[]string{"dn"},
		nil,
	)

	searchResult, err := client.Search(searchRequest)
	if err != nil {
		var ldapErr *ldap.Error
		if errors.As(err, &ldapErr) && ldapErr.ResultCode == ldap.LDAPResultNoSuchObject {
			return false, nil
		}

		return false, err
	}

	return len(searchResult.Entries) > 0, nil
}

func (a *ActiveDirectoryUserManagement) GetRequiredAttributes() []string {
	requiredFieldNames := GetBaseRequiredAttributesForImplementation(schema.LDAPImplementationActiveDirectory)

	return append(requiredFieldNames, a.provider.config.UserManagement.RequiredAttributes...)
}

func (a *ActiveDirectoryUserManagement) GetSupportedAttributes() map[string]UserManagementAttributeMetadata {
	blocklist := map[string]bool{
		"group_member":       true,
		"group_name":         true,
		"member_of":          true,
		"distinguished_name": true,
	}

	fieldNames := getFieldNames(a.provider.config.Attributes)

	metadata := make(map[string]UserManagementAttributeMetadata, len(fieldNames))

	for _, fieldName := range fieldNames {
		if blocklist[fieldName] {
			continue
		}

		if meta, exists := attributeMetadataMap[fieldName]; exists {
			metadata[fieldName] = meta
		}
	}

	if meta, exists := attributeMetadataMap["password"]; exists {
		metadata["password"] = meta
	}

	if meta, exists := attributeMetadataMap["groups"]; exists {
		metadata["groups"] = meta
	}

	for key, extraAttr := range a.provider.config.Attributes.Extra {
		attrName := extraAttr.Name
		if attrName == "" {
			attrName = key
		}

		var inputType AttributeType

		switch extraAttr.ValueType {
		case ValueTypeBoolean:
			inputType = Checkbox
		case ValueTypeInteger:
			inputType = Number
		case ValueTypeString, "":
			inputType = Text
		default:
			inputType = Text
		}

		metadata[PrefixAttributeExtra+attrName] = UserManagementAttributeMetadata{
			Type:     inputType,
			Multiple: extraAttr.MultiValued,
		}
	}

	return metadata
}

// GetDefaultUserObjectClasses returns the default object classes for users.
func (a *ActiveDirectoryUserManagement) GetDefaultUserObjectClasses() []string {
	return a.provider.config.UserManagement.UserObjectClasses
}

// GetDefaultGroupObjectClasses returns the default object classes for groups.
func (a *ActiveDirectoryUserManagement) GetDefaultGroupObjectClasses() []string {
	return a.provider.config.UserManagement.GroupObjectClasses
}

func (a *ActiveDirectoryUserManagement) ValidateUserData(userData *UserDetailsExtended) error {
	if userData == nil {
		return fmt.Errorf("user data cannot be nil")
	}

	requiredAttributes := a.GetRequiredAttributes()

	attributeValues := a.buildAttributeValueMap(userData)

	var missingAttributes []string

	for _, attr := range requiredAttributes {
		value, exists := attributeValues[attr]
		if !exists || utils.IsEmptyValue(value) {
			missingAttributes = append(missingAttributes, attr)
		}
	}

	if len(missingAttributes) > 0 {
		return fmt.Errorf("missing required attributes: %s", strings.Join(missingAttributes, ", "))
	}

	if userData.CommonName == "" {
		if userData.GetGivenName() != "" {
			userData.CommonName = fmt.Sprintf("%s %s", userData.GetGivenName(), userData.GetFamilyName())
		} else {
			userData.CommonName = userData.GetFamilyName()
		}
	}

	if userData.UserDetails != nil && len(userData.GetEmails()) > 0 {
		if len(userData.GetEmails()) > 1 {
			return fmt.Errorf("multiple emails not supported, only one email address is allowed")
		}

		for _, email := range userData.GetEmails() {
			if !utils.ValidateEmailString(email) {
				return fmt.Errorf("invalid email address: %s", email)
			}
		}
	}

	return nil
}

// buildAttributeValueMap creates a map of attribute names to their values from UserDetailsExtended.
func (a *ActiveDirectoryUserManagement) buildAttributeValueMap(userData *UserDetailsExtended) map[string]interface{} {
	values := make(map[string]interface{})

	values[AttributeUsername] = userData.GetUsername()
	values[AttributePassword] = userData.Password
	values[AttributeDisplayName] = userData.GetDisplayName()
	values[AttributeGivenName] = userData.GetGivenName()
	values[AttributeFamilyName] = userData.GetFamilyName()
	values[AttributeMiddleName] = userData.MiddleName
	values[AttributeNickname] = userData.Nickname
	values[AttributeCommonName] = userData.CommonName

	values[AttributeProfile] = userData.GetProfile()
	values[AttributePicture] = userData.GetPicture()
	values[AttributeWebsite] = userData.GetWebsite()
	values[AttributeGender] = userData.Gender
	values[AttributeBirthdate] = userData.Birthdate
	values[AttributeLocale] = userData.GetLocale()
	values[AttributeZoneInfo] = userData.ZoneInfo

	values[AttributePhoneNumber] = userData.PhoneNumber
	values[AttributePhoneExtension] = userData.PhoneExtension

	if len(userData.GetEmails()) > 0 {
		values[AttributeMail] = userData.GetEmails()[0]
		values["emails"] = userData.GetEmails()
	}

	if userData.Address != nil {
		values[AttributeAddressStreetAddress] = userData.Address.StreetAddress
		values[AttributeAddressLocality] = userData.Address.Locality
		values[AttributeAddressRegion] = userData.Address.Region
		values[AttributeAddressPostalCode] = userData.Address.PostalCode
		values[AttributeAddressCountry] = userData.Address.Country
	}

	if userData.Extra != nil {
		for ldapAttr, attrConfig := range a.provider.config.Attributes.Extra {
			attrName := attrConfig.Name
			if attrName == "" {
				attrName = ldapAttr
			}

			if value, exists := userData.Extra[attrName]; exists {
				values[attrName] = value
			}
		}
	}

	return values
}

func (a *ActiveDirectoryUserManagement) ValidatePartialUpdate(userData *UserDetailsExtended, updateMask []string) error {
	if userData == nil {
		return fmt.Errorf("user data cannot be nil")
	}

	maskSet := make(map[string]bool)
	for _, field := range updateMask {
		maskSet[field] = true
	}

	if maskSet[AttributeMail] && userData.UserDetails != nil && len(userData.GetEmails()) > 0 {
		for _, email := range userData.GetEmails() {
			if !utils.ValidateEmailString(email) {
				return fmt.Errorf("invalid email address: %s", email)
			}
		}
	}

	return nil
}
