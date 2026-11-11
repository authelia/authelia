// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authentication

// MustGetUserDetailsExtendedSafe returns the extended details for the given username, always returning a usable value
// even on error. An anonymous session has no username, so the backend is not consulted and empty details are returned.
func MustGetUserDetailsExtendedSafe(username string, provider UserProvider) (details UserDetailsExtended, err error) {
	return mustGetSafe(username, provider.GetDetailsExtended, UserDetailsExtended{UserDetails: &UserDetails{}})
}

// MustGetUserDetailsExtendedCachedSafe is the cached equivalent of MustGetUserDetailsExtendedSafe. An anonymous
// session has no username, so the backend is not consulted and empty details are returned.
func MustGetUserDetailsExtendedCachedSafe(username string, provider UserProvider) (details UserDetailsExtended, err error) {
	return mustGetSafe(username, provider.GetDetailsExtendedCached, UserDetailsExtended{UserDetails: &UserDetails{}})
}

// MustGetUserDetailsSafe returns the details for the given username, always returning a usable value even on error. An
// anonymous session has no username, so the backend is not consulted and empty details are returned.
func MustGetUserDetailsSafe(username string, provider UserProvider) (details UserDetails, err error) {
	return mustGetSafe(username, provider.GetDetails, UserDetails{})
}

// MustGetUserDetailsCachedSafe is the cached equivalent of MustGetUserDetailsSafe. An anonymous session has no username,
// so the backend is not consulted and empty details are returned.
func MustGetUserDetailsCachedSafe(username string, provider UserProvider) (details UserDetails, err error) {
	return mustGetSafe(username, provider.GetDetailsCached, UserDetails{})
}

func mustGetSafe[T any](username string, get func(username string) (*T, error), fallback T) (value T, err error) {
	if username == "" {
		return fallback, nil
	}

	var result *T

	if result, err = get(username); result == nil {
		return fallback, err
	}

	return *result, err
}
