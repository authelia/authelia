package authorization

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPMatcher(t *testing.T) {
	// Default policy is 'allow all ips' if no IP is defined
	assert.True(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{}))

	assert.True(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{"127.0.0.1"}))
	assert.False(t, isIPMatching(net.ParseIP("127.1"), []string{"127.0.0.1"}))
	assert.False(t, isIPMatching(net.ParseIP("not-an-ip"), []string{"127.0.0.1"}))

	assert.False(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{"10.0.0.1"}))
	assert.False(t, isIPMatching(net.ParseIP("127.0.0.1"), []string{"10.0.0.0/8"}))

	assert.True(t, isIPMatching(net.ParseIP("10.230.5.1"), []string{"10.0.0.0/8"}))
	assert.True(t, isIPMatching(net.ParseIP("10.230.5.1"), []string{"192.168.0.0/24", "10.0.0.0/8"}))
}
