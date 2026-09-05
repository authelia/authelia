package oidc

import (
	"context"
	"time"
)

// GarbageCollectionFrequency returns the frequency the expired rows of the OAuth 2.0 DPoP tables should be collected
// at, or zero when DPoP is disabled and no collection is required. This implements the
// middlewares.GarbageCollectorProvider interface.
func (p *OpenIDConnectProvider) GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	if p == nil || p.Config == nil || !p.DPoP.Enabled {
		return 0
	}

	return frequencyGarbageCollectionOAuth2DPoP
}

// GarbageCollection removes the expired rows of the OAuth 2.0 DPoP tables. Both tables gain a row per request which is
// only meaningful until the value it records expires, so without this the tables grow for the lifetime of the
// deployment. This implements the middlewares.GarbageCollectorProvider interface.
func (p *OpenIDConnectProvider) GarbageCollection(ctx context.Context) (err error) {
	if p == nil || p.Store == nil {
		return nil
	}

	now := time.Now()

	if err = p.provider.DeleteExpiredOAuth2DPoPProofs(ctx, now); err != nil {
		return err
	}

	return p.provider.DeleteExpiredOAuth2DPoPNonces(ctx, now)
}
