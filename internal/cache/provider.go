package cache

import (
	"context"
	"time"

	"github.com/authelia/authelia/v4/internal/model"
)

type Provider interface {
	model.StartupCheck

	GetSession(ctx context.Context, id []byte, issuer string) (data []byte, err error)
	SaveSession(ctx context.Context, id, data []byte, issuer string, expiration time.Duration) (err error)
	DeleteSession(ctx context.Context, id []byte, issuer string) (err error)
	RegenerateSession(ctx context.Context, oldID, id []byte, issuer string, expiration time.Duration) (err error)
	GarbageCollectSessions(ctx context.Context) (err error)

	GetCachedValue(ctx context.Context, category, lookup string, encrypted bool) (data []byte, err error)
	SetCacheValue(ctx context.Context, category, lookup string, data []byte, encrypted bool, expiration time.Duration) (err error)
}
