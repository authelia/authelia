package authentication

// MustGetUserDetailsExtendedSafe returns the extended details for the given username, always returning a usable value
// even on error. An anonymous session has no username, so the backend is not consulted and empty details are returned.
func MustGetUserDetailsExtendedSafe(username string, provider UserProvider) (details UserDetailsExtended, err error) {
	details = UserDetailsExtended{UserDetails: &UserDetails{}}

	if username == "" {
		return details, nil
	}

	var d *UserDetailsExtended

	d, err = provider.GetDetailsExtended(username)

	if d == nil {
		return details, err
	}

	return *d, err
}

// MustGetUserDetailsExtendedCachedSafe is the cached equivalent of MustGetUserDetailsExtendedSafe. An anonymous
// session has no username, so the backend is not consulted and empty details are returned.
func MustGetUserDetailsExtendedCachedSafe(username string, provider UserProvider) (details UserDetailsExtended, err error) {
	details = UserDetailsExtended{UserDetails: &UserDetails{}}

	if username == "" {
		return details, nil
	}

	var d *UserDetailsExtended

	d, err = provider.GetDetailsExtendedCached(username)

	if d == nil {
		return details, err
	}

	return *d, err
}

// MustGetUserDetailsSafe returns the details for the given username, always returning a usable value even on error. An
// anonymous session has no username, so the backend is not consulted and empty details are returned.
func MustGetUserDetailsSafe(username string, provider UserProvider) (details UserDetails, err error) {
	if username == "" {
		return details, nil
	}

	var d *UserDetails

	d, err = provider.GetDetails(username)

	if d == nil {
		return details, err
	}

	return *d, err
}
