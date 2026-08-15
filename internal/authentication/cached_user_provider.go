package authentication

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

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

type CachedUserProvider struct {
	UserProvider

	lifespan schema.RefreshIntervalDuration

	details  CachedUserDetails
	extended CachedUserDetailsExtended
}

// GetDetails retrieves the user's information skipping the cache but ensuring the cache is updated.
func (p *CachedUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	if details, err = p.UserProvider.GetDetails(username); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			p.invalidate(username)
		}

		return nil, err
	}

	p.details.Lock()

	p.details.values[username] = CachedUserDetailsItem{UserDetails: details, expires: time.Now().Add(p.lifespan.Value())}

	p.details.Unlock()

	p.invalidateExtended(username)

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

		p.details.values[username] = CachedUserDetailsItem{UserDetails: details, expires: now.Add(p.lifespan.Value())}

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

	p.extended.Lock()

	p.extended.values[username] = CachedUserDetailsExtendedItem{UserDetailsExtended: details, expires: expires}

	p.extended.Unlock()

	if details.UserDetails != nil {
		p.details.Lock()

		p.details.values[username] = CachedUserDetailsItem{UserDetails: details.UserDetails, expires: expires}

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

		p.extended.values[username] = CachedUserDetailsExtendedItem{UserDetailsExtended: details, expires: now.Add(p.lifespan.Value())}

		p.extended.Unlock()

		return details, nil
	}
}

func (p *CachedUserProvider) invalidate(username string) {
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

type CachedUserDetails struct {
	singleflight.Group
	sync.Mutex

	values map[string]CachedUserDetailsItem
}

type CachedUserDetailsExtended struct {
	singleflight.Group
	sync.Mutex

	values map[string]CachedUserDetailsExtendedItem
}

type CachedUserDetailsItem struct {
	*UserDetails

	expires time.Time
}

type CachedUserDetailsExtendedItem struct {
	*UserDetailsExtended

	expires time.Time
}
