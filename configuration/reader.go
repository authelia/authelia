package configuration

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/clems4ever/authelia/configuration/schema"
	"github.com/clems4ever/authelia/configuration/validator"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Read a YAML configuration and create a Configuration object out of it.
func Read(configPath string) (*schema.Configuration, []error) {
	config := schema.Configuration{}

	data, err := ioutil.ReadFile(configPath)
	check(err)

	err = yaml.Unmarshal([]byte(data), &config)

	if err != nil {
		return nil, []error{err}
	}

	val := schema.NewStructValidator()
	validator.Validate(&config, val)

	if val.HasErrors() {
		return nil, val.Errors()
	}

	return &config, nil
}
