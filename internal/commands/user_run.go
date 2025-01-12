package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/authentication"
)

// LoadProvidersAuthenticationRunE is a special PreRunE that loads the user authentication provider into the CmdCtx.
func (ctx *CmdCtx) LoadProvidersAuthenticationRunE(cmd *cobra.Command, args []string) (err error) {
	ctx.providers.UserProvider = getAuthenticationProvider(ctx)

	if err = doStartupCheck(ctx, providerNameUser, ctx.providers.UserProvider, false); err != nil {
		return err
	}

	return nil
}

// UserChangePasswordRunE updates user's password .
func (ctx *CmdCtx) UserChangePasswordRunE(cmd *cobra.Command, args []string) (err error) {
	if len(args) < 1 {
		return errors.New("invalid number of arguments")
	}

	var (
		username = args[0]
		password string
	)

	if len(args) > 1 {
		password = args[1]
	} else {
		fmt.Printf("Change Password for user %s\n", username)

		if password, err = askPassword(); err != nil {
			return err
		}
	}

	if err = ctx.providers.UserProvider.UpdatePassword(username, password); err != nil {
		return err
	}

	ctx.log.Info("password changed!")

	return nil
}

// UserShowInfoRunE shows user info.
func (ctx *CmdCtx) UserShowInfoRunE(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 1 {
		return errors.New("invalid number of arguments")
	}

	var username = args[0]

	var details *authentication.UserDetailsExtended

	var provider, ok = ctx.providers.UserProvider.(*authentication.DBUserProvider)

	if !ok {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if details, err = provider.GetDetailsExtended(username); err != nil {
		return err
	}

	fmt.Printf(`User '%s' Info:
	Display Name:	%s
	Email:		%s
	Groups:		%v
	Disabled:	%v
`, username, details.GetDisplayName(), strings.Join(details.GetEmails(), ", "), strings.Join(details.GetGroups(), ", "), details.Disabled)

	return nil
}

// UserAddRunE adds a user.
//nolint: gocyclo
func (ctx *CmdCtx) UserAddRunE(cmd *cobra.Command, args []string) (err error) {
	if ctx.config.AuthenticationBackend.DB == nil {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	var (
		username    = args[0]
		flags       = cmd.Flags()
		password    string
		email       string
		displayName string
		groups      []string
	)

	password, err = flags.GetString("password")
	if err != nil || password == "" {
		if password, err = askPassword(); err != nil {
			return err
		}
	}

	email, err = flags.GetString("email")
	if err != nil || email == "" {
		if email, err = askEmail(); err != nil {
			return err
		}
	}

	displayName, err = flags.GetString("display-name")
	if err != nil || displayName == "" {
		if displayName, err = askDisplayName(); err != nil {
			return err
		}
	}

	groups, err = flags.GetStringSlice("group")
	if err != nil || len(groups) == 0 {
		if groups, err = askGroups(); err != nil {
			return err
		}
	}

	var provider, ok = ctx.providers.UserProvider.(*authentication.DBUserProvider)

	if !ok {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	err = provider.AddUser(username, displayName, password, authentication.WithEmail(email), authentication.WithGroups(groups))
	if err != nil {
		return err
	}

	fmt.Println("user added.")

	return nil
}

// UserDeleteRunE deletes a user.
func (ctx *CmdCtx) UserDeleteRunE(cmd *cobra.Command, args []string) (err error) {
	if ctx.config.AuthenticationBackend.DB == nil {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if len(args) != 1 {
		return errors.New("invalid number of arguments")
	}

	var username = args[0]

	var provider, ok = ctx.providers.UserProvider.(*authentication.DBUserProvider)

	if !ok {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if err = provider.DeleteUser(username); err != nil {
		return err
	}

	fmt.Println("user deleted.")

	return nil
}

// UserChangeNameRunE changes the user display name.
func (ctx *CmdCtx) UserChangeNameRunE(cmd *cobra.Command, args []string) (err error) {
	if ctx.config.AuthenticationBackend.DB == nil {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if len(args) != 2 {
		return errors.New("invalid number of arguments")
	}

	var username = args[0]

	var name = args[1]

	var provider, ok = ctx.providers.UserProvider.(*authentication.DBUserProvider)

	if !ok {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if err = provider.ChangeDisplayName(username, name); err != nil {
		return err
	}

	fmt.Println("user's display name changed.")

	return nil
}

// UserChangeEmailRunE changes the user's email.
func (ctx *CmdCtx) UserChangeEmailRunE(cmd *cobra.Command, args []string) (err error) {
	if ctx.config.AuthenticationBackend.DB == nil {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	var username = args[0]

	var email = args[1]

	var provider, ok = ctx.providers.UserProvider.(*authentication.DBUserProvider)

	if !ok {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if err = provider.ChangeEmail(username, email); err != nil {
		return err
	}

	fmt.Println("user's email changed.")

	return nil
}

// UserChangeGroupsRunE changes the user's group name.
func (ctx *CmdCtx) UserChangeGroupsRunE(cmd *cobra.Command, args []string) (err error) {
	if ctx.config.AuthenticationBackend.DB == nil {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if len(args) < 2 {
		return errors.New("invalid number of arguments")
	}

	var username = args[0]

	var groups = args[1:]

	var provider, ok = ctx.providers.UserProvider.(*authentication.DBUserProvider)

	if !ok {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if err = provider.ChangeGroups(username, groups); err != nil {
		return err
	}

	fmt.Println("user's groups changed.")

	return nil
}

// UserListRunE list users.
func (ctx *CmdCtx) UserListRunE(cmd *cobra.Command, args []string) (err error) {
	if ctx.config.AuthenticationBackend.DB == nil {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	w := tabwriter.NewWriter(os.Stdout, 10, 1, 4, ' ', 0)

	_, _ = fmt.Fprintln(w, " Username\tDisplay Name\tEmail\tGroups\tDisabled")

	var provider, ok = ctx.providers.UserProvider.(*authentication.DBUserProvider)

	if !ok {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	users, err := provider.ListUsers()
	if err != nil {
		return err
	}

	for _, u := range users {
		disabled := "no"

		if u.Disabled {
			disabled = "yes"
		}

		_, _ = fmt.Fprintf(w, " %s\t%s\t%s\t%s\t%s\n",
			u.Username,
			u.DisplayName,
			strings.Join(u.Emails, ", "),
			strings.Join(u.Groups, ", "),
			disabled,
		)
	}

	_ = w.Flush()

	return nil
}

// UserDisableRunE disables a user.
func (ctx *CmdCtx) UserDisableRunE(cmd *cobra.Command, args []string) (err error) {
	if ctx.config.AuthenticationBackend.DB == nil {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if len(args) != 1 {
		return errors.New("invalid number of arguments")
	}

	var username = args[0]

	var provider, ok = ctx.providers.UserProvider.(*authentication.DBUserProvider)

	if !ok {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if err = provider.DisableUser(username); err != nil {
		return err
	}

	fmt.Println("user disabled.")

	return nil
}

// UserEnableRunE enables a user.
func (ctx *CmdCtx) UserEnableRunE(cmd *cobra.Command, args []string) (err error) {
	if ctx.config.AuthenticationBackend.DB == nil {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if len(args) != 1 {
		return errors.New("invalid number of arguments")
	}

	var username = args[0]

	var provider, ok = ctx.providers.UserProvider.(*authentication.DBUserProvider)

	if !ok {
		return errors.New("this command is only available for 'db' authentication backend")
	}

	if err = provider.EnableUser(username); err != nil {
		return err
	}

	fmt.Println("user enabled.")

	return nil
}
