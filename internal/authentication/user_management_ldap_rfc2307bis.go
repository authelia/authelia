package authentication

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"

	"github.com/authelia/authelia/v4/internal/utils"
)

type RFC2307bisUserManagement struct {
	provider *LDAPUserProvider
}

func (r *RFC2307bisUserManagement) GetRequiredFields() []string {
	return []string{
		"username",
		"password",
		"full_name",
		"last_name",
		"emails",
	}
}

func (r *RFC2307bisUserManagement) GetSupportedFields() []string {
	return []string{
		"display_name",
		"emails",
		"groups",
		"first_name",
		"last_name",
		"middle_name",
		"full_name",
		"nickname",
		"phone_number",
		"phone_extension",
		"profile",
		"picture",
		"website",
		"gender",
		"birthdate",
		"locale",
		"zone_info",
		"address",
		"address.street_address",
		"address.locality",
		"address.region",
		"address.postal_code",
		"address.country",
		"extra",
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

// GetDefaultGroupObjectClasses returns the default object classes for groups.
func (r *RFC2307bisUserManagement) GetDefaultGroupObjectClasses() []string {
	return []string{
		"top",
		"groupOfNames",
	}
}

// GetFieldMetadata describes the fields that are required to create new users for the RFC2307bis Backend.
func (r *RFC2307bisUserManagement) GetFieldMetadata() map[string]FieldMetadata {
	return map[string]FieldMetadata{
		"username": {
			DisplayName: "Username",
			Description: "Unique identifier for the user (maps to uid attribute)",
			Type:        "string",
			MaxLength:   100,
		},
		"password": {
			DisplayName: "Password",
			Description: "User's password",
			Type:        "password",
		},
		"full_name": {
			DisplayName: "Common Name",
			Description: "Full name or display name (maps to cn attribute)",
			Type:        "string",
		},
		"first_name": {
			DisplayName: "First Name",
			Description: "User's first/given name",
			Type:        "string",
		},
		"last_name": {
			DisplayName: "Last Name",
			Description: "User's last/family name (maps to sn attribute)",
			Type:        "string",
		},
		"emails": {
			DisplayName: "Email Address",
			Description: "Primary email address",
			Type:        "email[]",
		},
		"groups": {
			DisplayName: "Groups",
			Description: "Groups the user should be added to",
			Type:        "array",
		},
	}
}

// ValidateUserData validates the userDetails struct contains all the required fields for new users exist.
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
		return ErrUsernameIsRequired
	}

	if userData.GetFamilyName() == "" {
		return ErrLastNameIsRequired
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

// ValidatePartialUpdate validates data for partial updates (PATCH with field mask).
func (r *RFC2307bisUserManagement) ValidatePartialUpdate(userData *UserDetailsExtended, updateMask []string) error {
	if userData == nil {
		return fmt.Errorf("user data cannot be nil")
	}

	maskSet := make(map[string]bool)
	for _, field := range updateMask {
		maskSet[field] = true
	}

	if maskSet["emails"] && userData.UserDetails != nil && len(userData.GetEmails()) > 0 {
		for _, email := range userData.GetEmails() {
			if !utils.ValidateEmailString(email) {
				return fmt.Errorf("invalid email address: %s", email)
			}
		}
	}

	if maskSet["profile"] && userData.Profile != nil {
	}

	if maskSet["picture"] && userData.Picture != nil {
	}

	if maskSet["website"] && userData.Website != nil {
	}

	return nil
}

//nolint:gocyclo
func (r *RFC2307bisUserManagement) UpdateUserWithMask(username string, userData *UserDetailsExtended, updateMask []string) error {
	if userData == nil || userData.UserDetails == nil {
		return fmt.Errorf("userData and userData.UserDetails cannot be nil")
	}

	var (
		client LDAPExtendedClient
		err    error
	)

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

	for _, field := range updateMask {
		switch {
		case field == "first_name":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.GivenName, userData.GivenName)
		case field == "last_name":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.FamilyName, userData.FamilyName)
		case field == "middle_name":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.MiddleName, userData.MiddleName)
		case field == "nickname":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Nickname, userData.Nickname)
		case field == "gender":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Gender, userData.Gender)
		case field == "birthdate":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Birthdate, userData.Birthdate)
		case field == "zone_info":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.ZoneInfo, userData.ZoneInfo)
		case field == "phone_number":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.PhoneNumber, userData.PhoneNumber)
		case field == "phone_extension":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.PhoneExtension, userData.PhoneExtension)
		case field == "locale":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Locale, userData.GetLocale())
		case field == "profile":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Profile, userData.GetProfile())
		case field == "picture":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Picture, userData.GetPicture())
		case field == "website":
			r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Website, userData.GetWebsite())
		case field == "display_name":
			if userData.GetDisplayName() != "" {
				r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.DisplayName, userData.GetDisplayName())
			}
		case field == "emails":
			//TODO: handle multiple emails, this will require authelia-internal "primary" email tracking. See https://github.com/authelia/authelia/discussions/6093
			if len(userData.Emails) > 0 {
				r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Mail, userData.Emails[0])
			}
		case field == "groups":
			if userData.GetGroups() != nil {
				if err := r.UpdateUserGroups(username, userData.GetGroups()); err != nil {
					return err
				}
			}
		case field == "extra":
			for jsonKey, value := range userData.Extra {
				if value == nil {
					continue
				}

				ldapAttr := r.getLDAPAttributeForExtraField(jsonKey)
				if ldapAttr == "" {
					r.provider.log.Warnf("No LDAP attribute mapping found for extra field '%s', skipping", jsonKey)
					continue
				}

				r.replaceAttributeIfPresent(modifyRequest, ldapAttr, fmt.Sprintf("%v", value))
			}
		case field == "address":
			if userData.Address != nil {
				r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.StreetAddress, userData.Address.StreetAddress)
				r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Locality, userData.Address.Locality)
				r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Region, userData.Address.Region)
				r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.PostalCode, userData.Address.PostalCode)
				r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Country, userData.Address.Country)
			}
		case strings.HasPrefix(field, "address."):
			if userData.Address != nil {
				subField := strings.TrimPrefix(field, "address.")
				switch subField {
				case "street_address":
					r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.StreetAddress, userData.Address.StreetAddress)
				case "locality":
					r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Locality, userData.Address.Locality)
				case "region":
					r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Region, userData.Address.Region)
				case "postal_code":
					r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.PostalCode, userData.Address.PostalCode)
				case "country":
					r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Country, userData.Address.Country)
				}
			}
		}
	}

	if len(modifyRequest.Changes) == 0 {
		r.provider.log.Debugf("No changes detected for user '%s', skipping update", username)
		return nil
	}

	if err = r.provider.modify(client, modifyRequest); err != nil {
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

//nolint:gocyclo
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

	var client LDAPExtendedClient
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

	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.GivenName, userData.GivenName)
	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.FamilyName, userData.FamilyName)

	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.MiddleName, userData.MiddleName)
	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Nickname, userData.Nickname)
	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Gender, userData.Gender)
	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Birthdate, userData.Birthdate)
	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.ZoneInfo, userData.ZoneInfo)
	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.PhoneNumber, userData.PhoneNumber)
	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.PhoneExtension, userData.PhoneExtension)

	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Locale, userData.GetLocale())
	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Profile, userData.GetProfile())
	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Picture, userData.GetPicture())
	r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Website, userData.GetWebsite())

	if userData.Address != nil {
		r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.StreetAddress, userData.Address.StreetAddress)
		r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Locality, userData.Address.Locality)
		r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Region, userData.Address.Region)
		r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.PostalCode, userData.Address.PostalCode)
		r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Country, userData.Address.Country)
	}

	if userData.GetGroups() != nil {
		err := r.UpdateUserGroups(username, userData.GetGroups())
		if err != nil {
			return err
		}
	}

	if userData.GetDisplayName() != "" {
		r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.DisplayName, userData.GetDisplayName())
	}

	if len(userData.Emails) > 0 {
		r.replaceAttributeIfPresent(modifyRequest, r.provider.config.Attributes.Mail, userData.Emails[0])
	}

	for ldapAttr, value := range userData.Extra {
		if value != nil {
			r.replaceAttributeIfPresent(modifyRequest, ldapAttr, fmt.Sprintf("%v", value))
		}
	}

	if len(modifyRequest.Changes) == 0 {
		r.provider.log.Debugf("No changes detected for user '%s', skipping update", username)
		return nil
	}

	r.provider.log.Debugf("Sending modify request for user '%s' with %d changes:", username, len(modifyRequest.Changes))

	if err = r.provider.modify(client, modifyRequest); err != nil {
		return fmt.Errorf("unable to update user '%s': %w", username, err)
	}

	r.provider.log.Infof("Successfully updated user '%s'", username)

	return nil
}

//nolint:gocyclo
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

	var client LDAPExtendedClient
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

	// Required Attributes.
	addRequest.Attribute(ldapAttrObjectClass, r.GetDefaultObjectClasses())
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Username, userData.GetUsername())
	r.addAttributeIfPresent(addRequest, ldapAttrCommonName, userData.CommonName)
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.FamilyName, userData.FamilyName)
	r.addAttributeIfPresent(addRequest, ldapAttributeUserPassword, userData.Password)

	// Optional attributes.
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Nickname, userData.Nickname)
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.MiddleName, userData.MiddleName)
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Gender, userData.Gender)
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Birthdate, userData.Birthdate)
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.ZoneInfo, userData.ZoneInfo)
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.PhoneNumber, userData.PhoneNumber)
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.PhoneExtension, userData.PhoneExtension)

	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Locale, userData.GetLocale())
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Profile, userData.GetProfile())
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Picture, userData.GetPicture())
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Website, userData.GetWebsite())

	if userData.Address != nil {
		r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.StreetAddress, userData.Address.StreetAddress)
		r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Locality, userData.Address.Locality)
		r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Region, userData.Address.Region)
		r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.PostalCode, userData.Address.PostalCode)
		r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Country, userData.Address.Country)
	}

	if len(userData.Emails) > 0 {
		addRequest.Attribute(r.provider.config.Attributes.Mail, []string{userData.Emails[0]})
	}

	if userData.GivenName != "" {
		addRequest.Attribute(r.provider.config.Attributes.GivenName, []string{userData.GivenName})
	}

	// Attempt to build displayName from other attributes.
	if userData.GetDisplayName() == "" {
		//nolint:gocritic
		if userData.GetGivenName() != "" && userData.GetFamilyName() != "" {
			userData.DisplayName = fmt.Sprintf("%s %s", userData.GetGivenName(), userData.GetFamilyName())
		} else if userData.GetGivenName() != "" {
			userData.DisplayName = userData.GetGivenName()
		} else if userData.GetFamilyName() != "" {
			userData.DisplayName = userData.GetFamilyName()
		}
	}

	if userData.GetDisplayName() != "" {
		r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.DisplayName, userData.GetDisplayName())
	}

	if err = client.Add(addRequest); err != nil {
		return fmt.Errorf("failed to add user '%s': %w", userData.Username, err)
	}

	return nil
}

func (r *RFC2307bisUserManagement) DeleteUser(username string) (err error) {
	var client LDAPExtendedClient
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

func (p *LDAPUserProvider) getReferral(err error) (referral string, ok bool) {
	if !p.config.PermitReferrals {
		return "", false
	}

	return ldapGetReferral(err)
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

func (r *RFC2307bisUserManagement) replaceAttributeIfPresent(req *ldap.ModifyRequest, ldapAttr, value string) {
	if ldapAttr == "" {
		return
	}

	if value == "" {
		req.Delete(ldapAttr, []string{})
	} else {
		req.Replace(ldapAttr, []string{value})
	}
}

func (r *RFC2307bisUserManagement) addAttributeIfPresent(req *ldap.AddRequest, ldapAttr, value string) {
	if ldapAttr == "" {
		return
	}

	if value == "" {
		return
	}

	req.Attribute(ldapAttr, []string{value})
}

//nolint:gocyclo
func (r *RFC2307bisUserManagement) UpdateUserGroups(username string, groups []string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	var (
		client LDAPExtendedClient
		err    error
	)
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

func (r *RFC2307bisUserManagement) GetGroups(client LDAPExtendedClient) ([]*ldap.Entry, error) {
	searchRequest := ldap.NewSearchRequest(
		r.provider.groupsBaseDN,
		ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(|(objectClass=groupOfNames)(objectClass=groupOfUniqueNames)(objectClass=posixGroup))",
		[]string{"cn", "member", "uniqueMember", "gidNumber"},
		nil,
	)

	searchResult, err := r.provider.search(client, searchRequest)
	if err != nil {
		var ldapErr *ldap.Error
		if errors.As(err, &ldapErr) && ldapErr.ResultCode == ldap.LDAPResultNoSuchObject {
			return []*ldap.Entry{}, nil
		}

		return nil, fmt.Errorf("error occurred searching for all groups: %w", err)
	}

	return searchResult.Entries, nil
}

// CreateGroup creates a new group in LDAP.
func (r *RFC2307bisUserManagement) CreateGroup(client LDAPExtendedClient, groupName, groupDN string) error {
	addRequest := ldap.NewAddRequest(groupDN, nil)

	// RFC2307bis group object classes.
	addRequest.Attribute("objectClass", []string{
		"top",
		"groupOfNames",
	})

	addRequest.Attribute(r.provider.config.Attributes.GroupName, []string{groupName})
	// placeholderDN := fmt.Sprintf("cn=placeholder,%s", r.provider.groupsBaseDN).
	addRequest.Attribute(r.provider.config.Attributes.GroupMember, []string{groupDN})

	if err := client.Add(addRequest); err != nil {
		return err
	}

	return nil
}

// DeleteGroup deletes a group in LDAP.
//

func (r *RFC2307bisUserManagement) DeleteGroup(client LDAPExtendedClient, groupName, groupDN string) error {
	// Check if group exists first.
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

func (r *RFC2307bisUserManagement) getGroupObject(client LDAPExtendedClient, groupDN string) (*ldap.Entry, error) {
	searchRequest := ldap.NewSearchRequest(
		groupDN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		1, 0, false,
		"(|(objectClass=groupOfNames)(objectClass=groupOfUniqueNames)(objectClass=posixGroup))",
		[]string{"cn", "member", "uniqueMember", "gidNumber"},
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

	//TODO: make sure the first element is the proper group; **somehow**.
	return searchResult.Entries[0], nil
}

func (r *RFC2307bisUserManagement) getLDAPAttributeForExtraField(jsonKey string) string {
	for ldapAttr, config := range r.provider.config.Attributes.Extra {
		if config.Name != "" && config.Name == jsonKey {
			return ldapAttr
		}

		if config.Name == "" && ldapAttr == jsonKey {
			return ldapAttr
		}
	}

	return ""
}

// groupExists checks if a group exists in LDAP.
func (r *RFC2307bisUserManagement) groupExists(client LDAPExtendedClient, groupDN string) (bool, error) {
	searchRequest := ldap.NewSearchRequest(
		groupDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		1, 0, false,
		"(|(objectClass=groupOfNames)(objectClass=groupOfUniqueNames)(objectClass=posixGroup))",
		[]string{"dn"},
		nil,
	)

	searchResult, err := client.Search(searchRequest)
	if err != nil {
		// Check if it's a "No Such Object" error.
		var ldapErr *ldap.Error
		if errors.As(err, &ldapErr) && ldapErr.ResultCode == ldap.LDAPResultNoSuchObject {
			return false, nil
		}

		return false, err
	}

	return len(searchResult.Entries) > 0, nil
}

// isUserMemberOfGroup checks if a user is already a member of a group.
func (r *RFC2307bisUserManagement) isUserMemberOfGroup(client LDAPExtendedClient, userDN, groupDN string) (bool, error) {
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

// removeUserFromGroup removes a user from a group.
func (r *RFC2307bisUserManagement) removeUserFromGroup(client LDAPExtendedClient, userDN, groupName string) error {
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

// addUserToGroup adds a user to a group, creating the group if it doesn't exist.
// TODO: Remove this.
func (r *RFC2307bisUserManagement) addUserToGroup(client LDAPExtendedClient, userDN, username, groupName string) error {
	groupDN := fmt.Sprintf("%s=%s,%s",
		r.provider.config.Attributes.GroupName,
		ldap.EscapeFilter(groupName),
		r.provider.groupsBaseDN)

	exists, err := r.groupExists(client, groupDN)
	if err != nil {
		return fmt.Errorf("failed to check if group '%s' exists: %w", groupName, err)
	}

	if !exists {
		if err := r.CreateGroup(client, groupName, groupDN); err != nil {
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

// getCurrentUserGroups retrieves all groups that the user currently belongs to.
func (r *RFC2307bisUserManagement) getCurrentUserGroups(client LDAPExtendedClient, userDN string) ([]string, error) {
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
