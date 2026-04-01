package authentication

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-crypt/crypt/algorithm"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

type FileUserManagement struct {
	provider *FileUserProvider
}

func (f *FileUserManagement) GetRequiredAttributes() []string {
	requiredFieldNames := GetBaseRequiredAttributesForImplementation(schema.FileImplementation)

	if f.provider.config.UserManagement.RequiredAttributes != nil {
		return append(requiredFieldNames, f.provider.config.UserManagement.RequiredAttributes...)
	}

	return requiredFieldNames
}

func (f *FileUserManagement) GetSupportedAttributes() map[string]UserManagementAttributeMetadata {
	blocklist := map[string]bool{
		"group_member":       true,
		"group_name":         true,
		"member_of":          true,
		"distinguished_name": true,
	}

	metadata := make(map[string]UserManagementAttributeMetadata)

	for fieldName, meta := range attributeMetadataMap {
		if blocklist[fieldName] {
			continue
		}

		if fieldName == AttributeGroups {
			metadata[fieldName] = UserManagementAttributeMetadata{
				Type:     Text,
				Multiple: true,
			}
		} else {
			metadata[fieldName] = meta
		}
	}

	if f.provider.config.ExtraAttributes != nil {
		for attrName, attrConfig := range f.provider.config.ExtraAttributes {
			var inputType AttributeType

			switch attrConfig.ValueType {
			case ValueTypeBoolean:
				inputType = Checkbox
			case ValueTypeInteger, ValueTypeString, "":
				inputType = Text
			default:
				inputType = Text
			}

			metadata[PrefixAttributeExtra+attrName] = UserManagementAttributeMetadata{
				Type:     inputType,
				Multiple: attrConfig.MultiValued,
			}
		}
	}

	return metadata
}

// ValidateUserData validates the userDetails struct contains all the required fields for new users.
func (f *FileUserManagement) ValidateUserData(userData *UserDetailsExtended) error {
	if userData == nil {
		return fmt.Errorf("user data cannot be nil")
	}

	if userData.UserDetails == nil {
		return fmt.Errorf("user details cannot be nil")
	}

	requiredAttributes := f.GetRequiredAttributes()
	attributeValues := f.buildAttributeValueMap(userData)

	var missingAttributes []string

	for _, attr := range requiredAttributes {
		value, exists := attributeValues[attr]
		if !exists || utils.IsEmptyValue(value) {
			if attr == AttributeDisplayName {
				givenName := attributeValues[AttributeGivenName]

				familyName := attributeValues[AttributeFamilyName]
				if !utils.IsEmptyValue(givenName) || !utils.IsEmptyValue(familyName) {
					continue
				}
			}

			missingAttributes = append(missingAttributes, attr)
		}
	}

	if len(missingAttributes) > 0 {
		return fmt.Errorf("missing required attributes: %s", strings.Join(missingAttributes, ", "))
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
func (f *FileUserManagement) buildAttributeValueMap(userData *UserDetailsExtended) map[string]interface{} {
	values := make(map[string]interface{})

	values[AttributeUsername] = userData.GetUsername()
	values[AttributePassword] = userData.Password
	values[AttributeDisplayName] = userData.GetDisplayName()
	values[AttributeGivenName] = userData.GetGivenName()
	values[AttributeFamilyName] = userData.GetFamilyName()
	values[AttributeMiddleName] = userData.MiddleName
	values[AttributeNickname] = userData.Nickname

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

	// Add extra attributes.
	if userData.Extra != nil {
		for attrName, value := range userData.Extra {
			if value != nil {
				values[PrefixAttributeExtra+attrName] = value
			}
		}
	}

	return values
}

// ValidatePartialUpdate validates data for partial updates (PATCH with field mask).
func (f *FileUserManagement) ValidatePartialUpdate(userData *UserDetailsExtended, updateMask []string) error {
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

//nolint:gocyclo
func (f *FileUserManagement) UpdateUserWithMask(username string, userData *UserDetailsExtended, updateMask []string) error {
	if userData == nil || userData.UserDetails == nil {
		return fmt.Errorf("userData and userData.UserDetails cannot be nil")
	}

	existingDetails, err := f.provider.database.GetUserDetails(username)
	if err != nil {
		return fmt.Errorf("unable to retrieve user for update of user '%s': %w", username, err)
	}

	if existingDetails.Disabled {
		return fmt.Errorf("cannot update disabled user '%s'", username)
	}

	updatedDetails := existingDetails

	for _, field := range updateMask {
		switch {
		case field == AttributeGivenName:
			updatedDetails.GivenName = userData.GivenName
		case field == AttributeFamilyName:
			updatedDetails.FamilyName = userData.FamilyName
		case field == AttributeMiddleName:
			updatedDetails.MiddleName = userData.MiddleName
		case field == AttributeNickname:
			updatedDetails.Nickname = userData.Nickname
		case field == AttributeGender:
			updatedDetails.Gender = userData.Gender
		case field == AttributeBirthdate:
			updatedDetails.Birthdate = userData.Birthdate
		case field == AttributeZoneInfo:
			updatedDetails.ZoneInfo = userData.ZoneInfo
		case field == AttributePhoneNumber:
			updatedDetails.PhoneNumber = userData.PhoneNumber
		case field == AttributePhoneExtension:
			updatedDetails.PhoneExtension = userData.PhoneExtension
		case field == AttributeLocale:
			updatedDetails.Locale = userData.Locale
		case field == AttributeProfile:
			updatedDetails.Profile = userData.Profile
		case field == AttributePicture:
			updatedDetails.Picture = userData.Picture
		case field == AttributeWebsite:
			updatedDetails.Website = userData.Website
		case field == AttributeDisplayName:
			if userData.GetDisplayName() != "" {
				updatedDetails.DisplayName = userData.GetDisplayName()
			}
		case field == AttributeMail:
			if len(userData.Emails) > 0 {
				updatedDetails.Email = userData.Emails[0]
			}
		case field == AttributeGroups:
			if userData.GetGroups() != nil {
				updatedDetails.Groups = userData.GetGroups()
			}
		case strings.HasPrefix(field, PrefixAttributeExtra):
			extraField := strings.TrimPrefix(field, PrefixAttributeExtra)

			if userData.Extra != nil {
				if value, exists := userData.Extra[extraField]; exists {
					// Skip empty strings.
					if strValue, ok := value.(string); ok && strValue == "" {
						// Remove the field if it's an empty string.
						if updatedDetails.Extra != nil {
							delete(updatedDetails.Extra, extraField)
						}
					} else {
						if updatedDetails.Extra == nil {
							updatedDetails.Extra = make(map[string]any)
						}

						// Convert the value to the proper type based on configuration.
						convertedValue, err := f.convertExtraAttributeValue(extraField, value)
						if err != nil {
							return fmt.Errorf("failed to convert extra attribute '%s' for user '%s': %w", extraField, username, err)
						}

						updatedDetails.Extra[extraField] = convertedValue
					}
				}
			}
		case field == AttributeAddress:
			if userData.Address != nil {
				if updatedDetails.Address == nil {
					updatedDetails.Address = &FileUserDatabaseUserDetailsAddressModel{}
				}

				updatedDetails.Address.StreetAddress = userData.Address.StreetAddress
				updatedDetails.Address.Locality = userData.Address.Locality
				updatedDetails.Address.Region = userData.Address.Region
				updatedDetails.Address.PostalCode = userData.Address.PostalCode
				updatedDetails.Address.Country = userData.Address.Country
			}
		case strings.HasPrefix(field, PrefixAttributeAddress):
			if userData.Address != nil {
				if updatedDetails.Address == nil {
					updatedDetails.Address = &FileUserDatabaseUserDetailsAddressModel{}
				}

				subField := strings.TrimPrefix(field, PrefixAttributeAddress)
				switch subField {
				case AttributeAddressStreetAddress:
					updatedDetails.Address.StreetAddress = userData.Address.StreetAddress
				case AttributeAddressLocality:
					updatedDetails.Address.Locality = userData.Address.Locality
				case AttributeAddressRegion:
					updatedDetails.Address.Region = userData.Address.Region
				case AttributeAddressPostalCode:
					updatedDetails.Address.PostalCode = userData.Address.PostalCode
				case AttributeAddressCountry:
					updatedDetails.Address.Country = userData.Address.Country
				}
			}
		}
	}

	// Save updated details.
	f.provider.database.SetUserDetails(username, &updatedDetails)

	f.provider.mutex.Lock()
	f.provider.setTimeoutReload(f.provider.timeoutReload)
	f.provider.mutex.Unlock()

	if err := f.provider.database.Save(); err != nil {
		return fmt.Errorf("unable to save user '%s': %w", username, err)
	}

	return nil
}

//nolint:gocyclo
func (f *FileUserManagement) AddUser(userData *UserDetailsExtended) (err error) {
	if userData == nil || userData.UserDetails == nil {
		return fmt.Errorf("userData and userData.UserDetails cannot be nil")
	}

	if err = f.ValidateUserData(userData); err != nil {
		return fmt.Errorf("validation failed for user '%s': %w", userData.Username, err)
	}

	if userData.Password == "" {
		return fmt.Errorf("password is required to create user '%s'", userData.Username)
	}

	var digest algorithm.Digest
	if digest, err = f.provider.hash.Hash(userData.Password); err != nil {
		return fmt.Errorf("failed to hash password for user '%s': %w", userData.Username, err)
	}

	var email string
	if len(userData.Emails) > 0 {
		email = userData.Emails[0]
	}

	var groups []string
	if len(userData.Groups) > 0 {
		groups = userData.Groups
	}

	displayName := userData.GetDisplayName()
	if displayName == "" {
		//nolint:gocritic
		if userData.GetGivenName() != "" && userData.GetFamilyName() != "" {
			displayName = fmt.Sprintf("%s %s", userData.GetGivenName(), userData.GetFamilyName())
		} else if userData.GetGivenName() != "" {
			displayName = userData.GetGivenName()
		} else if userData.GetFamilyName() != "" {
			displayName = userData.GetFamilyName()
		}
	}

	details := FileUserDatabaseUserDetails{
		Username:       userData.Username,
		DisplayName:    displayName,
		Password:       schema.NewPasswordDigest(digest),
		Email:          email,
		Groups:         groups,
		GivenName:      userData.GivenName,
		FamilyName:     userData.FamilyName,
		MiddleName:     userData.MiddleName,
		Nickname:       userData.Nickname,
		Gender:         userData.Gender,
		Birthdate:      userData.Birthdate,
		Website:        userData.Website,
		Profile:        userData.Profile,
		Picture:        userData.Picture,
		ZoneInfo:       userData.ZoneInfo,
		Locale:         userData.Locale,
		PhoneNumber:    userData.PhoneNumber,
		PhoneExtension: userData.PhoneExtension,
		Disabled:       false,
	}

	// Add address if provided.
	if userData.Address != nil {
		details.Address = &FileUserDatabaseUserDetailsAddressModel{
			StreetAddress: userData.Address.StreetAddress,
			Locality:      userData.Address.Locality,
			Region:        userData.Address.Region,
			PostalCode:    userData.Address.PostalCode,
			Country:       userData.Address.Country,
		}
	}

	if userData.Extra != nil {
		for key, value := range userData.Extra {
			if value != nil {
				if strValue, ok := value.(string); ok && strValue == "" {
					continue
				}

				if details.Extra == nil {
					details.Extra = make(map[string]any)
				}

				// Convert the value to the proper type based on configuration.
				convertedValue, err := f.convertExtraAttributeValue(key, value)
				if err != nil {
					return fmt.Errorf("failed to convert extra attribute '%s' for user '%s': %w", key, userData.Username, err)
				}

				details.Extra[key] = convertedValue
			}
		}
	}

	f.provider.database.SetUserDetails(details.Username, &details)

	f.provider.mutex.Lock()
	f.provider.setTimeoutReload(f.provider.timeoutReload)
	f.provider.mutex.Unlock()

	if err = f.provider.database.Save(); err != nil {
		return fmt.Errorf("failed to save user '%s': %w", userData.Username, err)
	}

	return nil
}

func (f *FileUserManagement) DeleteUser(username string) (err error) {
	// Check if user exists.
	_, err = f.provider.database.GetUserDetails(username)
	if err != nil {
		return fmt.Errorf("unable to retrieve user for deletion of user '%s': %w", username, err)
	}

	f.provider.database.DeleteUserDetails(username)

	f.provider.mutex.Lock()
	f.provider.setTimeoutReload(f.provider.timeoutReload)
	f.provider.mutex.Unlock()

	if err = f.provider.database.Save(); err != nil {
		return fmt.Errorf("unable to delete user '%s': %w", username, err)
	}

	f.provider.log.Debugf("User '%s' was successfully deleted.", username)

	return nil
}

// ListGroups returns a list of all group names from all users.
func (f *FileUserManagement) ListGroups() ([]string, error) {
	allUsers := f.provider.database.GetAllUsers()

	groupSet := make(map[string]bool)

	for _, user := range allUsers {
		for _, group := range user.Groups {
			if group != "" {
				groupSet[group] = true
			}
		}
	}

	groups := make([]string, 0, len(groupSet))
	for group := range groupSet {
		groups = append(groups, group)
	}

	return groups, nil
}

// AddGroup creates a new group by returning an error - file backend doesn't have standalone groups.
func (f *FileUserManagement) AddGroup(groupName string) error {
	return fmt.Errorf("standalone group creation is not supported for file-based authentication - groups are managed as user attributes")
}

// DeleteGroup deletes a group by returning an error - file backend doesn't have standalone groups.
func (f *FileUserManagement) DeleteGroup(groupName string) error {
	return fmt.Errorf("standalone group deletion is not supported for file-based authentication - groups are managed as user attributes")
}

// convertExtraAttributeValue converts an extra attribute value to the proper type based on configuration.
func (f *FileUserManagement) convertExtraAttributeValue(attrName string, value any) (any, error) {
	attrConfig, exists := f.provider.config.ExtraAttributes[attrName]
	if !exists {
		return value, nil
	}

	if attrConfig.MultiValued {
		if slice, ok := value.([]interface{}); ok {
			return slice, nil
		}

		if strSlice, ok := value.([]string); ok {
			result := make([]interface{}, len(strSlice))
			for i, s := range strSlice {
				result[i] = s
			}

			return result, nil
		}

		return []interface{}{value}, nil
	}

	switch attrConfig.ValueType {
	case ValueTypeBoolean:
		if boolVal, ok := value.(bool); ok {
			return boolVal, nil
		}

		if strVal, ok := value.(string); ok {
			switch strings.ToUpper(strVal) {
			case "TRUE", "T", "1", "YES", "Y":
				return true, nil
			case "FALSE", "F", "0", "NO", "N", "":
				return false, nil
			default:
				return nil, fmt.Errorf("invalid boolean value: %s", strVal)
			}
		}
	case ValueTypeInteger:
		if intVal, ok := value.(int); ok {
			return intVal, nil
		}

		if int64Val, ok := value.(int64); ok {
			return int64Val, nil
		}

		if strVal, ok := value.(string); ok {
			return strconv.ParseInt(strVal, 10, 64)
		}
	case ValueTypeString, "":
		if strVal, ok := value.(string); ok {
			return strVal, nil
		}

		return fmt.Sprintf("%v", value), nil
	}

	return value, nil
}
