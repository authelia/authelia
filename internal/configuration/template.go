package configuration

import (
	_ "embed" // Embed config.template.yml.
	"fmt"
	"os"
)

//go:embed config.template.yml
var conftemplate []byte

// EnsureConfigurationExists is an auxiliary function to the main Configuration tools that ensures the Configuration
// template is created if it doesn't already exist.
func EnsureConfigurationExists(path string) (created bool, err error) {
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.WriteFile(path, conftemplate, 0600); err != nil {
				return false, fmt.Errorf(errFmtGenerateConfiguration, err)
			}

			return true, nil
		}

		return false, fmt.Errorf(errFmtGenerateConfiguration, err)
	}

	return false, nil
}
