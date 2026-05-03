package configuration

import (
	"fmt"
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

type testDottedMapKeysConf struct {
	Scopes          map[string]testDottedScope          `koanf:"scopes"`
	ClaimsPolicies  map[string]testDottedClaimsPolicy   `koanf:"claims_policies"`
	AuthzPolicies   map[string]testDottedAuthzPolicy    `koanf:"authorization_policies"`
	CustomLifespans map[string]testDottedCustomLifespan `koanf:"custom"`
}

type testDottedScope struct {
	Claims []string `koanf:"claims"`
}

type testDottedClaimsPolicy struct {
	CustomClaims map[string]testDottedCustomClaim `koanf:"custom_claims"`
}

type testDottedCustomClaim struct {
	Name      string `koanf:"name"`
	Attribute string `koanf:"attribute"`
}

type testDottedAuthzPolicy struct {
	DefaultPolicy string `koanf:"default_policy"`
}

type testDottedCustomLifespan struct {
	AccessToken string `koanf:"access_token"`
}

func TestKoanfRemapKeys_DottedScopeNames(t *testing.T) {
	val := schema.NewStructValidator()

	ko := koanf.New(".")

	configYAML := []byte(`
scopes:
  my.scope:
    claims:
      - sub
      - email
  normal_scope:
    claims:
      - profile
`)

	require.NoError(t, ko.Load(rawbytes.Provider(configYAML), yaml.Parser()))

	final, err := koanfRemapKeys(val, ko, nil, nil)
	require.NoError(t, err)

	conf := &testDottedMapKeysConf{}

	require.NoError(t, final.Unmarshal("", conf))

	require.Contains(t, conf.Scopes, "my.scope", "dotted scope name should be preserved as a single map key")
	assert.Equal(t, []string{"sub", "email"}, conf.Scopes["my.scope"].Claims)

	require.Contains(t, conf.Scopes, "normal_scope")
	assert.Equal(t, []string{"profile"}, conf.Scopes["normal_scope"].Claims)

	assert.NotContains(t, conf.Scopes, "my", "dotted scope name should not be split into nested keys")
}

func TestKoanfRemapKeys_DottedCustomClaimNames(t *testing.T) {
	val := schema.NewStructValidator()

	ko := koanf.New(".")

	configYAML := []byte(`
claims_policies:
  my_policy:
    custom_claims:
      http://example.com/claim:
        name: example_claim
        attribute: display_name
      simple_claim:
        name: simple
        attribute: email
`)

	require.NoError(t, ko.Load(rawbytes.Provider(configYAML), yaml.Parser()))

	final, err := koanfRemapKeys(val, ko, nil, nil)
	require.NoError(t, err)

	conf := &testDottedMapKeysConf{}

	require.NoError(t, final.Unmarshal("", conf))

	require.Contains(t, conf.ClaimsPolicies, "my_policy")

	policy := conf.ClaimsPolicies["my_policy"]

	require.Contains(t, policy.CustomClaims, "http://example.com/claim", "URI-style claim name with dots should be preserved")
	assert.Equal(t, "example_claim", policy.CustomClaims["http://example.com/claim"].Name)
	assert.Equal(t, "display_name", policy.CustomClaims["http://example.com/claim"].Attribute)

	require.Contains(t, policy.CustomClaims, "simple_claim")
	assert.Equal(t, "simple", policy.CustomClaims["simple_claim"].Name)
}

func TestKoanfRemapKeys_DottedAuthorizationPolicyNames(t *testing.T) {
	val := schema.NewStructValidator()

	ko := koanf.New(".")

	configYAML := []byte(`
authorization_policies:
  my.policy:
    default_policy: two_factor
  normal_policy:
    default_policy: one_factor
`)

	require.NoError(t, ko.Load(rawbytes.Provider(configYAML), yaml.Parser()))

	final, err := koanfRemapKeys(val, ko, nil, nil)
	require.NoError(t, err)

	conf := &testDottedMapKeysConf{}

	require.NoError(t, final.Unmarshal("", conf))

	require.Contains(t, conf.AuthzPolicies, "my.policy", "dotted policy name should be preserved")
	assert.Equal(t, "two_factor", conf.AuthzPolicies["my.policy"].DefaultPolicy)

	require.Contains(t, conf.AuthzPolicies, "normal_policy")

	assert.NotContains(t, conf.AuthzPolicies, "my", "dotted policy name should not be split")
}

func TestKoanfRemapKeys_DottedCustomLifespanNames(t *testing.T) {
	val := schema.NewStructValidator()

	ko := koanf.New(".")

	configYAML := []byte(`
custom:
  my.lifespan:
    access_token: 1h
  normal_lifespan:
    access_token: 2h
`)

	require.NoError(t, ko.Load(rawbytes.Provider(configYAML), yaml.Parser()))

	final, err := koanfRemapKeys(val, ko, nil, nil)
	require.NoError(t, err)

	conf := &testDottedMapKeysConf{}

	require.NoError(t, final.Unmarshal("", conf))

	require.Contains(t, conf.CustomLifespans, "my.lifespan", "dotted lifespan name should be preserved")
	assert.Equal(t, "1h", conf.CustomLifespans["my.lifespan"].AccessToken)

	require.Contains(t, conf.CustomLifespans, "normal_lifespan")

	assert.NotContains(t, conf.CustomLifespans, "my", "dotted lifespan name should not be split")
}

func TestKoanfRemapKeys_DottedMapKeysWithDeprecationRemap(t *testing.T) {
	ds := map[string]Deprecation{
		"astring": {
			Key:     "astring",
			NewKey:  "a_non_subitem_string",
			Version: model.SemanticVersion{Major: 4, Minor: 30},
			AutoMap: true,
		},
	}

	val := schema.NewStructValidator()

	ko := koanf.New(".")

	configYAML := []byte(`
astring: test
scopes:
  my.scope:
    claims:
      - sub
`)

	require.NoError(t, ko.Load(rawbytes.Provider(configYAML), yaml.Parser()))

	final, err := koanfRemapKeys(val, ko, ds, nil)
	require.NoError(t, err)

	type combinedConf struct {
		testDeprecationsConf `koanf:",squash"`
		Scopes               map[string]testDottedScope `koanf:"scopes"`
	}

	conf := &combinedConf{}

	require.NoError(t, final.Unmarshal("", conf))

	assert.Equal(t, "test", conf.ANonSubItemString, "deprecation remap should still work")

	require.Contains(t, conf.Scopes, "my.scope", "dotted scope name should be preserved even with active deprecation remaps")
	assert.Equal(t, []string{"sub"}, conf.Scopes["my.scope"].Claims)

	assert.NotContains(t, conf.Scopes, "my", "dotted scope name should not be split")
}

func TestKoanfRemapKeys_SpecialCharactersInMapKeys(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		yamlKey  string
		expected string
	}{
		{"Colon", "urn:example:scope", "urn:example:scope", "urn:example:scope"},
		{"ColonInClaimURI", "http://example.com:8080/claim", "http://example.com:8080/claim", "http://example.com:8080/claim"},
		{"Dot", "my.scope", "my.scope", "my.scope"},
		{"MultipleDots", "org.example.scope.v2", "org.example.scope.v2", "org.example.scope.v2"},
		{"Tilde", "scope~draft", "scope~draft", "scope~draft"},
		{"Slash", "org/scope", "org/scope", "org/scope"},
		{"Underscore", "my_scope", "my_scope", "my_scope"},
		{"Hyphen", "my-scope", "my-scope", "my-scope"},
		{"MixedSpecialChars", "urn:ietf:params:oauth:scope:example.read", "urn:ietf:params:oauth:scope:example.read", "urn:ietf:params:oauth:scope:example.read"},
		{"URNStyle", "urn:authelia:scope:pam", "urn:authelia:scope:pam", "urn:authelia:scope:pam"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val := schema.NewStructValidator()

			ko := koanf.New(".")

			configYAML := []byte(fmt.Sprintf(`
scopes:
  '%s':
    claims:
      - sub
`, tc.yamlKey))

			require.NoError(t, ko.Load(rawbytes.Provider(configYAML), yaml.Parser()))

			final, err := koanfRemapKeys(val, ko, nil, nil)
			require.NoError(t, err)

			conf := &testDottedMapKeysConf{}

			require.NoError(t, final.Unmarshal("", conf))

			require.Contains(t, conf.Scopes, tc.expected, "map key %q should be preserved", tc.key)
			assert.Equal(t, []string{"sub"}, conf.Scopes[tc.expected].Claims)
		})
	}
}
