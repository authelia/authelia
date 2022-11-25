package commands

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/authelia/authelia/v4/internal/utils"
)

func recoverErr(i any) error {
	switch v := i.(type) {
	case nil:
		return nil
	case string:
		return fmt.Errorf("recovered panic: %s", v)
	case error:
		return fmt.Errorf("recovered panic: %w", v)
	default:
		return fmt.Errorf("recovered panic with unknown type: %v", v)
	}
}

func configFilterExisting(configs []string) (finalConfigs []string) {
	var err error

	for _, c := range configs {
		if _, err = os.Stat(c); err == nil || !os.IsNotExist(err) {
			finalConfigs = append(finalConfigs, c)
		}
	}

	return finalConfigs
}

//nolint:gocyclo
func flagsGetRandomCharacters(flags *pflag.FlagSet, flagNameLength, flagNameCharSet, flagNameCharacters string) (r string, err error) {
	var (
		n       int
		charset string
	)

	if n, err = flags.GetInt(flagNameLength); err != nil {
		return "", err
	}

	if n < 1 {
		return "", fmt.Errorf("flag --%s with value '%d' is invalid: must be at least 1", flagNameLength, n)
	}

	useCharSet, useCharacters := flags.Changed(flagNameCharSet), flags.Changed(flagNameCharacters)

	if useCharSet && useCharacters {
		return "", fmt.Errorf("flag --%s and flag --%s are mutually exclusive, only one may be used", flagNameCharSet, flagNameCharacters)
	}

	switch {
	case useCharSet, !useCharSet && !useCharacters:
		var c string

		if c, err = flags.GetString(flagNameCharSet); err != nil {
			return "", err
		}

		switch c {
		case "ascii":
			charset = utils.CharSetASCII
		case "alphanumeric":
			charset = utils.CharSetAlphaNumeric
		case "alphanumeric-lower":
			charset = utils.CharSetAlphabeticLower + utils.CharSetNumeric
		case "alphanumeric-upper":
			charset = utils.CharSetAlphabeticUpper + utils.CharSetNumeric
		case "alphabetic":
			charset = utils.CharSetAlphabetic
		case "alphabetic-lower":
			charset = utils.CharSetAlphabeticLower
		case "alphabetic-upper":
			charset = utils.CharSetAlphabeticUpper
		case "numeric-hex":
			charset = utils.CharSetNumericHex
		case "numeric":
			charset = utils.CharSetNumeric
		case "rfc3986":
			charset = utils.CharSetRFC3986Unreserved
		case "rfc3986-lower":
			charset = utils.CharSetAlphabeticLower + utils.CharSetNumeric + utils.CharSetSymbolicRFC3986Unreserved
		case "rfc3986-upper":
			charset = utils.CharSetAlphabeticUpper + utils.CharSetNumeric + utils.CharSetSymbolicRFC3986Unreserved
		default:
			return "", fmt.Errorf("flag '--%s' with value '%s' is invalid, must be one of 'ascii', 'alphanumeric', 'alphabetic', 'numeric', 'numeric-hex', or 'rfc3986'", flagNameCharSet, c)
		}
	case useCharacters:
		if charset, err = flags.GetString(flagNameCharacters); err != nil {
			return "", err
		}
	}

	return utils.RandomString(n, charset, true), nil
}

func termReadPasswordStrWithPrompt(prompt string) (data string, err error) {
	var d []byte

	if d, err = termReadPasswordWithPrompt(prompt); err != nil {
		return "", err
	}

	return string(d), nil
}

func termReadPasswordWithPrompt(prompt string) (data []byte, err error) {
	fd := int(syscall.Stdin) //nolint:unconvert,nolintlint

	if isTerm := term.IsTerminal(fd); !isTerm {
		return nil, ErrStdinIsNotTerminal
	}

	fmt.Print(prompt)

	if data, err = term.ReadPassword(fd); err != nil {
		return nil, err
	}

	fmt.Println("")

	return data, nil
}
