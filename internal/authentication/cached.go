package authentication

import (
	"crypto/hmac"
	"crypto/rand"
	"fmt"
	"hash"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// NewCredentialCacheHMAC creates a new CredentialCacheHMAC with a given hash.Hash func and lifespan.
func NewCredentialCacheHMAC(h func() hash.Hash, lifespan time.Duration) *CredentialCacheHMAC {
	secret := make([]byte, h().BlockSize())

	_, _ = rand.Read(secret)

	return &CredentialCacheHMAC{
		mu:       sync.Mutex{},
		hash:     hmac.New(h, secret),
		lifespan: lifespan,

		values: map[string]CachedCredential{},
	}
}

// CredentialCacheHMAC implements in-memory credential caching using a HMAC function and effective lifespan.
type CredentialCacheHMAC struct {
	mu    sync.Mutex
	hash  hash.Hash
	group singleflight.Group

	lifespan time.Duration

	values map[string]CachedCredential
}

func (c *CredentialCacheHMAC) Check(ctx Context, username, password string) (valid, cached bool, err error) {
	var (
		key string
		sum []byte
	)

	if key, sum, err = c.sum(username, password); err != nil {
		return false, false, err
	}

	var raw any

	ctx.GetLogger().WithFields(map[string]any{"group": "basic-cache", "key": key, "username": username}).Trace("Attempting Single Flighted Check")

	raw, err, _ = c.group.Do(key, c.check(ctx, username, password, key, sum))

	result := raw.(*FlightResult)

	return result.Valid, result.Cached, err
}

func (c *CredentialCacheHMAC) sum(username, password string) (hex string, sum []byte, err error) {
	c.mu.Lock()

	defer c.mu.Unlock()

	defer c.hash.Reset()

	if _, err = c.hash.Write([]byte(password)); err != nil {
		return "", nil, fmt.Errorf("error occurred calculating cache hmac: %w", err)
	}

	if _, err = c.hash.Write([]byte(username)); err != nil {
		return "", nil, fmt.Errorf("error occurred calculating cache hmac: %w", err)
	}

	sum = c.hash.Sum(nil)

	return fmt.Sprintf("%x", sum), sum, nil
}

func (c *CredentialCacheHMAC) check(ctx Context, username, password, key string, sum []byte) func() (value any, err error) {
	return func() (value any, err error) {
		ctx.GetLogger().WithFields(map[string]any{"group": "basic-cache", "key": key, "username": username}).Trace("Performing Single Flighted Check")

		var match, valid bool

		if match, _ = c.valid(ctx, username, sum); match {
			return &FlightResult{Cached: true, Valid: true}, nil
		}

		if valid, err = ctx.GetUserProvider().CheckUserPassword(username, password); err != nil {
			return &FlightResult{Cached: false, Valid: valid}, err
		}

		if valid {
			if err = c.put(ctx, username, sum); err != nil {
				ctx.GetLogger().WithError(err).Errorf("Error occurred saving basic authorization credentials to cache for user '%s'", username)
			}

			return &FlightResult{Cached: false, Valid: valid}, nil
		}

		return &FlightResult{Cached: false, Valid: valid}, nil
	}
}

func (c *CredentialCacheHMAC) valid(ctx Context, username string, value []byte) (valid, ok bool) {
	var (
		entry CachedCredential
	)

	c.mu.Lock()

	defer c.mu.Unlock()

	if entry, ok = c.values[username]; ok {
		if entry.expires.Before(ctx.GetClock().Now()) {
			delete(c.values, username)

			return false, false
		}
	}

	valid = hmac.Equal(value, entry.value)

	return valid, true
}

func (c *CredentialCacheHMAC) put(ctx Context, username string, value []byte) (err error) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.values[username] = CachedCredential{expires: ctx.GetClock().Now().Add(c.lifespan), value: value}

	return nil
}

type FlightResult struct {
	Valid  bool
	Cached bool
}

// CachedCredential is a cached credential which has an expiration and checksum value.
type CachedCredential struct {
	expires time.Time
	value   []byte
}
