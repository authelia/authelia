package middlewares

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustelem/zxcvbn"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewPasswordPolicyProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		have     schema.PasswordPolicy
		expected PasswordPolicyProvider
	}{
		{
			desc:     "ShouldReturnUnconfiguredProvider",
			have:     schema.PasswordPolicy{},
			expected: &StandardPasswordPolicyProvider{},
		},
		{
			desc:     "ShouldReturnProviderWhenZxcvbn",
			have:     schema.PasswordPolicy{ZXCVBN: schema.PasswordPolicyZXCVBN{Enabled: true, MinScore: 10}},
			expected: &ZXCVBNPasswordPolicyProvider{minScore: 10},
		},
		{
			desc:     "ShouldReturnConfiguredProviderWithMin",
			have:     schema.PasswordPolicy{Standard: schema.PasswordPolicyStandard{Enabled: true, MinLength: 8}},
			expected: &StandardPasswordPolicyProvider{min: 8},
		},
		{
			desc:     "ShouldReturnConfiguredProviderWitHMinMax",
			have:     schema.PasswordPolicy{Standard: schema.PasswordPolicyStandard{Enabled: true, MinLength: 8, MaxLength: 100}},
			expected: &StandardPasswordPolicyProvider{min: 8, max: 100},
		},
		{
			desc:     "ShouldReturnConfiguredProviderWithMinLowercase",
			have:     schema.PasswordPolicy{Standard: schema.PasswordPolicyStandard{Enabled: true, MinLength: 8, RequireLowercase: true}},
			expected: &StandardPasswordPolicyProvider{min: 8, patterns: []regexp.Regexp{*regexp.MustCompile(`[a-z]+`)}},
		},
		{
			desc:     "ShouldReturnConfiguredProviderWithMinLowercaseUppercase",
			have:     schema.PasswordPolicy{Standard: schema.PasswordPolicyStandard{Enabled: true, MinLength: 8, RequireLowercase: true, RequireUppercase: true}},
			expected: &StandardPasswordPolicyProvider{min: 8, patterns: []regexp.Regexp{*regexp.MustCompile(`[a-z]+`), *regexp.MustCompile(`[A-Z]+`)}},
		},
		{
			desc:     "ShouldReturnConfiguredProviderWithMinLowercaseUppercaseNumber",
			have:     schema.PasswordPolicy{Standard: schema.PasswordPolicyStandard{Enabled: true, MinLength: 8, RequireLowercase: true, RequireUppercase: true, RequireNumber: true}},
			expected: &StandardPasswordPolicyProvider{min: 8, patterns: []regexp.Regexp{*regexp.MustCompile(`[a-z]+`), *regexp.MustCompile(`[A-Z]+`), *regexp.MustCompile(`[0-9]+`)}},
		},
		{
			desc:     "ShouldReturnConfiguredProviderWithMinLowercaseUppercaseSpecial",
			have:     schema.PasswordPolicy{Standard: schema.PasswordPolicyStandard{Enabled: true, MinLength: 8, RequireLowercase: true, RequireUppercase: true, RequireSpecial: true}},
			expected: &StandardPasswordPolicyProvider{min: 8, patterns: []regexp.Regexp{*regexp.MustCompile(`[a-z]+`), *regexp.MustCompile(`[A-Z]+`), *regexp.MustCompile(`[^a-zA-Z0-9]+`)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := NewPasswordPolicyProvider(tc.have)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestPasswordPolicyProvider_Validate(t *testing.T) {
	testCases := []struct {
		desc     string
		config   schema.PasswordPolicy
		have     []string
		expected []error
	}{
		{
			desc:     "ShouldValidateAllPasswords",
			config:   schema.PasswordPolicy{},
			have:     []string{"a", "1", "a really str0ng pass12nm3kjl12word@@#4"},
			expected: []error{nil, nil, nil},
		},
		{
			desc:     "ShouldValidatePasswordMinLength",
			config:   schema.PasswordPolicy{Standard: schema.PasswordPolicyStandard{Enabled: true, MinLength: 8}},
			have:     []string{"a", "b123", "1111111", "aaaaaaaa", "1o23nm1kio2n3k12jn"},
			expected: []error{errPasswordPolicyNoMet, errPasswordPolicyNoMet, errPasswordPolicyNoMet, nil, nil},
		},
		{
			desc:   "ShouldValidatePasswordMaxLength",
			config: schema.PasswordPolicy{Standard: schema.PasswordPolicyStandard{Enabled: true, MaxLength: 30}},
			have: []string{
				"a1234567894654wkjnkjasnskjandkjansdkjnas",
				"012345678901234567890123456789a",
				"0123456789012345678901234567890123456789",
				"012345678901234567890123456789",
				"1o23nm1kio2n3k12jn",
			},
			expected: []error{errPasswordPolicyNoMet, errPasswordPolicyNoMet, errPasswordPolicyNoMet, nil, nil},
		},
		{
			desc:     "ShouldValidatePasswordAdvancedLowerUpperMin8",
			config:   schema.PasswordPolicy{Standard: schema.PasswordPolicyStandard{Enabled: true, MinLength: 8, RequireLowercase: true, RequireUppercase: true}},
			have:     []string{"a", "b123", "1111111", "aaaaaaaa", "1o23nm1kio2n3k12jn", "ANJKJQ@#NEK!@#NJK!@#", "qjik2nkjAkjlmn123"},
			expected: []error{errPasswordPolicyNoMet, errPasswordPolicyNoMet, errPasswordPolicyNoMet, errPasswordPolicyNoMet, errPasswordPolicyNoMet, errPasswordPolicyNoMet, nil},
		},
		{
			desc:   "ShouldValidatePasswordAdvancedAllMax100Min8",
			config: schema.PasswordPolicy{Standard: schema.PasswordPolicyStandard{Enabled: true, MinLength: 8, MaxLength: 100, RequireLowercase: true, RequireUppercase: true, RequireNumber: true, RequireSpecial: true}},
			have: []string{
				"a",
				"b123",
				"1111111",
				"aaaaaaaa",
				"1o23nm1kio2n3k12jn",
				"ANJKJQ@#NEK!@#NJK!@#",
				"qjik2nkjAkjlmn123",
				"qjik2n@jAkjlmn123",
				"qjik2n@jAkjlmn123qjik2n@jAkjlmn123qjik2n@jAkjlmn123qjik2n@jAkjlmn123qjik2n@jAkjlmn123qjik2n@jAkjlmn123",
			},
			expected: []error{
				errPasswordPolicyNoMet,
				errPasswordPolicyNoMet,
				errPasswordPolicyNoMet,
				errPasswordPolicyNoMet,
				errPasswordPolicyNoMet,
				errPasswordPolicyNoMet,
				errPasswordPolicyNoMet,
				nil,
				errPasswordPolicyNoMet,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, len(tc.have), len(tc.expected))

			for i := 0; i < len(tc.have); i++ {
				provider := NewPasswordPolicyProvider(tc.config)
				t.Run(tc.have[i], func(t *testing.T) {
					assert.Equal(t, tc.expected[i], provider.Check(tc.have[i]))
				})
			}
		})
	}
}

func TestZXCVBNPasswordPolicyProvider_Check(t *testing.T) {
	weak := "a"
	strong := "A1!a2@B3#c4$D5%E6^f7&G8*h9"

	sWeak := zxcvbn.PasswordStrength(weak, nil).Score
	sStrong := zxcvbn.PasswordStrength(strong, nil).Score

	testCases := []struct {
		name     string
		password string
		minScore int
		wantErr  bool
	}{
		{name: "ShouldAllowWhenScoreMeetsMinimumWeak", password: weak, minScore: sWeak, wantErr: false},
		{name: "ShouldRejectWhenScoreBelowMinimumWeak", password: weak, minScore: sWeak + 1, wantErr: true},
		{name: "ShouldAllowWhenScoreMeetsMinimumStrong", password: strong, minScore: sStrong, wantErr: false},
		{name: "ShouldRejectWhenScoreBelowMinimumStrong", password: strong, minScore: sStrong + 1, wantErr: true},
		{name: "ShouldAllowWhenMinimumIsZero", password: weak, minScore: 0, wantErr: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := ZXCVBNPasswordPolicyProvider{minScore: tc.minScore}

			err := p.Check(tc.password)

			if tc.wantErr {
				require.ErrorIs(t, err, errPasswordPolicyNoMet)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
