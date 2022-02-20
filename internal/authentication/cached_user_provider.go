package authentication

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
)

// MemoryCachedUserProvider is a provider that just caches the GetUserDetails method of another user provider.
type MemoryCachedUserProvider struct {
	provider UserProvider
	cache    map[string]CachedUserDetails
	config   schema.CachedAuthenticationBackendConfiguration
	log      *logrus.Logger

	mutex *sync.RWMutex
}

// NewMemoryCachedUserProvider returns a new MemoryCachedUserProvider.
func NewMemoryCachedUserProvider(config schema.CachedAuthenticationBackendConfiguration, p UserProvider) (provider *MemoryCachedUserProvider) {
	return &MemoryCachedUserProvider{
		provider: p,
		cache:    map[string]CachedUserDetails{},
		config:   config,
		log:      logging.Logger(),
		mutex:    &sync.RWMutex{},
	}
}

// CheckUserPassword returns the underlying UserProvider's CheckUserPassword method.
func (p *MemoryCachedUserProvider) CheckUserPassword(username string, password string) (valid bool, err error) {
	return p.provider.CheckUserPassword(username, password)
}

// GetCurrentDetails is a special method that is the same as GetDetails but it bypasses the cached data and updates it.
func (p *MemoryCachedUserProvider) GetCurrentDetails(username string) (details *model.UserDetails, err error) {
	if details, err = p.GetDetails(username); err != nil {
		return nil, err
	}

	p.mutex.Lock()

	p.cache[username] = CachedUserDetails{
		validUtil: time.Now().Add(*p.config.Duration),
		details:   details,
	}

	p.mutex.Unlock()

	return details, nil
}

// GetDetails checks the cache for details, if they don't exist or are expired we obtain them from the underlying
// UserProvider and return them to the user, otherwise we return the values from the cache.
func (p *MemoryCachedUserProvider) GetDetails(username string) (details *model.UserDetails, err error) {
	var (
		cache CachedUserDetails
		ok    bool
	)

	p.mutex.RLock()

	cache, ok = p.cache[username]

	p.mutex.RUnlock()

	if ok && cache.validUtil.After(time.Now()) {
		return cache.details, nil
	}

	if details, err = p.provider.GetDetails(username); err != nil {
		p.log.Errorf("Error occurred trying to update the user details cache: %+v", err)

		// If we don't have details in the cache for the user, we have to return nil/err here.
		if !ok || cache.details == nil {
			return nil, err
		}

		return cache.details, nil
	}

	p.mutex.Lock()

	p.cache[username] = CachedUserDetails{
		validUtil: time.Now().Add(*p.config.Duration),
		details:   details,
	}

	p.mutex.Unlock()

	return details, nil
}

// UpdatePassword returns the underlying UserProvider's UpdatePassword method.
func (p *MemoryCachedUserProvider) UpdatePassword(username string, newPassword string) (err error) {
	return p.provider.UpdatePassword(username, newPassword)
}

// StartupCheck returns the underlying UserProvider's StartupCheck method.
func (p *MemoryCachedUserProvider) StartupCheck() (err error) {
	return p.provider.StartupCheck()
}
