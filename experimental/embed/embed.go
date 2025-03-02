package embed

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
)

func ProvidersStartupCheck(config *schema.Configuration, providers middlewares.Providers) (err error) {
	var (
		failures []string
	)

	if err = doStartupCheck(providers.StorageProvider, false); err != nil {
		failures = append(failures, "storage")
	}

	if err = doStartupCheck(providers.UserProvider, false); err != nil {
		failures = append(failures, "user")
	}

	if err = doStartupCheck(providers.Notifier, config.Notifier.DisableStartupCheck); err != nil {
		failures = append(failures, "notification")
	}

	if err = doStartupCheck(providers.NTP, config.NTP.DisableStartupCheck); err != nil {
		if !config.NTP.DisableFailure {
			failures = append(failures, "ntp")
		}
	}

	if len(failures) != 0 {
		return fmt.Errorf("errors occurred performing checks on the '%s' providers", strings.Join(failures, ", "))
	}

	return nil
}

func doStartupCheck(provider model.StartupCheck, disabled bool) error {
	if disabled {
		return nil
	}

	if provider == nil {
		return fmt.Errorf("unrecognized provider or it is not configured properly")
	}

	return provider.StartupCheck()
}
