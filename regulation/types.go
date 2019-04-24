package regulation

import (
	"time"

	"github.com/clems4ever/authelia/storage"
)

// Regulator an authentication regulator preventing attackers to brute force the service.
type Regulator struct {
	// Is the regulation enabled.
	enabled bool
	// The number of failed authentication attempt before banning the user
	maxRetries int
	// If a user does the max number of retries within that duration, she will be banned.
	findTime time.Duration
	// If a user has been banned, this duration is the timelapse during which the user is banned.
	banTime time.Duration

	storageProvider storage.Provider
}
