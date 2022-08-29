package commands

import (
	"fmt"
	"os"
)

func recoverErr(i interface{}) error {
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
