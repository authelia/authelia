//nolint:all
package memory

import (
	"time"
)

// Config provider settings
type Config struct{}

// Provider backend manager
type Provider struct {
	config Config
	db     Map
}

type item struct {
	data           []byte
	lastActiveTime int64
	expiration     time.Duration
}
