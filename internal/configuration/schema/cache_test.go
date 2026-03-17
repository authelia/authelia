package schema

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"testing"

	"go.yaml.in/yaml/v4"
)

func Test(t *testing.T) {
	v := &Cache{
		Redis: &RedisCache{
			Address: &AddressTCP{Address{
				valid:  true,
				socket: false,
				umask:  0,
				port:   6379,
				fd:     nil,
				url:    &url.URL{Scheme: "tcp", Host: "redis:6379"},
			}},
			Database: 0,
			Username: "authelia",
			Password: "password",
			TLS: &TLS{
				MinimumVersion: TLSVersion{tls.VersionTLS12},
				MaximumVersion: TLSVersion{tls.VersionTLS13},
				SkipVerify:     false,
				ServerName:     "redis",
			},
			DialTimeout:                0,
			ReadTimeout:                0,
			WriteTimeout:               0,
			ConnectionTimeout:          0,
			IdleTimeout:                0,
			PoolTimeout:                0,
			MinimumRetryBackoff:        0,
			MaximumRetryBackoff:        0,
			MaximumRetries:             0,
			PoolSize:                   0,
			PoolMinimumIdleConnections: 0,
			PoolMaximumIdleConnections: 0,
			PoolMaximumConnections:     0,
		},
		RedisSentinel: &RedisSentinelCache{
			MasterName:       "sentinel",
			SentinelUsername: "authelia-sentinel",
			SentinelPassword: "password-sentinel",
			Addresses: []*AddressTCP{
				{
					Address{
						valid:  true,
						socket: false,
						umask:  0,
						port:   26379,
						fd:     nil,
						url:    &url.URL{Scheme: "tcp", Host: "sentinel:26379"},
					},
				},
			},
			RouteByReplica: false,
			RouteByLatency: false,
			RouteRandomly:  false,
			Database:       0,
			Username:       "authelia",
			Password:       "password",
			TLS: &TLS{
				MinimumVersion: TLSVersion{tls.VersionTLS12},
				MaximumVersion: TLSVersion{tls.VersionTLS13},
			},
			DialTimeout:                0,
			ReadTimeout:                0,
			WriteTimeout:               0,
			ConnectionTimeout:          0,
			IdleTimeout:                0,
			PoolTimeout:                0,
			MinimumRetryBackoff:        0,
			MaximumRetryBackoff:        0,
			MaximumRetries:             0,
			PoolSize:                   0,
			PoolMinimumIdleConnections: 0,
			PoolMaximumIdleConnections: 0,
			PoolMaximumConnections:     0,
		},
		RedisCluster: &RedisClusterCache{
			Addresses: []*AddressTCP{
				{
					Address{
						valid:  true,
						socket: false,
						umask:  0,
						port:   6379,
						fd:     nil,
						url:    &url.URL{Scheme: "tcp", Host: "redis-cluster:6379"},
					},
				},
			},
			RouteByReplica: false,
			RouteByLatency: false,
			RouteRandomly:  false,
			Username:       "authelia",
			Password:       "password",
			TLS: &TLS{
				MinimumVersion: TLSVersion{tls.VersionTLS12},
				MaximumVersion: TLSVersion{tls.VersionTLS13},
				SkipVerify:     false,
				ServerName:     "redis-cluster",
			},
			DialTimeout:                0,
			ReadTimeout:                0,
			WriteTimeout:               0,
			ConnectionTimeout:          0,
			IdleTimeout:                0,
			PoolTimeout:                0,
			MinimumRetryBackoff:        0,
			MaximumRetryBackoff:        0,
			MaximumRedirects:           0,
			PoolSize:                   0,
			PoolMinimumIdleConnections: 0,
			PoolMaximumIdleConnections: 0,
			PoolMaximumConnections:     0,
		},
	}

	data, err := yaml.Marshal(v)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))

}
