package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInfo_SetDefaultMethod_ShouldConfigureConfigDefault(t *testing.T) {
	none := "none"

	testName := func(i int, have UserInfo, availableMethods []string) string {
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
		if len(availableMethods) != 0 {
			available = strings.Join(availableMethods, " ")
		}

		return fmt.Sprintf("%d/method %s%s/available methods %s", i+1, method, has, available)
	}

	testCases := []struct {
		have             UserInfo
		availableMethods []string
		changed          bool
		want             UserInfo
	}{
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      true,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			availableMethods: []string{SecondFactorMethodWebauthn, SecondFactorMethodDuo},
			changed:          true,
			want: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      true,
				HasTOTP:     true,
				HasWebauthn: true,
			},
		},
		{
			have: UserInfo{
				HasDuo:      true,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			availableMethods: []string{SecondFactorMethodTOTP, SecondFactorMethodWebauthn, SecondFactorMethodDuo},
			changed:          true,
			want: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      true,
				HasTOTP:     true,
				HasWebauthn: true,
			},
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      true,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			availableMethods: []string{SecondFactorMethodTOTP},
			changed:          true,
			want: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      true,
				HasTOTP:     false,
				HasWebauthn: false,
			},
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			availableMethods: []string{SecondFactorMethodTOTP},
			changed:          true,
			want: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			availableMethods: []string{SecondFactorMethodWebauthn},
			changed:          true,
			want: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodTOTP,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
			availableMethods: []string{SecondFactorMethodDuo},
			changed:          true,
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      false,
				HasTOTP:     false,
				HasWebauthn: false,
			},
		},
		{
			have: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			availableMethods: []string{SecondFactorMethodTOTP, SecondFactorMethodWebauthn, SecondFactorMethodDuo},
			changed:          false,
			want: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			availableMethods: []string{SecondFactorMethodWebauthn, SecondFactorMethodDuo},
			changed:          true,
			want: UserInfo{
				Method:      SecondFactorMethodWebauthn,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			availableMethods: []string{SecondFactorMethodDuo},
			changed:          true,
			want: UserInfo{
				Method:      SecondFactorMethodDuo,
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
		},
		{
			have: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
			availableMethods: nil,
			changed:          false,
			want: UserInfo{
				Method:      "",
				HasDuo:      false,
				HasTOTP:     true,
				HasWebauthn: true,
			},
		},
	}

	for i, tc := range testCases {
		t.Run(testName(i, tc.have, tc.availableMethods), func(t *testing.T) {
			changed := tc.have.SetDefaultPreferred2FAMethod(tc.availableMethods)

			assert.Equal(t, tc.changed, changed)
			assert.Equal(t, tc.want, tc.have)
		})
	}
}
