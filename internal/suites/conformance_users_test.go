// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authentication"
)

func TestOIDCConformanceUserSuppliesEveryStandardClaim(t *testing.T) {
	database := authentication.NewFileUserDatabase("OIDCConformance/users.yml", false, false, nil)

	require.NoError(t, database.Load())

	user, err := database.GetUserDetails(testUsername)
	require.NoError(t, err)

	details := user.ToExtendedUserDetails()

	// scope=profile, less the claims Authelia derives rather than reads: preferred_username comes from the username
	// and updated_at is generated.
	assert.NotEmpty(t, details.GetDisplayName(), "backs the 'name' claim")
	assert.NotEmpty(t, details.GetGivenName(), "backs the 'given_name' claim")
	assert.NotEmpty(t, details.GetFamilyName(), "backs the 'family_name' claim")
	assert.NotEmpty(t, details.GetMiddleName(), "backs the 'middle_name' claim")
	assert.NotEmpty(t, details.GetNickname(), "backs the 'nickname' claim")
	assert.NotEmpty(t, details.GetProfile(), "backs the 'profile' claim")
	assert.NotEmpty(t, details.GetPicture(), "backs the 'picture' claim")
	assert.NotEmpty(t, details.GetWebsite(), "backs the 'website' claim")
	assert.NotEmpty(t, details.GetGender(), "backs the 'gender' claim")
	assert.NotEmpty(t, details.GetBirthdate(), "backs the 'birthdate' claim")
	assert.NotEmpty(t, details.GetZoneInfo(), "backs the 'zoneinfo' claim")
	assert.NotEmpty(t, details.GetLocale(), "backs the 'locale' claim")

	// scope=email, less email_verified which Authelia always reports.
	assert.NotEmpty(t, details.GetEmails(), "backs the 'email' claim")

	// scope=phone. Authelia emits phone_number_verified only when a number is present, so the number carries both
	// claims of this scope.
	assert.NotEmpty(t, details.GetPhoneNumberRFC3966(), "backs the 'phone_number' and 'phone_number_verified' claims")

	// scope=address, whose single claim is an object Authelia assembles from these five.
	assert.NotEmpty(t, details.GetStreetAddress(), "backs the 'address' claim")
	assert.NotEmpty(t, details.GetLocality(), "backs the 'address' claim")
	assert.NotEmpty(t, details.GetRegion(), "backs the 'address' claim")
	assert.NotEmpty(t, details.GetPostalCode(), "backs the 'address' claim")
	assert.NotEmpty(t, details.GetCountry(), "backs the 'address' claim")
}
