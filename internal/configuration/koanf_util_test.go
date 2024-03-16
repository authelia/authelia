package configuration

import (
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

type testDeprecationsConf struct {
	SubItems []testDeprecationsConfSubItem `koanf:"subitems"`

	ANonSubItemString string `koanf:"a_non_subitem_string"`
	ANonSubItemInt    int    `koanf:"a_non_subitem_int"`
	ANonSubItemBool   bool   `koanf:"a_non_subitem_bool"`
}

type testDeprecationsConfSubItem struct {
	AString string `koanf:"a_string"`
	AnInt   int    `koanf:"an_int"`
	ABool   bool   `koanf:"a_bool"`
}

func TestSubItemRemap(t *testing.T) {
	ds := map[string]Deprecation{
		"astring": {
			Key:     "astring",
			NewKey:  "a_non_subitem_string",
			Version: model.SemanticVersion{Major: 4, Minor: 30},
			AutoMap: true,
		},
		"subitems[].astring": {
			Key:     "subitems[].astring",
			NewKey:  "subitems[].a_string",
			Version: model.SemanticVersion{Major: 4, Minor: 30},
			AutoMap: true,
		},
	}

	val := schema.NewStructValidator()

	ko := koanf.New(".")

	configYAML := []byte(`
astring: test
subitems:
- astring: example
- an_int: 1
`)

	require.NoError(t, ko.Load(rawbytes.Provider(configYAML), yaml.Parser()))

	final, err := koanfRemapKeys(val, ko, ds, nil)
	require.NoError(t, err)

	conf := &testDeprecationsConf{}

	require.NoError(t, final.Unmarshal("", conf))

	assert.Equal(t, "test", conf.ANonSubItemString)
	assert.Equal(t, 0, conf.ANonSubItemInt)
	assert.False(t, conf.ANonSubItemBool)

	require.Len(t, conf.SubItems, 2)
	assert.Equal(t, "example", conf.SubItems[0].AString)
	assert.Equal(t, 0, conf.SubItems[0].AnInt)
	assert.Equal(t, "", conf.SubItems[1].AString)
	assert.Equal(t, 1, conf.SubItems[1].AnInt)
}
