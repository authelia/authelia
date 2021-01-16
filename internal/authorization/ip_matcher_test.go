package authorization

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestIPMatcher(t *testing.T) {
	// Default policy is 'allow all ips' if no IP is defined
	assert.True(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{}, schema.DefaultACLNetwork))

	assert.True(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{"127.0.0.1"}, schema.DefaultACLNetwork))
	assert.False(t, isIPMatching(net.ParseIP("127.1"), []string{"127.0.0.1"}, schema.DefaultACLNetwork))
	assert.False(t, isIPMatching(net.ParseIP("not-an-ip"), []string{"127.0.0.1"}, schema.DefaultACLNetwork))

	assert.False(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{"10.0.0.1"}, schema.DefaultACLNetwork))
	assert.False(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{"10.0.0.0/8"}, schema.DefaultACLNetwork))

	assert.True(t, isIPMatching(net.ParseIP("10.230.5.1"), []string{"10.0.0.0/8"}, schema.DefaultACLNetwork))
	assert.True(t, isIPMatching(net.ParseIP("10.230.5.1"), []string{"192.168.0.0/24", "10.0.0.0/8"}, schema.DefaultACLNetwork))

	// Test network groups
	assert.True(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{}, schema.DefaultACLNetwork))

	assert.True(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{"localhost"}, schema.DefaultACLNetwork))
	assert.False(t, isIPMatching(net.ParseIP("127.1"), []string{"localhost"}, schema.DefaultACLNetwork))
	assert.False(t, isIPMatching(net.ParseIP("not-an-ip"), []string{"localhost"}, schema.DefaultACLNetwork))

	assert.False(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{"internal"}, schema.DefaultACLNetwork))
	assert.False(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{"internal"}, schema.DefaultACLNetwork))

	assert.True(t, isIPMatching(net.ParseIP("10.230.5.1"), []string{"internal"}, schema.DefaultACLNetwork))
	assert.True(t, isIPMatching(net.ParseIP("10.230.5.1"), []string{"192.168.0.0/24", "internal"}, schema.DefaultACLNetwork))
}
