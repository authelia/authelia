package trust

import (
	"crypto/x509"
)

type Opt func(provider *Production)

func WithPaths(paths ...string) Opt {
	return func(provider *Production) {
		provider.config.Paths = paths
	}
}

func WithSystem(system bool) Opt {
	return func(provider *Production) {
		provider.config.System = system
	}
}

func WithInvalid(invalid bool) Opt {
	return func(provider *Production) {
		provider.config.Invalid = invalid
	}
}

func WithExpired(expired bool) Opt {
	return func(provider *Production) {
		provider.config.Expired = expired
	}
}

func WithFuture(future bool) Opt {
	return func(provider *Production) {
		provider.config.Future = future
	}
}

func WithStatic(static []*x509.Certificate) Opt {
	return func(provider *Production) {
		provider.config.Static = static
	}
}
