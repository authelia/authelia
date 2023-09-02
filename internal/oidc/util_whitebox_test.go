package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSigningAlgLess(t *testing.T) {
	assert.False(t, isSigningAlgLess(SigningAlgRSAUsingSHA256, SigningAlgRSAUsingSHA256))
	assert.False(t, isSigningAlgLess(SigningAlgRSAUsingSHA256, SigningAlgHMACUsingSHA256))
	assert.True(t, isSigningAlgLess(SigningAlgHMACUsingSHA256, SigningAlgNone))
	assert.True(t, isSigningAlgLess(SigningAlgHMACUsingSHA256, SigningAlgRSAUsingSHA512))
	assert.True(t, isSigningAlgLess(SigningAlgHMACUsingSHA256, SigningAlgRSAPSSUsingSHA256))
	assert.True(t, isSigningAlgLess(SigningAlgHMACUsingSHA256, SigningAlgECDSAUsingP521AndSHA512))
	assert.True(t, isSigningAlgLess(SigningAlgRSAUsingSHA256, SigningAlgECDSAUsingP521AndSHA512))
	assert.True(t, isSigningAlgLess(SigningAlgECDSAUsingP521AndSHA512, "JS121"))
	assert.False(t, isSigningAlgLess("JS121", SigningAlgECDSAUsingP521AndSHA512))
	assert.False(t, isSigningAlgLess("JS121", "TS512"))
}
