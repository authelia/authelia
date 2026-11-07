// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authentication

import (
	"errors"
	"fmt"
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
		details: CachedUserDetails{
			values: map[string]CachedUserDetailsItem{},
		},
		extended: CachedUserDetailsExtended{
			values: map[string]CachedUserDetailsExtendedItem{},
		},
	}
}

// CachedUserProvider is a UserProvider which decorates another UserProvider with an expiring cache of the user details
// it retrieves, collapsing concurrent retrievals of the same user into a single backend call.
type CachedUserProvider struct {
	UserProvider

	lifespan schema.RefreshIntervalDuration

	details  CachedUserDetails
	extended CachedUserDetailsExtended

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

	p.details.Lock()

	p.details.values[key] = CachedUserDetailsItem{UserDetails: details, expires: time.Now().Add(p.lifespan.Value())}

	p.details.Unlock()

	p.invalidateExtended(key)

	return details, nil
}

// GetDetailsCached retrieves the user's information from the cache if available.
func (p *CachedUserProvider) GetDetailsCached(username string) (details *UserDetails, err error) {
	var (
		result any
		ok     bool
	)

	if result, err, _ = p.details.Do(username, p.doGetDetailsCached(username)); err != nil {
		return nil, err
	}

	if details, ok = result.(*UserDetails); !ok {
		return nil, fmt.Errorf("error occurred retrieving user details from cache for user '%s'", username)
	}

	return details, err
}

func (p *CachedUserProvider) doGetDetailsCached(username string) func() (result any, err error) {
	return func() (result any, err error) {
		now := time.Now()
		generation := p.generation.Load()

		p.details.Lock()

		if cached, ok := p.details.values[username]; ok && (p.lifespan.Never() || cached.expires.After(now)) {
			p.details.Unlock()

			cp := *cached.UserDetails

			return &cp, nil
		} else {
			delete(p.details.values, username)
		}

		p.details.Unlock()

		var details *UserDetails

		if details, err = p.UserProvider.GetDetails(username); err != nil {
			return nil, err
		}

		p.details.Lock()

		if p.generation.Load() == generation {
			p.details.values[canonicalUsername(username, details)] = CachedUserDetailsItem{UserDetails: details, expires: now.Add(p.lifespan.Value())}
		}

		p.details.Unlock()

		return details, nil
	}
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

	p.extended.Lock()

	p.extended.values[key] = CachedUserDetailsExtendedItem{UserDetailsExtended: details, expires: expires}

	p.extended.Unlock()

	if details.UserDetails != nil {
		p.details.Lock()

		p.details.values[key] = CachedUserDetailsItem{UserDetails: details.UserDetails, expires: expires}

		p.details.Unlock()
	}

	return details, nil
}

// GetDetailsExtendedCached retrieves the user's extended information from the cache if available.
func (p *CachedUserProvider) GetDetailsExtendedCached(username string) (details *UserDetailsExtended, err error) {
	var (
		result any
		ok     bool
	)

	if result, err, _ = p.extended.Do(username, p.doGetDetailsExtendedCached(username)); err != nil {
		return nil, err
	}

	if details, ok = result.(*UserDetailsExtended); !ok {
		return nil, fmt.Errorf("error occurred retrieving user details from cache for user '%s'", username)
	}

	return details, err
}

func (p *CachedUserProvider) doGetDetailsExtendedCached(username string) func() (result any, err error) {
	return func() (result any, err error) {
		now := time.Now()
		generation := p.generation.Load()

		p.extended.Lock()

		if cached, ok := p.extended.values[username]; ok && (p.lifespan.Never() || cached.expires.After(now)) {
			p.extended.Unlock()

			cp := *cached.UserDetailsExtended

			return &cp, nil
		} else {
			delete(p.extended.values, username)
		}

		p.extended.Unlock()

		var details *UserDetailsExtended

		if details, err = p.UserProvider.GetDetailsExtended(username); err != nil {
			return nil, err
		}

		p.extended.Lock()

		if p.generation.Load() == generation {
			p.extended.values[canonicalUsername(username, details.UserDetails)] = CachedUserDetailsExtendedItem{UserDetailsExtended: details, expires: now.Add(p.lifespan.Value())}
		}

		p.extended.Unlock()

		return details, nil
	}
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

	p.details.Lock()

	delete(p.details.values, username)

	p.details.Unlock()

	p.invalidateExtended(username)
}

func (p *CachedUserProvider) invalidateExtended(username string) {
	p.extended.Lock()

	delete(p.extended.values, username)

	p.extended.Unlock()
}

func canonicalUsername(input string, details *UserDetails) string {
	if details == nil || details.Username == "" {
		return input
	}

	return details.Username
}
