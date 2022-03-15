package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInfo_SetDefaultMethod_ShouldConfigureConfigDefault(t *testing.T) {
	var (
		info    UserInfo
		changed bool
	)

	info = UserInfo{
		Method:      SecondFactorMethodTOTP,
		HasDuo:      true,
		HasTOTP:     true,
		HasWebauthn: true,
	}

	changed = info.SetDefaultPreferred2FAMethod([]string{SecondFactorMethodWebauthn, SecondFactorMethodDuo})

	assert.True(t, changed)
	assert.Equal(t, SecondFactorMethodWebauthn, info.Method)

	info = UserInfo{
		Method:      "",
		HasDuo:      true,
		HasTOTP:     true,
		HasWebauthn: true,
	}

	changed = info.SetDefaultPreferred2FAMethod([]string{SecondFactorMethodTOTP, SecondFactorMethodWebauthn, SecondFactorMethodDuo})

	assert.True(t, changed)
	assert.Equal(t, SecondFactorMethodTOTP, info.Method)

	info = UserInfo{
		Method:      "webauthn",
		HasDuo:      true,
		HasTOTP:     false,
		HasWebauthn: false,
	}

	changed = info.SetDefaultPreferred2FAMethod([]string{SecondFactorMethodTOTP, SecondFactorMethodDuo})

	assert.True(t, changed)
	assert.Equal(t, SecondFactorMethodDuo, info.Method)

	info = UserInfo{
		Method:      "webauthn",
		HasDuo:      false,
		HasTOTP:     false,
		HasWebauthn: false,
	}

	changed = info.SetDefaultPreferred2FAMethod([]string{SecondFactorMethodTOTP})

	assert.True(t, changed)
	assert.Equal(t, SecondFactorMethodTOTP, info.Method)

	info = UserInfo{
		Method:      "totp",
		HasDuo:      false,
		HasTOTP:     false,
		HasWebauthn: false,
	}

	changed = info.SetDefaultPreferred2FAMethod([]string{SecondFactorMethodWebauthn})

	assert.True(t, changed)
	assert.Equal(t, SecondFactorMethodWebauthn, info.Method)

	info = UserInfo{
		Method:      "totp",
		HasDuo:      false,
		HasTOTP:     false,
		HasWebauthn: false,
	}

	changed = info.SetDefaultPreferred2FAMethod([]string{SecondFactorMethodDuo})

	assert.True(t, changed)
	assert.Equal(t, SecondFactorMethodDuo, info.Method)

	info = UserInfo{
		Method:      "",
		HasDuo:      false,
		HasTOTP:     false,
		HasWebauthn: false,
	}

	changed = info.SetDefaultPreferred2FAMethod([]string{SecondFactorMethodTOTP, SecondFactorMethodWebauthn, SecondFactorMethodDuo})

	assert.True(t, changed)
	assert.Equal(t, SecondFactorMethodTOTP, info.Method)

	info = UserInfo{
		Method:      "webauthn",
		HasDuo:      false,
		HasTOTP:     true,
		HasWebauthn: true,
	}

	changed = info.SetDefaultPreferred2FAMethod([]string{SecondFactorMethodTOTP})

	assert.True(t, changed)
	assert.Equal(t, SecondFactorMethodTOTP, info.Method)

	info.Method = ""

	changed = info.SetDefaultPreferred2FAMethod([]string{SecondFactorMethodWebauthn, SecondFactorMethodDuo})

	assert.True(t, changed)
	assert.Equal(t, SecondFactorMethodWebauthn, info.Method)

	info.Method = ""

	changed = info.SetDefaultPreferred2FAMethod([]string{SecondFactorMethodDuo})

	assert.True(t, changed)
	assert.Equal(t, SecondFactorMethodDuo, info.Method)

	info.Method = ""

	changed = info.SetDefaultPreferred2FAMethod(nil)

	assert.False(t, changed)
	assert.Equal(t, "", info.Method)
}
