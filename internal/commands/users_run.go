package commands

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"golang.org/x/text/language"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// ConfigValidateAdministrationRunE validates the administration config before running user management commands.
func (ctx *CmdCtx) ConfigValidateAdministrationRunE(_ *cobra.Command, _ []string) (err error) {
	if errs := ctx.cconfig.validator.Errors(); len(errs) != 0 {
		var (
			i int
			e error
		)

		for i, e = range errs {
			if i == 0 {
				err = e
				continue
			}

			err = fmt.Errorf("%w, %v", err, e)
		}

		return err
	}

	validator.ValidateAdministration(ctx.config, ctx.cconfig.validator)

	if errs := ctx.cconfig.validator.Errors(); len(errs) != 0 {
		var (
			i int
			e error
		)

		for i, e = range errs {
			if i == 0 {
				err = e
				continue
			}

			err = fmt.Errorf("%w, %v", err, e)
		}

		return err
	}

	return nil
}

// ConfigValidateUserBackendRunE validates the authentication backend config before running user management commands.
func (ctx *CmdCtx) ConfigValidateUserBackendRunE(_ *cobra.Command, _ []string) (err error) {
	if errs := ctx.cconfig.validator.Errors(); len(errs) != 0 {
		var (
			i int
			e error
		)

		for i, e = range errs {
			if i == 0 {
				err = e
				continue
			}

			err = fmt.Errorf("%w, %v", err, e)
		}

		return err
	}

	validator.ValidateAuthenticationBackend(ctx.config, ctx.cconfig.validator)

	if errs := ctx.cconfig.validator.Errors(); len(errs) != 0 {
		var (
			i int
			e error
		)

		for i, e = range errs {
			if i == 0 {
				err = e
				continue
			}

			err = fmt.Errorf("%w, %v", err, e)
		}

		return err
	}

	return nil
}

// LoadProvidersUserBackendRunE loads the authentication backend (user provider) into the CmdCtx.
func (ctx *CmdCtx) LoadProvidersUserBackendRunE(_ *cobra.Command, _ []string) (err error) {
	ctx.providers.UserProvider = middlewares.NewAuthenticationProvider(ctx.config, ctx.trusted)

	if ctx.providers.UserProvider == nil {
		return fmt.Errorf("user management requires a configured authentication backend (file or ldap)")
	}

	return nil
}

func (ctx *CmdCtx) UsersSchemaPrintRunE() func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		return nil
	}
}

func (ctx *CmdCtx) UsersGetRunE() func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err := ctx.providers.StorageProvider.Close(); err != nil {
				panic(err)
			}
		}()

		var format string

		if format, err = cmd.Flags().GetString(cmdFlagNameFormat); err != nil {
			return err
		}

		if format == cmdFlagValueFormatJSON && cmd.Flags().Changed(cmdFlagNameFields) {
			return fmt.Errorf("flag '--%s' cannot be used with '--%s %s'", cmdFlagNameFields, cmdFlagNameFormat, cmdFlagValueFormatJSON)
		}

		var fields []string

		if cmd.Flags().Changed(cmdFlagNameFields) {
			if fields, err = cmd.Flags().GetStringSlice(cmdFlagNameFields); err != nil {
				return err
			}

			supported := ctx.providers.UserProvider.GetSupportedAttributes()

			for _, f := range fields {
				if _, ok := supported[f]; !ok {
					return fmt.Errorf("field '%s' is not supported by the configured authentication backend", f)
				}
			}
		} else {
			required := ctx.providers.UserProvider.GetRequiredAttributes()

			for _, f := range required {
				if f != authentication.AttributePassword {
					fields = append(fields, f)
				}
			}
		}

		var user *authentication.UserDetailsExtended

		if user, err = ctx.providers.UserProvider.GetUser(args[0]); err != nil {
			return fmt.Errorf("error occurred retrieving user '%s': %w", args[0], err)
		}

		user.Password = ""

		return FormatUserOutput(cmd.OutOrStdout(), []authentication.UserDetailsExtended{*user}, fields, format)
	}
}

// userFieldExtractors maps attribute names (as returned by GetSupportedAttributes) to functions
// that extract a display string from a UserDetailsExtended. extra.* fields are handled dynamically
// in FormatUserOutput and do not need entries here.
var userFieldExtractors = map[string]func(*authentication.UserDetailsExtended) string{
	authentication.AttributeUsername:             func(u *authentication.UserDetailsExtended) string { return u.GetUsername() },
	authentication.AttributeDisplayName:          func(u *authentication.UserDetailsExtended) string { return u.GetDisplayName() },
	authentication.AttributeMail:                 func(u *authentication.UserDetailsExtended) string { return strings.Join(u.GetEmails(), ", ") },
	authentication.AttributeGroups:               func(u *authentication.UserDetailsExtended) string { return strings.Join(u.GetGroups(), ", ") },
	authentication.AttributeGivenName:            func(u *authentication.UserDetailsExtended) string { return u.GetGivenName() },
	authentication.AttributeFamilyName:           func(u *authentication.UserDetailsExtended) string { return u.GetFamilyName() },
	authentication.AttributeMiddleName:           func(u *authentication.UserDetailsExtended) string { return u.GetMiddleName() },
	authentication.AttributeCommonName:           func(u *authentication.UserDetailsExtended) string { return u.CommonName },
	authentication.AttributeNickname:             func(u *authentication.UserDetailsExtended) string { return u.GetNickname() },
	authentication.AttributeGender:               func(u *authentication.UserDetailsExtended) string { return u.GetGender() },
	authentication.AttributeBirthdate:            func(u *authentication.UserDetailsExtended) string { return u.GetBirthdate() },
	authentication.AttributeWebsite:              func(u *authentication.UserDetailsExtended) string { return u.GetWebsite() },
	authentication.AttributeProfile:              func(u *authentication.UserDetailsExtended) string { return u.GetProfile() },
	authentication.AttributePicture:              func(u *authentication.UserDetailsExtended) string { return u.GetPicture() },
	authentication.AttributeZoneInfo:             func(u *authentication.UserDetailsExtended) string { return u.GetZoneInfo() },
	authentication.AttributeLocale:               func(u *authentication.UserDetailsExtended) string { return u.GetLocale() },
	authentication.AttributePhoneNumber:          func(u *authentication.UserDetailsExtended) string { return u.GetPhoneNumber() },
	authentication.AttributePhoneExtension:       func(u *authentication.UserDetailsExtended) string { return u.GetPhoneExtension() },
	authentication.AttributeAddressStreetAddress: func(u *authentication.UserDetailsExtended) string { return u.GetStreetAddress() },
	authentication.AttributeAddressLocality:      func(u *authentication.UserDetailsExtended) string { return u.GetLocality() },
	authentication.AttributeAddressRegion:        func(u *authentication.UserDetailsExtended) string { return u.GetRegion() },
	authentication.AttributeAddressPostalCode:    func(u *authentication.UserDetailsExtended) string { return u.GetPostalCode() },
	authentication.AttributeAddressCountry:       func(u *authentication.UserDetailsExtended) string { return u.GetCountry() },
}

func (ctx *CmdCtx) UsersListRunE() func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err := ctx.providers.StorageProvider.Close(); err != nil {
				panic(err)
			}
		}()

		var format string

		if format, err = cmd.Flags().GetString(cmdFlagNameFormat); err != nil {
			return err
		}

		if format == cmdFlagValueFormatJSON && cmd.Flags().Changed(cmdFlagNameFields) {
			return fmt.Errorf("flag '--%s' cannot be used with '--%s %s'", cmdFlagNameFields, cmdFlagNameFormat, cmdFlagValueFormatJSON)
		}

		var fields []string

		if cmd.Flags().Changed(cmdFlagNameFields) {
			if fields, err = cmd.Flags().GetStringSlice(cmdFlagNameFields); err != nil {
				return err
			}

			supported := ctx.providers.UserProvider.GetSupportedAttributes()

			for _, f := range fields {
				if _, ok := supported[f]; !ok {
					return fmt.Errorf("field '%s' is not supported by the configured authentication backend", f)
				}
			}
		} else {
			required := ctx.providers.UserProvider.GetRequiredAttributes()

			for _, f := range required {
				if f != authentication.AttributePassword {
					fields = append(fields, f)
				}
			}
		}

		var users []authentication.UserDetailsExtended

		if users, err = ctx.providers.UserProvider.ListUsers(); err != nil {
			return fmt.Errorf("error occurred retrieving user list: %w", err)
		}

		for i := range users {
			users[i].Password = ""
		}

		return FormatUserOutput(cmd.OutOrStdout(), users, fields, format)
	}
}

// FormatUserOutput writes a list of users to w in the requested format (table or json).
func FormatUserOutput(w io.Writer, users []authentication.UserDetailsExtended, fields []string, format string) (err error) {
	switch format {
	case cmdFlagValueFormatJSON:
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)

		if err = encoder.Encode(users); err != nil {
			return fmt.Errorf("error occurred encoding users as JSON: %w", err)
		}

		return nil
	default:
		tw := tabwriter.NewWriter(w, 1, 1, 4, ' ', 0)

		_, _ = fmt.Fprintln(tw, strings.Join(fields, "\t"))

		for _, u := range users {
			values := make([]string, len(fields))

			for i, field := range fields {
				if extractor, ok := userFieldExtractors[field]; ok {
					values[i] = extractor(&u)
				} else if strings.HasPrefix(field, authentication.PrefixAttributeExtra) {
					extraKey := strings.TrimPrefix(field, authentication.PrefixAttributeExtra)
					if u.Extra != nil {
						if v, exists := u.Extra[extraKey]; exists {
							values[i] = fmt.Sprint(v)
						}
					}
				}
			}

			_, _ = fmt.Fprintln(tw, strings.Join(values, "\t"))
		}

		return tw.Flush()
	}
}

func (ctx *CmdCtx) UsersAddRunE() func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err := ctx.providers.StorageProvider.Close(); err != nil {
				panic(err)
			}
		}()

		var newUser *authentication.UserDetailsExtended

		if cmd.Flags().Changed(cmdFlagNameFile) {
			if newUser, err = usersAddFromJSON(cmd); err != nil {
				return err
			}
		} else {
			if newUser, err = ctx.usersAddInteractive(cmd); err != nil {
				return err
			}
		}

		if err = ctx.providers.UserProvider.ValidateUserData(newUser); err != nil {
			switch {
			case errors.Is(err, authentication.ErrUsernameIsRequired):
				return fmt.Errorf("username is required")
			case errors.Is(err, authentication.ErrFamilyNameIsRequired):
				return fmt.Errorf("family name (last name) is required for this backend")
			default:
				return fmt.Errorf("validation failed for user '%s': %w", newUser.GetUsername(), err)
			}
		}

		passwordPolicy := middlewares.NewPasswordPolicyProvider(ctx.config.PasswordPolicy)

		if err = passwordPolicy.Check(newUser.Password); err != nil {
			return fmt.Errorf("password does not meet policy requirements: %w", err)
		}

		if err = ctx.providers.UserProvider.AddUser(newUser); err != nil {
			return fmt.Errorf("error occurred creating user '%s': %w", newUser.GetUsername(), err)
		}

		if _, err = ctx.providers.StorageProvider.LoadUserMetadataByUsername(ctx, newUser.GetUsername()); err != nil {
			if err = ctx.providers.StorageProvider.CreateNewUserMetadata(ctx, newUser.GetUsername()); err != nil {
				return fmt.Errorf("error occurred creating metadata for user '%s': %w", newUser.GetUsername(), err)
			}
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully created user '%s'.\n", newUser.GetUsername())

		return nil
	}
}

func usersAddFromJSON(cmd *cobra.Command) (user *authentication.UserDetailsExtended, err error) {
	var filePath string

	if filePath, err = cmd.Flags().GetString(cmdFlagNameFile); err != nil {
		return nil, err
	}

	var data []byte

	switch filePath {
	case "-":
		if data, err = io.ReadAll(cmd.InOrStdin()); err != nil {
			return nil, fmt.Errorf("error reading from stdin: %w", err)
		}
	default:
		if data, err = os.ReadFile(filePath); err != nil {
			return nil, fmt.Errorf("error reading file '%s': %w", filePath, err)
		}
	}

	user = &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}

	if err = json.Unmarshal(data, user); err != nil {
		return nil, fmt.Errorf("error parsing user JSON: %w", err)
	}

	return user, nil
}

func (ctx *CmdCtx) usersAddInteractive(cmd *cobra.Command) (user *authentication.UserDetailsExtended, err error) {
	required := ctx.providers.UserProvider.GetRequiredAttributes()
	supported := ctx.providers.UserProvider.GetSupportedAttributes()

	user = &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}

	out := cmd.OutOrStdout()

	for _, attr := range required {
		meta, ok := supported[attr]
		if !ok {
			meta = authentication.UserManagementAttributeMetadata{Type: authentication.Text}
		}

		if err = usersAddPromptAttribute(out, user, attr, meta); err != nil {
			return nil, err
		}
	}

	requiredSet := make(map[string]struct{}, len(required))
	for _, r := range required {
		requiredSet[r] = struct{}{}
	}

	optional := make([]string, 0, len(supported))
	for attr := range supported {
		if _, isRequired := requiredSet[attr]; !isRequired {
			optional = append(optional, attr)
		}
	}

	sort.Strings(optional)

	if len(optional) == 0 {
		return user, nil
	}

	_, _ = fmt.Fprintf(out, "Add additional attributes? Supported: %s\n(Enter comma-separated list or leave empty to skip): ", strings.Join(optional, ", "))

	scanner := bufio.NewScanner(os.Stdin)

	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line != "" {
			for _, sel := range strings.Split(line, ",") {
				sel = strings.TrimSpace(sel)

				if sel == "" {
					continue
				}

				meta, ok := supported[sel]
				if !ok {
					return nil, fmt.Errorf("attribute '%s' is not supported by the configured authentication backend", sel)
				}

				if err = usersAddPromptAttribute(out, user, sel, meta); err != nil {
					return nil, err
				}
			}
		}
	}

	return user, nil
}

func usersAddPromptAttribute(out io.Writer, user *authentication.UserDetailsExtended, attr string, meta authentication.UserManagementAttributeMetadata) (err error) {
	label := strings.ReplaceAll(attr, "_", " ")

	switch meta.Type {
	case authentication.Password:
		var password string

		if password, err = termReadPasswordWithPrompt(fmt.Sprintf("Enter %s: ", label), ""); err != nil {
			return err
		}

		var confirm string

		if confirm, err = termReadPasswordWithPrompt(fmt.Sprintf("Confirm %s: ", label), ""); err != nil {
			return err
		}

		if password != confirm {
			return fmt.Errorf("passwords do not match")
		}

		return usersSetField(user, attr, password)

	case authentication.Groups:
		_, _ = fmt.Fprintf(out, "Enter %s (comma-separated): ", label)

		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			parts := strings.Split(scanner.Text(), ",")
			groups := make([]string, 0, len(parts))

			for _, p := range parts {
				if p = strings.TrimSpace(p); p != "" {
					groups = append(groups, p)
				}
			}

			return usersSetFieldMultiple(user, attr, groups)
		}

	case authentication.Checkbox:
		_, _ = fmt.Fprintf(out, "Enter %s (yes/no): ", label)

		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			val := strings.ToLower(strings.TrimSpace(scanner.Text()))
			return usersSetFieldBool(user, attr, val == "yes" || val == "y" || val == "true")
		}

	case authentication.Number:
		_, _ = fmt.Fprintf(out, "Enter %s: ", label)

		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			val := strings.TrimSpace(scanner.Text())

			var n int64

			if n, err = strconv.ParseInt(val, 10, 64); err != nil {
				return fmt.Errorf("invalid number for attribute '%s': %w", attr, err)
			}

			return usersSetFieldNumber(user, attr, n)
		}

	default:
		if meta.Multiple {
			_, _ = fmt.Fprintf(out, "Enter %s (comma-separated): ", label)

			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				parts := strings.Split(scanner.Text(), ",")
				values := make([]string, 0, len(parts))

				for _, p := range parts {
					if p = strings.TrimSpace(p); p != "" {
						values = append(values, p)
					}
				}

				return usersSetFieldMultiple(user, attr, values)
			}
		} else {
			_, _ = fmt.Fprintf(out, "Enter %s: ", label)

			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				return usersSetField(user, attr, strings.TrimSpace(scanner.Text()))
			}
		}
	}

	return nil
}

func usersSetField(user *authentication.UserDetailsExtended, attr, value string) (err error) {
	switch attr {
	case authentication.AttributeUsername:
		user.Username = value
	case authentication.AttributePassword:
		user.Password = value
	case authentication.AttributeDisplayName:
		user.DisplayName = value
	case authentication.AttributeMail:
		user.Emails = []string{value}
	case authentication.AttributeGivenName:
		user.GivenName = value
	case authentication.AttributeFamilyName:
		user.FamilyName = value
	case authentication.AttributeMiddleName:
		user.MiddleName = value
	case authentication.AttributeCommonName:
		user.CommonName = value
	case authentication.AttributeNickname:
		user.Nickname = value
	case authentication.AttributeGender:
		user.Gender = value
	case authentication.AttributeBirthdate:
		user.Birthdate = value
	case authentication.AttributeZoneInfo:
		user.ZoneInfo = value
	case authentication.AttributePhoneNumber:
		user.PhoneNumber = value
	case authentication.AttributePhoneExtension:
		user.PhoneExtension = value
	case authentication.AttributeProfile:
		var u *url.URL

		if u, err = url.Parse(value); err != nil {
			return fmt.Errorf("invalid URL for attribute '%s': %w", attr, err)
		}

		user.Profile = u
	case authentication.AttributePicture:
		var u *url.URL

		if u, err = url.Parse(value); err != nil {
			return fmt.Errorf("invalid URL for attribute '%s': %w", attr, err)
		}

		user.Picture = u
	case authentication.AttributeWebsite:
		var u *url.URL

		if u, err = url.Parse(value); err != nil {
			return fmt.Errorf("invalid URL for attribute '%s': %w", attr, err)
		}

		user.Website = u
	case authentication.AttributeLocale:
		var tag language.Tag

		if tag, err = language.Parse(value); err != nil {
			return fmt.Errorf("invalid locale for attribute '%s': %w", attr, err)
		}

		user.Locale = &tag
	case authentication.AttributeAddressStreetAddress:
		if user.Address == nil {
			user.Address = &authentication.UserDetailsAddress{}
		}

		user.Address.StreetAddress = value
	case authentication.AttributeAddressLocality:
		if user.Address == nil {
			user.Address = &authentication.UserDetailsAddress{}
		}

		user.Address.Locality = value
	case authentication.AttributeAddressRegion:
		if user.Address == nil {
			user.Address = &authentication.UserDetailsAddress{}
		}

		user.Address.Region = value
	case authentication.AttributeAddressPostalCode:
		if user.Address == nil {
			user.Address = &authentication.UserDetailsAddress{}
		}

		user.Address.PostalCode = value
	case authentication.AttributeAddressCountry:
		if user.Address == nil {
			user.Address = &authentication.UserDetailsAddress{}
		}

		user.Address.Country = value
	default:
		if strings.HasPrefix(attr, authentication.PrefixAttributeExtra) {
			extraKey := strings.TrimPrefix(attr, authentication.PrefixAttributeExtra)

			if user.Extra == nil {
				user.Extra = make(map[string]any)
			}

			user.Extra[extraKey] = value
		}
	}

	return nil
}

func usersSetFieldMultiple(user *authentication.UserDetailsExtended, attr string, values []string) error {
	switch attr {
	case authentication.AttributeGroups:
		user.Groups = values
	case authentication.AttributeMail:
		user.Emails = values
	default:
		if strings.HasPrefix(attr, authentication.PrefixAttributeExtra) {
			extraKey := strings.TrimPrefix(attr, authentication.PrefixAttributeExtra)

			if user.Extra == nil {
				user.Extra = make(map[string]any)
			}

			user.Extra[extraKey] = values
		}
	}

	return nil
}

func usersSetFieldBool(user *authentication.UserDetailsExtended, attr string, value bool) error {
	if strings.HasPrefix(attr, authentication.PrefixAttributeExtra) {
		extraKey := strings.TrimPrefix(attr, authentication.PrefixAttributeExtra)

		if user.Extra == nil {
			user.Extra = make(map[string]any)
		}

		user.Extra[extraKey] = value
	}

	return nil
}

func usersSetFieldNumber(user *authentication.UserDetailsExtended, attr string, value int64) error {
	if strings.HasPrefix(attr, authentication.PrefixAttributeExtra) {
		extraKey := strings.TrimPrefix(attr, authentication.PrefixAttributeExtra)

		if user.Extra == nil {
			user.Extra = make(map[string]any)
		}

		user.Extra[extraKey] = value
	}

	return nil
}

func (ctx *CmdCtx) UsersUpdateRunE() func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		return nil
	}
}

func (ctx *CmdCtx) UsersDeleteRunE() func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err := ctx.providers.StorageProvider.Close(); err != nil {
				panic(err)
			}
		}()

		username := args[0]

		if err = ctx.providers.UserProvider.DeleteUser(username); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error occurred deleting user '%s' from the authentication backend: %v\n", username, err)
		}

		if err = ctx.providers.StorageProvider.DeleteTOTPConfiguration(ctx, username); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error occurred deleting TOTP data for user '%s': %v\n", username, err)
		}

		if err = ctx.providers.StorageProvider.DeleteWebAuthnCredentialByUsername(ctx, username, ""); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error occurred deleting WebAuthn data for user '%s': %v\n", username, err)
		}

		if err = ctx.providers.StorageProvider.DeletePreferredDuoDevice(ctx, username); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error occurred deleting Duo data for user '%s': %v\n", username, err)
		}

		if err = ctx.providers.StorageProvider.DeleteUserByUsername(ctx, username); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error occurred deleting metadata for user '%s': %v\n", username, err)
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted user '%s'.\n", username)

		return nil
	}
}

func (ctx *CmdCtx) UsersGroupsListRunE() func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		return nil
	}
}

func (ctx *CmdCtx) UsersGroupsAddRunE() func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		// groupName := args[0].
		return nil
	}
}

func (ctx *CmdCtx) UsersGroupsDeleteRunE() func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		// groupName := args[0].
		return nil
	}
}
