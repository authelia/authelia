// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAMRMappingEmpty(t *testing.T) {
	mapping, err := NewAuthenticationMethodsReferencesMapping("")

	assert.NoError(t, err)
	assert.Nil(t, mapping)

	amr, err := mapping.Evaluate(map[string]any{"amr": []any{"mfa"}})

	assert.NoError(t, err)
	assert.Nil(t, amr)
}

func TestAMRMappingMFA(t *testing.T) {
	mapping, err := NewAuthenticationMethodsReferencesMapping(`has(claims.amr) && 'mfa' in claims.amr ? ['pwd','otp'] : ['pwd']`)
	require.NoError(t, err)
	require.NotNil(t, mapping)

	amr, err := mapping.Evaluate(map[string]any{"amr": []any{"mfa"}})
	assert.NoError(t, err)
	assert.Equal(t, []string{"pwd", "otp"}, amr)

	amr, err = mapping.Evaluate(map[string]any{"amr": []any{"pwd"}})
	assert.NoError(t, err)
	assert.Equal(t, []string{"pwd"}, amr)

	amr, err = mapping.Evaluate(map[string]any{})
	assert.NoError(t, err)
	assert.Equal(t, []string{"pwd"}, amr)
}

func TestAMRMappingMissingClaimErrors(t *testing.T) {
	mapping, err := NewAuthenticationMethodsReferencesMapping(`'mfa' in claims.amr ? ['pwd','otp'] : ['pwd']`)
	require.NoError(t, err)
	require.NotNil(t, mapping)

	amr, err := mapping.Evaluate(map[string]any{})
	assert.Error(t, err)
	assert.Nil(t, amr)
}

func TestAMRMappingWrongOutputType(t *testing.T) {
	mapping, err := NewAuthenticationMethodsReferencesMapping(`'two_factor'`)

	assert.Nil(t, mapping)
	assert.Error(t, err)
}

func TestAMRMappingCompileError(t *testing.T) {
	mapping, err := NewAuthenticationMethodsReferencesMapping(`[`)

	assert.Nil(t, mapping)
	assert.Error(t, err)

	mapping, err = NewAuthenticationMethodsReferencesMapping(`other.amr`)

	assert.Nil(t, mapping)
	assert.Error(t, err)
}

func TestAMRMappingDynamicClaimsCannotProduceNonStrings(t *testing.T) {
	mapping, err := NewAuthenticationMethodsReferencesMapping(`claims.amr`)

	assert.Nil(t, mapping)
	assert.Error(t, err)

	mapping, err = NewAuthenticationMethodsReferencesMapping(`claims.amr.map(v, string(v))`)
	require.NoError(t, err)

	amr, err := mapping.Evaluate(map[string]any{"amr": []any{"pwd", 1, true}})

	assert.NoError(t, err)
	assert.Equal(t, []string{"pwd", "1", "true"}, amr)

	amr, err = mapping.Evaluate(map[string]any{"amr": "pwd"})

	assert.Error(t, err)
	assert.Nil(t, amr)
}
