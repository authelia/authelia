package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInfo_SetDefaultMethod_ShouldConfigureConfigDefault(t *testing.T) {
	var info UserInfo

	info = UserInfo{
		Method:      SecondFactorMethodTOTP,
		HasDuo:      true,
		HasTOTP:     true,
		HasWebauthn: true,
	}

	info.SetDefaultMethod([]string{SecondFactorMethodWebauthn, SecondFactorMethodDuo})

	assert.Equal(t, SecondFactorMethodWebauthn, info.Method)

	info = UserInfo{
		Method:      "",
		HasDuo:      true,
		HasTOTP:     true,
		HasWebauthn: true,
	}

	info.SetDefaultMethod([]string{SecondFactorMethodTOTP, SecondFactorMethodWebauthn, SecondFactorMethodDuo})

	assert.Equal(t, SecondFactorMethodTOTP, info.Method)

	info = UserInfo{
		Method:      "",
		HasDuo:      true,
		HasTOTP:     false,
		HasWebauthn: false,
	}

	info.SetDefaultMethod([]string{SecondFactorMethodTOTP, SecondFactorMethodWebauthn, SecondFactorMethodDuo})

	assert.Equal(t, SecondFactorMethodDuo, info.Method)

	info = UserInfo{
		Method:      "",
		HasDuo:      false,
		HasTOTP:     false,
		HasWebauthn: false,
	}

	info.SetDefaultMethod([]string{SecondFactorMethodTOTP, SecondFactorMethodWebauthn, SecondFactorMethodDuo})

	assert.Equal(t, SecondFactorMethodTOTP, info.Method)

	info.Method = ""

	info.SetDefaultMethod([]string{SecondFactorMethodWebauthn, SecondFactorMethodDuo})

	assert.Equal(t, SecondFactorMethodWebauthn, info.Method)

	info.Method = ""

	info.SetDefaultMethod([]string{SecondFactorMethodDuo})

	assert.Equal(t, SecondFactorMethodDuo, info.Method)

	info.Method = ""

	info.SetDefaultMethod(nil)

	assert.Equal(t, SecondFactorMethodTOTP, info.Method)
}
