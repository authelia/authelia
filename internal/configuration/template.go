package configuration

import (
	_ "embed"
	"io/ioutil"
)

//go:embed config.template.yml
var configTemplate []byte

func generateConfigFromTemplate(configPath string) error {
	err := ioutil.WriteFile(configPath, configTemplate, 0600)
	if err != nil {
		return err
	}

	return nil
}
