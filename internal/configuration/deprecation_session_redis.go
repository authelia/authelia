package configuration

import (
	"fmt"
	"strconv"
	"strings"

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
	portDefault := uint16(schema.DefaultRedisCachePort)
	if sentinel {
		portDefault = schema.DefaultRedisSentinelCachePort
	}

	host, port, err := getHostPort(keySessionRedisHost, keySessionRedisPort, "", portDefault, keys)
	if err != nil {
		return err
	}

	if !sentinel {
		if host == "" {
			return nil
		}

		address, err := schema.NewAddressFromNetworkValuesDefault(host, port, schema.AddressSchemeTCP, schema.AddressSchemeUnix)
		if err != nil {
			return err
		}

		keys[prefix+".address"] = address.String()

		return nil
	}

	var addresses []any

	if host != "" {
		addresses = append(addresses, sessionRedisAddress(host, port))
	}

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

		nodePort := uint16(schema.DefaultRedisSentinelCachePort)

		if raw, ok := m["port"]; ok {
			if parsed, perr := sessionRedisPort(raw); perr != nil {
				return perr
			} else if parsed != 0 {
				nodePort = parsed
			}
		}

		address := sessionRedisAddress(nodeHost, nodePort)

		if !anySliceContains(addresses, address) {
			addresses = append(addresses, address)
		}
	}

	if len(addresses) != 0 {
		keys[prefix+".addresses"] = addresses
	}

	return nil
}

func sessionRedisAddress(host string, port uint16) string {
	address := schema.NewAddressFromNetworkValues(schema.AddressSchemeTCP, strings.ToLower(host), port)

	return address.String()
}

func sessionRedisPort(value any) (port uint16, err error) {
	switch v := value.(type) {
	case uint16:
		return v, nil
	case int:
		if v < 0 || v > 65535 {
			return 0, fmt.Errorf("port %d is not a valid port number", v)
		}

		return uint16(v), nil
	case string:
		if v == "" {
			return 0, nil
		}

		p, perr := strconv.ParseUint(v, 10, 16)
		if perr != nil {
			return 0, fmt.Errorf("error occurred converting the port from a string: %w", perr)
		}

		return uint16(p), nil
	default:
		return 0, nil
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

// mapHasKeyPrefix reports whether any key in the map is the given key or is nested beneath it.
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
