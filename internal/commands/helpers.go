package commands

import (
	"bufio"
	"encoding/base32"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

func getStorageProvider(ctx *CmdCtx) (provider storage.Provider) {
	switch {
	case ctx.config.Storage.PostgreSQL != nil:
		return storage.NewPostgreSQLProvider(ctx.config, ctx.trusted)
	case ctx.config.Storage.MySQL != nil:
		return storage.NewMySQLProvider(ctx.config, ctx.trusted)
	case ctx.config.Storage.Local != nil:
		return storage.NewSQLiteProvider(ctx.config)
	default:
		return nil
	}
}

func containsIdentifier(identifier model.UserOpaqueIdentifier, identifiers []model.UserOpaqueIdentifier) bool {
	for i := 0; i < len(identifiers); i++ {
		if identifier.Service == identifiers[i].Service && identifier.SectorID == identifiers[i].SectorID && identifier.Username == identifiers[i].Username {
			return true
		}
	}

	return false
}

func storageWrapCheckSchemaErr(err error) error {
	switch {
	case errors.Is(err, errStorageSchemaIncompatible):
		return fmt.Errorf("command requires the use of a compatibe schema version: %w", err)
	case errors.Is(err, errStorageSchemaOutdated):
		return fmt.Errorf("command requires the use of a up to date schema version: %w", err)
	default:
		return err
	}
}

func storageTOTPGenerateRunEOptsFromFlags(flags *pflag.FlagSet) (force bool, filename, secret string, err error) {
	if force, err = flags.GetBool("force"); err != nil {
		return force, filename, secret, err
	}

	if filename, err = flags.GetString("path"); err != nil {
		return force, filename, secret, err
	}

	if secret, err = flags.GetString("secret"); err != nil {
		return force, filename, secret, err
	}

	secretLength := base32.StdEncoding.WithPadding(base32.NoPadding).DecodedLen(len(secret))
	if secret != "" && secretLength < schema.TOTPSecretSizeMinimum {
		return force, filename, secret, fmt.Errorf("decoded length of the base32 secret must have "+
			"a length of more than %d but '%s' has a decoded length of %d", schema.TOTPSecretSizeMinimum, secret, secretLength)
	}

	return force, filename, secret, nil
}

func storageWebAuthnDeleteRunEOptsFromFlags(flags *pflag.FlagSet, args []string) (all, byKID bool, description, kid, user string, err error) {
	if len(args) != 0 {
		user = args[0]
	}

	f := 0

	if flags.Changed(cmdFlagNameAll) {
		if all, err = flags.GetBool(cmdFlagNameAll); err != nil {
			return
		}

		f++
	}

	if flags.Changed(cmdFlagNameDescription) {
		if description, err = flags.GetString(cmdFlagNameDescription); err != nil {
			return
		}

		f++
	}

	if byKID = flags.Changed(cmdFlagNameKeyID); byKID {
		if kid, err = flags.GetString(cmdFlagNameKeyID); err != nil {
			return
		}

		f++
	}

	if f > 1 {
		err = fmt.Errorf("must only supply one of the flags --all, --description, and --kid but %d were specified", f)

		return
	}

	if f == 0 {
		err = fmt.Errorf("must supply one of the flags --all, --description, or --kid")

		return
	}

	if !byKID && len(user) == 0 {
		err = fmt.Errorf("must supply the username or the --kid flag")

		return
	}

	return
}

func getAuthenticationProvider(ctx *CmdCtx) authentication.UserProvider {
	switch {
	case ctx.config.AuthenticationBackend.File != nil:
		return authentication.NewFileUserProvider(ctx.config.AuthenticationBackend.File)
	case ctx.config.AuthenticationBackend.LDAP != nil:
		return authentication.NewLDAPUserProvider(ctx.config.AuthenticationBackend, ctx.trusted)
	case ctx.config.AuthenticationBackend.DB != nil:
		return authentication.NewDBUserProvider(ctx.config.AuthenticationBackend.DB, ctx.providers.StorageProvider)
	}

	return nil
}

func askPassword() (string, error) {
	var (
		password             string
		passwordConfirmation string
	)

	fmt.Print("Password: ")

	bytePassword, err := term.ReadPassword(syscall.Stdin)

	if err != nil {
		return "", fmt.Errorf("cant read password: %w", err)
	}

	password = string(bytePassword)

	fmt.Println()

	fmt.Print("Repeat password: ")

	bytePassword, err = term.ReadPassword(syscall.Stdin)

	if err != nil {
		return "", fmt.Errorf("cant read password: %w", err)
	}

	passwordConfirmation = string(bytePassword)

	fmt.Println()

	if password != passwordConfirmation {
		return "", errors.New("passwords doesn't match")
	}

	if password == "" {
		return "", errors.New("password cant be empty")
	}

	return password, nil
}

func askEmail() (string, error) {
	in := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")

	email, err := in.ReadString('\n')

	if err != nil {
		return "", fmt.Errorf("failed to read email: %s", err)
	}

	email = strings.TrimSpace(email)

	if email == "" {
		return "", errors.New("email is required")
	}

	return email, nil
}

func askDisplayName() (string, error) {
	in := bufio.NewReader(os.Stdin)

	fmt.Print("Display Name: ")

	displayName, err := in.ReadString('\n')

	if err != nil {
		return "", fmt.Errorf("failed to read display name: %w", err)
	}

	displayName = strings.TrimSpace(displayName)

	return displayName, nil
}

func askGroups() ([]string, error) {
	in := bufio.NewReader(os.Stdin)

	var groupsStr string

	fmt.Print("Groups (comma separated): ")

	groupsStr, err := in.ReadString('\n')

	if err != nil {
		return []string{}, fmt.Errorf("failed to read groups: %w", err)
	}

	groups := strings.Split(groupsStr, ",")

	for i := range groups {
		groups[i] = strings.TrimSpace(groups[i])
	}

	return groups, nil
}
