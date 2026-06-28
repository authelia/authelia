package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInjectCIFlag(t *testing.T) {
	testCases := []struct {
		Name string
		have string
		want string
	}{
		{
			Name: "ShouldPrependFlagWhenNoFlags",
			have: "^(api|admin)$",
			want: "(?i)^(api|admin)$",
		},
		{
			Name: "ShouldPrependFlagWhenEmpty",
			have: "",
			want: "(?i)",
		},
		{
			Name: "ShouldNotPrependFlagWhenAlreadyInsensitive",
			have: "(?i)^(api|admin)$",
			want: "(?i)^(api|admin)$",
		},
		{
			Name: "ShouldNotPrependFlagWhenExplicitlySensitive",
			have: "(?-i)^(api|admin)$",
			want: "(?-i)^(api|admin)$",
		},
		{
			Name: "ShouldNotPrependFlagWhenInsensitiveAmongOtherFlags",
			have: "(?is)foo",
			want: "(?is)foo",
		},
		{
			Name: "ShouldNotPrependFlagWhenInsensitiveAfterOtherFlags",
			have: "(?si)foo",
			want: "(?si)foo",
		},
		{
			Name: "ShouldNotPrependFlagWhenInsensitiveAmongManyFlags",
			have: "(?ims)foo",
			want: "(?ims)foo",
		},
		{
			Name: "ShouldNotPrependFlagWhenSensitiveAmongOtherFlags",
			have: "(?s-i)foo",
			want: "(?s-i)foo",
		},
		{
			Name: "ShouldPrependFlagWhenOtherFlagWithoutInsensitive",
			have: "(?s)foo",
			want: "(?is)foo",
		},
		{
			Name: "ShouldPrependFlagWhenMultilineFlagWithoutInsensitive",
			have: "(?m)foo",
			want: "(?im)foo",
		},
		{
			Name: "ShouldPrependFlagWhenNegatedFlagWithoutInsensitive",
			have: "(?-s)foo",
			want: "(?i-s)foo",
		},
		{
			Name: "ShouldPrependFlagWhenMixedFlagsWithoutInsensitive",
			have: "(?ms-U)foo",
			want: "(?ims-U)foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.want, injectCIFlag(tc.have))
		})
	}
}
