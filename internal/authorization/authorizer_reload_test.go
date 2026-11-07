// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authorization

import (
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func reloadTestACL(policy string) schema.AccessControl {
	return schema.AccessControl{
		DefaultPolicy: deny,
		Rules: []schema.AccessControlRule{
			{Domains: []string{"app.example.com"}, Policy: policy},
		},
	}
}

func reloadTestObject() Object {
	return NewObjectRaw(&url.URL{Scheme: "https", Host: "app.example.com", Path: "/"}, []byte("GET"))
}

func TestAuthorizerUpdateSwitchesNewEvaluations(t *testing.T) {
	authorizer := NewAuthorizer(&schema.Configuration{AccessControl: reloadTestACL(oneFactor)})

	assert.Equal(t, uint64(1), authorizer.Generation())

	_, level := authorizer.GetRequiredLevel(Subject{}, reloadTestObject())
	assert.Equal(t, OneFactor, level)
	assert.False(t, authorizer.IsSecondFactorEnabled())

	assert.Equal(t, uint64(2), authorizer.Update(reloadTestACL(twoFactor)))

	_, level = authorizer.GetRequiredLevel(Subject{}, reloadTestObject())
	assert.Equal(t, TwoFactor, level)
	assert.True(t, authorizer.IsSecondFactorEnabled())
	assert.Equal(t, uint64(2), authorizer.Generation())
}

func TestAuthorizerSnapshotIsPinnedAcrossUpdate(t *testing.T) {
	authorizer := NewAuthorizer(&schema.Configuration{AccessControl: reloadTestACL(oneFactor)})

	snapshot := authorizer.Snapshot()

	authorizer.Update(reloadTestACL(twoFactor))

	_, old := snapshot.GetRequiredLevel(Subject{}, reloadTestObject())
	_, current := authorizer.GetRequiredLevel(Subject{}, reloadTestObject())

	assert.Equal(t, OneFactor, old, "in-flight request must keep the generation it started with")
	assert.Equal(t, TwoFactor, current, "new requests must use the new generation")

	assert.False(t, snapshot.IsSecondFactorEnabled())
	assert.True(t, authorizer.IsSecondFactorEnabled())
	assert.Equal(t, uint64(1), snapshot.Generation())
	assert.Equal(t, uint64(2), authorizer.Generation())

	// Updating a pinned authorizer is a no-op and must not affect the source.
	assert.Equal(t, uint64(1), snapshot.Update(reloadTestACL(bypass)))
	assert.Equal(t, uint64(2), authorizer.Generation())

	// A snapshot of a snapshot stays on the same generation.
	assert.Equal(t, uint64(1), snapshot.Snapshot().Generation())
}

func TestAuthorizerConcurrentUpdateAndEvaluate(t *testing.T) {
	authorizer := NewAuthorizer(&schema.Configuration{AccessControl: reloadTestACL(oneFactor)})

	var wg sync.WaitGroup

	stop := make(chan struct{})

	for i := 0; i < 8; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-stop:
					return
				default:
					snapshot := authorizer.Snapshot()

					_, first := snapshot.GetRequiredLevel(Subject{}, reloadTestObject())
					_, second := snapshot.GetRequiredLevel(Subject{}, reloadTestObject())

					if first != second {
						t.Errorf("snapshot returned different levels within one request: %v != %v", first, second)

						return
					}
				}
			}
		}()
	}

	for i := 0; i < 200; i++ {
		if i%2 == 0 {
			authorizer.Update(reloadTestACL(twoFactor))
		} else {
			authorizer.Update(reloadTestACL(oneFactor))
		}
	}

	close(stop)
	wg.Wait()

	assert.Equal(t, uint64(201), authorizer.Generation())
}

func TestAuthorizerUpdateKeepsOpenIDConnectMFA(t *testing.T) {
	config := &schema.Configuration{AccessControl: reloadTestACL(oneFactor)}
	config.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{
		Clients: []schema.IdentityProvidersOpenIDConnectClient{{ID: "test", AuthorizationPolicy: twoFactor}},
	}

	authorizer := NewAuthorizer(config)

	assert.True(t, authorizer.IsSecondFactorEnabled())

	authorizer.Update(reloadTestACL(oneFactor))

	assert.True(t, authorizer.IsSecondFactorEnabled(), "OIDC two-factor requirement is static and must survive an update")
}
