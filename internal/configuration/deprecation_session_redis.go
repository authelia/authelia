// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func mapSessionRedisToCache(d MultiKeyMappedDeprecation, keys map[string]any, val *schema.StructValidator) {
	sentinel := isSessionRedisSentinel(keys)

	prefix := keyCacheRedis
	if sentinel {
		prefix = keyCacheRedisSentinel
	}

	if mapHasKeyPrefix(keyCacheRedis, keys) || mapHasKeyPrefix(keyCacheRedisSentinel, keys) || mapHasKeyPrefix(keyCacheRedisCluster, keys) {
		val.Push(fmt.Errorf(errFmtSessionRedisConflict, d.Version, prefix))

		deleteSessionRedisKeys(keys)

		return
	}

	if err := mapSessionRedisAddresses(prefix, sentinel, keys); err != nil {
		val.Push(fmt.Errorf(errFmtSessionRedisMapErr, prefix, err))

		deleteSessionRedisKeys(keys)

		return
	}

	if sentinel {
		if value, ok := keys[keySessionRedisHAName]; ok {
			keys[prefix+".master_name"] = value
		}
	}

	keys[prefix+".idle_timeout"] = time.Second * 300

	for old, new := range deprecationSessionRedisOptions {
		value, ok := keys[keySessionRedis+"."+old]
		if !ok {
			continue
		}

		if !sentinel && strings.HasPrefix(old, "high_availability.") {
			continue
		}

		keys[prefix+"."+new] = value
	}

	if _, ok := keys[keySessionStorage]; !ok {
		keys[keySessionStorage] = "cache"
	}

	val.PushWarning(fmt.Errorf(errFmtSessionRedisMapped, d.Version, prefix, d.Version.NextMajor()))

	deleteSessionRedisKeys(keys)
}

func isSessionRedisSentinel(keys map[string]any) bool {
	value, ok := keys[keySessionRedisHAName]
	if !ok {
		return false
	}

	name, ok := value.(string)

	return ok && name != ""
}

func mapSessionRedisAddresses(prefix string, sentinel bool, keys map[string]any) (err error) {
	host, _ := keys[keySessionRedisHost].(string)

	var raw int

	if value, ok := keys[keySessionRedisPort]; ok {
		if raw, err = sessionRedisPort(value); err != nil {
			return err
		}
	}

	var port uint16

	if !sentinel {
		if host == "" {
			return nil
		}

		if !path.IsAbs(host) {
			if port, err = sessionRedisPortDefault(raw, schema.DefaultRedisCachePort); err != nil {
				return err
			}
		}

		address, err := schema.NewAddressFromNetworkValuesDefault(host, port, schema.AddressSchemeTCP, schema.AddressSchemeUnix)
		if err != nil {
			return err
		}

		keys[prefix+".address"] = address.String()

		return nil
	}

	if port, err = sessionRedisPortDefault(raw, schema.DefaultRedisSentinelCachePort); err != nil {
		return err
	}

	var addresses []any

	if host != "" {
		addresses = append(addresses, sessionRedisAddress(host, port))
	}

	if addresses, err = sessionRedisSentinelNodeAddresses(addresses, keys); err != nil {
		return err
	}

	if len(addresses) != 0 {
		keys[prefix+".addresses"] = addresses
	}

	return nil
}

func sessionRedisSentinelNodeAddresses(addresses []any, keys map[string]any) (result []any, err error) {
	nodes, _ := keys[keySessionRedisHANodes].([]any)

	for _, node := range nodes {
		m, ok := node.(map[string]any)
		if !ok {
			continue
		}

		nodeHost, _ := m["host"].(string)
		if nodeHost == "" {
			continue
		}

		var raw int

		if value, ok := m["port"]; ok {
			if raw, err = sessionRedisPort(value); err != nil {
				return nil, err
			}
		}

		var nodePort uint16

		if nodePort, err = sessionRedisPortDefault(raw, schema.DefaultRedisSentinelCachePort); err != nil {
			return nil, err
		}

		address := sessionRedisAddress(nodeHost, nodePort)

		if !anySliceContains(addresses, address) {
			addresses = append(addresses, address)
		}
	}

	return addresses, nil
}

func sessionRedisAddress(host string, port uint16) string {
	address := schema.NewAddressFromNetworkValues(schema.AddressSchemeTCP, strings.ToLower(host), port)

	return address.String()
}

func sessionRedisPort(value any) (port int, err error) {
	if err = mapstructure.WeakDecode(value, &port); err != nil {
		return 0, fmt.Errorf("error occurred decoding the port: %w", err)
	}

	return port, nil
}

func sessionRedisPortDefault(port int, fallback uint16) (uint16, error) {
	switch {
	case port == 0:
		return fallback, nil
	case port < 1 || port > 65535:
		return 0, fmt.Errorf("port %d is not a valid port number", port)
	default:
		return uint16(port), nil
	}
}

func anySliceContains(values []any, value string) bool {
	for _, v := range values {
		if s, ok := v.(string); ok && s == value {
			return true
		}
	}

	return false
}

func mapHasKeyPrefix(prefix string, keys map[string]any) bool {
	for key := range keys {
		if key == prefix || strings.HasPrefix(key, prefix+".") {
			return true
		}
	}

	return false
}

func deleteSessionRedisKeys(keys map[string]any) {
	for key := range keys {
		if key == keySessionRedis || strings.HasPrefix(key, keySessionRedis+".") {
			delete(keys, key)
		}
	}
}
