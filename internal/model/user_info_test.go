package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInfo_SetDefaultMethod(t *testing.T) {
	none := "none"

	testName := func(i int, have UserInfo, methods []string, fallback string) string {
		method := have.Method

		if method == "" {
			method = none
		}

		has := ""

		if have.HasTOTP || have.HasDuo || have.HasWebAuthn {
			has += " has"

			if have.HasTOTP {
				has += " " + SecondFactorMethodTOTP
			}

			if have.HasDuo {
				has += " " + SecondFactorMethodDuo
			}

			if have.HasWebAuthn {
				has += " " + SecondFactorMethodWebAuthn
			}
		}

		available := none
		if len(methods) != 0 {
			available = strings.Join(methods, " ")
		}

		if fallback != "" {
			fallback = "/fallback " + fallback
		}

		return fmt.Sprintf("%d/method %s%s/available methods %s%s", i+1, method, has, available, fallback)
	}

	testCases := []struct {
		have UserInfo
		want UserInfo

		methods  []string
		fallback string

		changed bool
	}{
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      true,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			want: UserInfo{
				Method:      SecondFactorMethodWebAuthn,
				HasDuo:      true,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			methods: []string{SecondFactorMethodWebAuthn, SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				HasDuo:      true,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			want: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      true,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			methods: []string{SecondFactorMethodTOTP, SecondFactorMethodWebAuthn, SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodWebAuthn,
				HasDuo:      true,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      true,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			methods: []string{SecondFactorMethodTOTP},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodWebAuthn,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			methods: []string{SecondFactorMethodTOTP},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodWebAuthn,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			methods: []string{SecondFactorMethodWebAuthn},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			methods: []string{SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodWebAuthn,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			want: UserInfo{
				Method:      SecondFactorMethodWebAuthn,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			methods: []string{SecondFactorMethodTOTP, SecondFactorMethodWebAuthn, SecondFactorMethodDuo},
			changed: false,
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			want: UserInfo{
				Method:      SecondFactorMethodWebAuthn,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			methods: []string{SecondFactorMethodWebAuthn, SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			methods: []string{SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			want: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebAuthn: true,
			},
			methods: nil,
			changed: false,
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			methods:  []string{SecondFactorMethodTOTP, SecondFactorMethodWebAuthn, SecondFactorMethodDuo},
			fallback: SecondFactorMethodDuo,
			changed:  true,
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			methods:  []string{SecondFactorMethodTOTP, SecondFactorMethodWebAuthn},
			fallback: SecondFactorMethodDuo,
			changed:  true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      true,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      true,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			methods: []string{SecondFactorMethodWebAuthn, SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodWebAuthn,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			methods:  []string{SecondFactorMethodWebAuthn, SecondFactorMethodDuo},
			fallback: SecondFactorMethodWebAuthn,
			changed:  true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodWebAuthn,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebAuthn: false,
			},
			methods:  []string{SecondFactorMethodTOTP, SecondFactorMethodDuo},
			fallback: SecondFactorMethodDuo,
			changed:  true,
		},
	}

	for i, tc := range testCases {
		t.Run(testName(i, tc.have, tc.methods, tc.fallback), func(t *testing.T) {
			changed := tc.have.SetDefaultPreferred2FAMethod(tc.methods, tc.fallback)

			assert.Equal(t, tc.changed, changed)
			assert.Equal(t, tc.want, tc.have)
		})
	}
}
