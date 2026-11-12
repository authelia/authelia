// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryTransformMSSQLBitValues(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldReplaceBooleanLiterals", "UPDATE x SET revoked = TRUE WHERE id = @p1 AND active = FALSE;", "UPDATE x SET revoked = 1 WHERE id = @p1 AND active = 0;"},
		{"ShouldReplaceCaseBooleanLiterals", "clone_warning = CASE clone_warning WHEN TRUE THEN TRUE ELSE @p1 END", "clone_warning = CASE clone_warning WHEN 1 THEN 1 ELSE @p1 END"},
		{"ShouldNotReplacePartialWords", "SELECT TRUEISH, FALSEHOOD FROM x;", "SELECT TRUEISH, FALSEHOOD FROM x;"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, queryTransformMSSQLBitValues(tc.have))
		})
	}
}

func TestQueryTransformMSSQLLength(t *testing.T) {
	assert.Equal(t, "WHERE issuer = @p1 AND DATALENGTH(data) > 0;", queryTransformMSSQLLength("WHERE issuer = @p1 AND LENGTH(data) > 0;"))
	assert.Equal(t, "WHERE DATALENGTH(data) > 0;", queryTransformMSSQLLength("WHERE DATALENGTH(data) > 0;"))
}
