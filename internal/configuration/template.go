package configuration

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
)

//go:embed config.template.yml
var template []byte

// EnsureConfigurationExists is an auxilery function to the main configuration tools that ensures the configuration
// template is created if it doesn't already exist.
func EnsureConfigurationExists(paths []string) (created bool, err error) {
	if len(paths) != 1 {
		return false, nil
	}

	path := paths[0]

	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err := ioutil.WriteFile(path, template, 0600)
			if err != nil {
				return false, fmt.Errorf(errFmtGenerateConfiguration, err)
			}

			return true, nil
		}

		return false, fmt.Errorf(errFmtGenerateConfiguration, err)
	}

	return false, nil
}
