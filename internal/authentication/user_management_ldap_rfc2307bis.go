package authentication

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"

	"github.com/authelia/authelia/v4/internal/utils"
)

type RFC2307bisUserManagement struct {
	provider *LDAPUserProvider
}

var attributeMetadataMap = map[string]UserManagementAttributeMetadata{
	"username":        {Type: Text, Multiple: false},
	"groups":          {Type: Groups, Multiple: true},
	"password":        {Type: Password, Multiple: false},
	"display_name":    {Type: Text, Multiple: false},
	"family_name":     {Type: Text, Multiple: false},
	"given_name":      {Type: Text, Multiple: false},
	"middle_name":     {Type: Text, Multiple: false},
	"nickname":        {Type: Text, Multiple: false},
	"gender":          {Type: Text, Multiple: false},
	"birthdate":       {Type: Date, Multiple: false},
	"website":         {Type: Url, Multiple: false},
	"profile":         {Type: Url, Multiple: false},
	"picture":         {Type: Url, Multiple: false},
	"zoneinfo":        {Type: Text, Multiple: false},
	"locale":          {Type: Text, Multiple: false},
	"phone_number":    {Type: Telephone, Multiple: false},
	"phone_extension": {Type: Text, Multiple: false},
	"street_address":  {Type: Text, Multiple: false},
	"locality":        {Type: Text, Multiple: false},
	"region":          {Type: Text, Multiple: false},
	"postal_code":     {Type: Text, Multiple: false},
	"country":         {Type: Text, Multiple: false},
	"mail":            {Type: Email, Multiple: false},
}

func (r *RFC2307bisUserManagement) GetRequiredAttributes() []string {
	requiredFieldNames := GetBaseRequiredAttributesForImplementation(schema.LDAPImplementationRFC2307bis)

	return append(requiredFieldNames, r.provider.config.UserManagement.RequiredAttributes...)
}

func (r *RFC2307bisUserManagement) GetSupportedAttributes() map[string]UserManagementAttributeMetadata {
	blocklist := map[string]bool{
		"group_member":       true,
		"group_name":         true,
		"member_of":          true,
		"distinguished_name": true,
	}

	fieldNames := getFieldNames(r.provider.config.Attributes)

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

	for key, extraAttr := range r.provider.config.Attributes.Extra {
		attrName := extraAttr.Name
		if attrName == "" {
			attrName = key
		}

		var inputType AttributeType

		switch extraAttr.ValueType {
		case "boolean":
			inputType = Checkbox
		case "integer", "string", "":
			inputType = Text
		default:
			inputType = Text
		}

		metadata["extra."+attrName] = UserManagementAttributeMetadata{
			Type:     inputType,
			Multiple: extraAttr.MultiValued,
		}
	}

	return metadata
}

// GetDefaultUserObjectClasses returns the default object classes for users.
func (r *RFC2307bisUserManagement) GetDefaultUserObjectClasses() []string {
	return r.provider.config.UserManagement.UserObjectClasses
}

// GetDefaultGroupObjectClasses returns the default object classes for groups.
func (r *RFC2307bisUserManagement) GetDefaultGroupObjectClasses() []string {
	return r.provider.config.UserManagement.GroupObjectClasses
}

// ValidateUserData validates the userDetails struct contains all the required fields for new users.
func (r *RFC2307bisUserManagement) ValidateUserData(userData *UserDetailsExtended) error {
	//TODO: implement more verbose errors to enable frontend to show helpful errors.
	if userData == nil {
		return fmt.Errorf("user data cannot be nil")
	}

	requiredAttributes := r.GetRequiredAttributes()

	attributeValues := r.buildAttributeValueMap(userData)

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
func (r *RFC2307bisUserManagement) buildAttributeValueMap(userData *UserDetailsExtended) map[string]interface{} {
	values := make(map[string]interface{})

	values["username"] = userData.GetUsername()
	values["password"] = userData.Password
	values["display_name"] = userData.GetDisplayName()
	values["given_name"] = userData.GetGivenName()
	values["family_name"] = userData.GetFamilyName()
	values["middle_name"] = userData.MiddleName
	values["nickname"] = userData.Nickname
	values["common_name"] = userData.CommonName

	values["profile"] = userData.GetProfile()
	values["picture"] = userData.GetPicture()
	values["website"] = userData.GetWebsite()
	values["gender"] = userData.Gender
	values["birthdate"] = userData.Birthdate
	values["locale"] = userData.GetLocale()
	values["zone_info"] = userData.ZoneInfo

	values["phone_number"] = userData.PhoneNumber
	values["phone_extension"] = userData.PhoneExtension

	if len(userData.GetEmails()) > 0 {
		values["mail"] = userData.GetEmails()[0]
		values["emails"] = userData.GetEmails()
	}

	if userData.Address != nil {
		values["street_address"] = userData.Address.StreetAddress
		values["locality"] = userData.Address.Locality
		values["region"] = userData.Address.Region
		values["postal_code"] = userData.Address.PostalCode
		values["country"] = userData.Address.Country
	}

	// Add extra attributes.
	if userData.Extra != nil {
		for ldapAttr, attrConfig := range r.provider.config.Attributes.Extra {
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
		case strings.HasPrefix(field, "extra."):
			extraField := strings.TrimPrefix(field, "extra.")

			if userData.Extra != nil {
				if value, exists := userData.Extra[extraField]; exists && value != nil {
					ldapAttr := r.getLDAPAttributeForExtraField(extraField)
					if ldapAttr == "" {
						r.provider.log.Warnf("No LDAP attribute mapping found for extra field '%s', skipping", extraField)
						continue
					}

					r.replaceAttributeIfPresent(modifyRequest, ldapAttr, r.normalizeExtraAttributes(value))
				}
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

	userDN, err := r.provider.BuildUserDN(userData)
	if err != nil {
		return fmt.Errorf("unable to build DN for user '%s': %w", userData.Username, err)
	}

	addRequest := ldap.NewAddRequest(userDN, nil)

	addRequest.Attribute(ldapAttrObjectClass, r.GetDefaultUserObjectClasses())

	if r.provider.config.UserManagement.CreatedUsersRDNFormat != "" &&
		r.provider.config.UserManagement.CreatedUsersRDNAttribute != r.provider.config.Attributes.Username {
		r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.Username, userData.GetUsername())
	}

	r.addAttributeIfPresent(addRequest, ldapAttrCommonName, userData.CommonName)
	r.addAttributeIfPresent(addRequest, r.provider.config.Attributes.FamilyName, userData.FamilyName)
	r.addAttributeIfPresent(addRequest, ldapAttributeUserPassword, userData.Password)

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

	if userData.Extra != nil {
		for jsonKey, value := range userData.Extra {
			if value == nil {
				continue
			}

			ldapAttr := r.getLDAPAttributeForExtraField(jsonKey)
			if ldapAttr == "" {
				r.provider.log.Warnf("No LDAP attribute mapping found for extra field '%s', skipping", jsonKey)
				continue
			}

			r.addAttributeIfPresent(addRequest, ldapAttr, r.normalizeExtraAttributes(value))
		}
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

	if len(userData.Groups) > 0 {
		if err = r.UpdateUserGroups(userData.Username, userData.Groups); err != nil {
			return fmt.Errorf("failed to assign groups for user '%s': %w", userData.Username, err)
		}
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

	for _, groupName := range groupsToAdd {
		exists, err := r.groupExists(client, groupName)
		if err != nil {
			return fmt.Errorf("failed to check group existence for group '%s'", groupName)
		}

		if !exists {
			return fmt.Errorf("group '%s' does not exist: %w", groupName, ErrGroupNotFound)
		}
	}

	r.provider.log.Debugf("Group update for user '%s': adding %d groups, removing %d groups",
		username, len(groupsToAdd), len(groupsToRemove))

	var errs []error

	for _, groupName := range groupsToRemove {
		if err := r.removeUserFromGroup(client, userDN, groupName); err != nil {
			r.provider.log.WithError(err).Errorf("Failed to remove user '%s' from group '%s'", username, groupName)
			errs = append(errs, err)
		}
	}

	for _, groupName := range groupsToAdd {
		if err := r.addUserToGroup(client, groupName, userDN); err != nil {
			r.provider.log.WithError(err).Errorf("Failed to add user '%s' in group '%s'", username, groupName)
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// ListGroups returns a list of all group names.
func (r *RFC2307bisUserManagement) ListGroups() ([]string, error) {
	var (
		client LDAPExtendedClient
		err    error
	)
	if client, err = r.provider.factory.GetClient(); err != nil {
		return nil, fmt.Errorf("unable to get LDAP client for listing groups: %w", err)
	}

	defer func() {
		if err := r.provider.factory.ReleaseClient(client); err != nil {
			r.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	searchRequest := ldap.NewSearchRequest(
		r.provider.groupsBaseDN,
		ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(|(objectClass=groupOfNames)(objectClass=groupOfUniqueNames)(objectClass=posixGroup))",
		[]string{r.provider.config.Attributes.GroupName},
		nil,
	)

	searchResult, err := r.provider.search(client, searchRequest)
	if err != nil {
		var ldapErr *ldap.Error
		if errors.As(err, &ldapErr) && ldapErr.ResultCode == ldap.LDAPResultNoSuchObject {
			return []string{}, nil
		}

		return nil, fmt.Errorf("error occurred searching for all groups: %w", err)
	}

	groups := make([]string, 0, len(searchResult.Entries))
	for _, entry := range searchResult.Entries {
		groupName := entry.GetAttributeValue(r.provider.config.Attributes.GroupName)
		if groupName != "" {
			groups = append(groups, groupName)
		}
	}

	return groups, nil
}

func (r *RFC2307bisUserManagement) GetGroups() ([]*ldap.Entry, error) {
	var (
		client LDAPExtendedClient
		err    error
	)
	if client, err = r.provider.factory.GetClient(); err != nil {
		return nil, fmt.Errorf("unable to get LDAP client for group update: %w", err)
	}

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

func (p *LDAPUserProvider) BuildGroupDN(groupName string) string {
	baseDN := p.groupsBaseDN
	if p.config.UserManagement.CreatedGroupsDN != "" {
		baseDN = p.config.UserManagement.CreatedGroupsDN + "," + p.groupsBaseDN
	}

	rdn := fmt.Sprintf("%s=%s", p.config.Attributes.GroupName, ldap.EscapeFilter(groupName))

	return fmt.Sprintf("%s,%s", rdn, baseDN)
}

// AddGroup creates a new group in LDAP.
func (r *RFC2307bisUserManagement) AddGroup(groupName string) error {
	var (
		client LDAPExtendedClient
		err    error
	)
	if client, err = r.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to get LDAP client for group creation: %w", err)
	}

	defer func() {
		if err := r.provider.factory.ReleaseClient(client); err != nil {
			r.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	groupDN := r.provider.BuildGroupDN(groupName)

	// Check if group already exists.
	exists, err := r.groupExists(client, groupName)
	if err != nil {
		return fmt.Errorf("failed to check if group '%s' exists: %w", groupName, err)
	}

	if exists {
		return fmt.Errorf("error creating group '%s': %w", groupName, ErrGroupExists)
	}

	if err := r.createGroupInternal(client, groupName, groupDN); err != nil {
		return err
	}

	r.provider.log.Infof("Successfully created group '%s'", groupName)

	return nil
}

// createGroupInternal creates a group in LDAP using an existing client connection.
func (r *RFC2307bisUserManagement) createGroupInternal(client LDAPExtendedClient, groupName, groupDN string) error {
	addRequest := ldap.NewAddRequest(groupDN, nil)

	addRequest.Attribute("objectClass", r.GetDefaultGroupObjectClasses())
	addRequest.Attribute(r.provider.config.Attributes.GroupName, []string{groupName})

	// groupOfNames requires at least one member, so we add a placeholder.
	placeholderDN := fmt.Sprintf("cn=placeholder,%s", r.provider.config.BaseDN)
	addRequest.Attribute(r.provider.config.Attributes.GroupMember, []string{placeholderDN})

	if err := client.Add(addRequest); err != nil {
		if getLDAPResultCode(err) == ldap.LDAPResultEntryAlreadyExists {
			return fmt.Errorf("error creating group '%s': %w", groupName, ErrGroupExists)
		}

		return fmt.Errorf("failed to create group '%s': %w", groupName, err)
	}

	return nil
}

// DeleteGroup deletes a group in LDAP.
func (r *RFC2307bisUserManagement) DeleteGroup(groupName string) error {
	var (
		client LDAPExtendedClient
		err    error
	)
	if client, err = r.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to get LDAP client for group deletion: %w", err)
	}

	defer func() {
		if err := r.provider.factory.ReleaseClient(client); err != nil {
			r.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	groupDN := fmt.Sprintf("%s=%s,%s", r.provider.config.Attributes.GroupName, ldap.EscapeFilter(groupName), r.provider.groupsBaseDN)

	// Check if group exists first.
	exists, err := r.groupExists(client, groupName)
	if err != nil {
		return fmt.Errorf("failed to check if group '%s' exists: %w", groupName, err)
	}

	if !exists {
		r.provider.log.Debugf("Group '%s' doesn't exist, nothing to delete", groupName)
		return ErrGroupNotFound
	}

	deleteRequest := ldap.NewDelRequest(groupDN, nil)

	if err := client.Del(deleteRequest); err != nil {
		return fmt.Errorf("failed to delete group '%s': %w", groupName, err)
	}

	r.provider.log.Infof("Successfully deleted group '%s'", groupName)

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
func (r *RFC2307bisUserManagement) groupExists(client LDAPExtendedClient, groupName string) (bool, error) {
	groupDN := fmt.Sprintf("%s=%s,%s",
		r.provider.config.Attributes.GroupName,
		ldap.EscapeFilter(groupName),
		r.provider.groupsBaseDN)

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

// addUserToGroup adds a user to a group.
func (r *RFC2307bisUserManagement) addUserToGroup(client LDAPExtendedClient, groupName, userDn string) error {
	groupDN := fmt.Sprintf("%s=%s,%s",
		r.provider.config.Attributes.GroupName,
		ldap.EscapeFilter(groupName),
		r.provider.groupsBaseDN)

	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	modifyRequest.Add(r.provider.config.Attributes.GroupMember, []string{userDn})

	if err := client.Modify(modifyRequest); err != nil {
		return fmt.Errorf("failed to add user '%s' to group '%s': %w", userDn, groupName, err)
	}

	return nil
}

// removeUserFromGroup removes a user from a group.
func (r *RFC2307bisUserManagement) removeUserFromGroup(client LDAPExtendedClient, userDN, groupName string) error {
	groupDN := fmt.Sprintf("%s=%s,%s",
		r.provider.config.Attributes.GroupName,
		ldap.EscapeFilter(groupName),
		r.provider.groupsBaseDN)

	exists, err := r.groupExists(client, groupName)
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

func (r *RFC2307bisUserManagement) normalizeExtraAttributes(value any) string {
	var ldapValue string

	switch v := value.(type) {
	case bool:
		if v {
			ldapValue = "TRUE"
		} else {
			ldapValue = "FALSE"
		}
	default:
		ldapValue = fmt.Sprintf("%v", value)
	}

	return ldapValue
}
