package authorization

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldNotParseInvalidSubjects(t *testing.T) {
	subjectsSchema := [][]string{{"groups:z"}, {"group:z", "users:b"}}
	subjectsACL := schemaSubjectsToACL(subjectsSchema)

	require.Len(t, subjectsACL, 1)

	require.Len(t, subjectsACL[0].Subjects, 1)

	assert.True(t, subjectsACL[0].IsMatch(Subject{Username: "a", Groups: []string{"z"}}))
}

func TestShouldSplitDomainCorrectly(t *testing.T) {
	prefix, suffix := domainToPrefixSuffix("apple.example.com")

	assert.Equal(t, "apple", prefix)
	assert.Equal(t, "example.com", suffix)

	prefix, suffix = domainToPrefixSuffix("example")

	assert.Equal(t, "", prefix)
	assert.Equal(t, "example", suffix)
}

func TestShouldParseNetworks(t *testing.T) {
	schemaNetworks := []schema.ACLNetwork{
		{
			Name: "test",
			Networks: []string{
				"10.0.0.1",
			},
		},
		{
			Name: "second",
			Networks: []string{
				"10.0.0.1",
			},
		},
		{
			Name: "duplicate",
			Networks: []string{
				"10.0.0.1",
			},
		},
		{
			Name: "duplicate",
			Networks: []string{
				"10.0.0.1",
			},
		},
		{
			Name: "ipv6",
			Networks: []string{
				"fec0::1",
			},
		},
		{
			Name: "ipv6net",
			Networks: []string{
				"fec0::1/64",
			},
		},
		{
			Name: "net",
			Networks: []string{
				"10.0.0.0/8",
			},
		},
		{
			Name: "badnet",
			Networks: []string{
				"bad/8",
			},
		},
	}

	_, firstNetwork, err := net.ParseCIDR("10.0.0.1/32")
	require.NoError(t, err)

	_, secondNetwork, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	_, thirdNetwork, err := net.ParseCIDR("fec0::1/64")
	require.NoError(t, err)

	_, fourthNetwork, err := net.ParseCIDR("fec0::1/128")
	require.NoError(t, err)

	networksMap, networksCacheMap := parseSchemaNetworks(schemaNetworks)

	require.Len(t, networksMap, 6)
	require.Contains(t, networksMap, "test")
	require.Contains(t, networksMap, "second")
	require.Contains(t, networksMap, "duplicate")
	require.Contains(t, networksMap, "ipv6")
	require.Contains(t, networksMap, "ipv6net")
	require.Contains(t, networksMap, "net")
	require.Len(t, networksMap["test"], 1)

	require.Len(t, networksCacheMap, 7)
	require.Contains(t, networksCacheMap, "10.0.0.1")
	require.Contains(t, networksCacheMap, "10.0.0.1/32")
	require.Contains(t, networksCacheMap, "10.0.0.1/32")
	require.Contains(t, networksCacheMap, "10.0.0.0/8")
	require.Contains(t, networksCacheMap, "fec0::1")
	require.Contains(t, networksCacheMap, "fec0::1/128")
	require.Contains(t, networksCacheMap, "fec0::1/64")

	assert.Equal(t, firstNetwork, networksMap["test"][0])
	assert.Equal(t, secondNetwork, networksMap["net"][0])
	assert.Equal(t, thirdNetwork, networksMap["ipv6net"][0])
	assert.Equal(t, fourthNetwork, networksMap["ipv6"][0])

	assert.Equal(t, firstNetwork, networksCacheMap["10.0.0.1"])
	assert.Equal(t, firstNetwork, networksCacheMap["10.0.0.1/32"])

	assert.Equal(t, secondNetwork, networksCacheMap["10.0.0.0/8"])

	assert.Equal(t, thirdNetwork, networksCacheMap["fec0::1/64"])

	assert.Equal(t, fourthNetwork, networksCacheMap["fec0::1"])
	assert.Equal(t, fourthNetwork, networksCacheMap["fec0::1/128"])
}
