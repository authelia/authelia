// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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

		if have.HasTOTP || have.HasDuo || have.HasWebauthn {
			has += " has"

			if have.HasTOTP {
				has += " " + SecondFactorMethodTOTP
			}

			if have.HasDuo {
				has += " " + SecondFactorMethodDuo
			}

			if have.HasWebauthn {
				has += " " + SecondFactorMethodWebauthn
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
				HasWebauthn: true,
			},
			want: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      true,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			methods: []string{SecondFactorMethodWebauthn, SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				HasDuo:      true,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			want: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      true,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			methods: []string{SecondFactorMethodTOTP, SecondFactorMethodWebauthn, SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      true,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      true,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			methods: []string{SecondFactorMethodTOTP},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			methods: []string{SecondFactorMethodTOTP},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			methods: []string{SecondFactorMethodWebauthn},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			methods: []string{SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			want: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			methods: []string{SecondFactorMethodTOTP, SecondFactorMethodWebauthn, SecondFactorMethodDuo},
			changed: false,
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			want: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			methods: []string{SecondFactorMethodWebauthn, SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			methods: []string{SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			want: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			methods: nil,
			changed: false,
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			methods:  []string{SecondFactorMethodTOTP, SecondFactorMethodWebauthn, SecondFactorMethodDuo},
			fallback: SecondFactorMethodDuo,
			changed:  true,
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			methods:  []string{SecondFactorMethodTOTP, SecondFactorMethodWebauthn},
			fallback: SecondFactorMethodDuo,
			changed:  true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      true,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      true,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			methods: []string{SecondFactorMethodWebauthn, SecondFactorMethodDuo},
			changed: true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			methods:  []string{SecondFactorMethodWebauthn, SecondFactorMethodDuo},
			fallback: SecondFactorMethodWebauthn,
			changed:  true,
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
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
