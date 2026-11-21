// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authentication

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync/atomic"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewCachedUserProvider returns a UserProvider which caches the details of the given provider for the given lifespan.
// The provider is returned unwrapped when the refresh interval requires the details are always retrieved fresh.
func NewCachedUserProvider(provider UserProvider, lifespan schema.RefreshIntervalDuration) UserProvider {
	if lifespan.Always() {
		return provider
	}

	return &CachedUserProvider{
		UserProvider: provider,
		lifespan:     lifespan,
		details:      CachedUserValues[UserDetails]{values: map[string]CachedUserValue[UserDetails]{}},
		extended:     CachedUserValues[UserDetailsExtended]{values: map[string]CachedUserValue[UserDetailsExtended]{}},
	}
}

// CachedUserProvider is a UserProvider which decorates another UserProvider with an expiring cache of the user details
// it retrieves, collapsing concurrent retrievals of the same user into a single backend call.
type CachedUserProvider struct {
	UserProvider

	lifespan schema.RefreshIntervalDuration

	details  CachedUserValues[UserDetails]
	extended CachedUserValues[UserDetailsExtended]

	generation atomic.Uint64
}

// GetDetails retrieves the user's information skipping the cache but ensuring the cache is updated.
func (p *CachedUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	if details, err = p.UserProvider.GetDetails(username); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			p.invalidate(username)
		}

		return nil, err
	}

	key := canonicalUsername(username, details)

	p.retire(username, key)

	p.details.set(key, details, time.Now().Add(p.lifespan.Value()), nil)
	p.extended.delete(key)

	return details, nil
}

// GetDetailsCached retrieves the user's information from the cache if available.
func (p *CachedUserProvider) GetDetailsCached(username string) (details *UserDetails, err error) {
	return getCached(p, &p.details, username, p.UserProvider.GetDetails, identityUserDetails, cloneUserDetails)
}

// GetDetailsExtended retrieves the user's extended information skipping the cache but ensuring the cache is updated.
func (p *CachedUserProvider) GetDetailsExtended(username string) (details *UserDetailsExtended, err error) {
	if details, err = p.UserProvider.GetDetailsExtended(username); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			p.invalidate(username)
		}

		return nil, err
	}

	expires := time.Now().Add(p.lifespan.Value())
	key := canonicalUsername(username, details.UserDetails)

	p.retire(username, key)

	p.extended.set(key, details, expires, nil)

	if details.UserDetails != nil {
		p.details.set(key, details.UserDetails, expires, nil)
	}

	return details, nil
}

// GetDetailsExtendedCached retrieves the user's extended information from the cache if available.
func (p *CachedUserProvider) GetDetailsExtendedCached(username string) (details *UserDetailsExtended, err error) {
	return getCached(p, &p.extended, username, p.UserProvider.GetDetailsExtended, extendedUserDetails, cloneUserDetailsExtended)
}

func getCached[T any](p *CachedUserProvider, cache *CachedUserValues[T], username string, fetch func(string) (*T, error), details func(*T) *UserDetails, clone func(*T) *T) (value *T, err error) {
	result, err, _ := cache.Do(username, func() (any, error) {
		now, generation := time.Now(), p.generation.Load()

		if cached, ok := cache.get(username, now, p.lifespan.Never()); ok {
			return cached, nil
		}

		fetched, err := fetch(username)
		if err != nil {
			return nil, err
		}

		cache.set(canonicalUsername(username, details(fetched)), fetched, now.Add(p.lifespan.Value()), func() bool {
			return p.generation.Load() == generation
		})

		return fetched, nil
	})
	if err != nil {
		return nil, err
	}

	if value, _ = result.(*T); value == nil {
		return nil, fmt.Errorf("error occurred retrieving user details from cache for user '%s'", username)
	}

	return clone(value), nil
}

func (c *CachedUserValues[T]) get(key string, now time.Time, never bool) (value *T, ok bool) {
	c.Lock()

	defer c.Unlock()

	item, ok := c.values[key]
	if ok && (never || item.expires.After(now)) {
		return item.value, true
	}

	delete(c.values, key)

	return nil, false
}

func (c *CachedUserValues[T]) set(key string, value *T, expires time.Time, valid func() bool) {
	c.Lock()

	defer c.Unlock()

	if valid == nil || valid() {
		c.values[key] = CachedUserValue[T]{value: value, expires: expires}
	}
}

func (c *CachedUserValues[T]) delete(key string) {
	c.Lock()

	defer c.Unlock()

	delete(c.values, key)
}

func (p *CachedUserProvider) retire(usernames ...string) {
	p.generation.Add(1)

	for _, username := range usernames {
		p.details.Forget(username)
		p.extended.Forget(username)
	}
}

func (p *CachedUserProvider) invalidate(username string) {
	p.retire(username)

	p.details.delete(username)
	p.extended.delete(username)
}

func canonicalUsername(input string, details *UserDetails) string {
	if details == nil || details.Username == "" {
		return input
	}

	return details.Username
}

func identityUserDetails(details *UserDetails) *UserDetails {
	return details
}

func extendedUserDetails(details *UserDetailsExtended) *UserDetails {
	return details.UserDetails
}

func cloneUserDetails(details *UserDetails) *UserDetails {
	if details == nil {
		return nil
	}

	clone := *details

	clone.Emails = slices.Clone(details.Emails)
	clone.Groups = slices.Clone(details.Groups)

	return &clone
}

func cloneUserDetailsExtended(details *UserDetailsExtended) *UserDetailsExtended {
	clone := *details

	clone.Extra = maps.Clone(details.Extra)
	clone.UserDetails = cloneUserDetails(details.UserDetails)

	return &clone
}
