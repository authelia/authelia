package suites

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/otp"
	"github.com/authelia/otp/totp"
)

type OptionsTOTP struct {
	Secret            string
	ValidationOptions totp.ValidateOpts
}

func (rs *RodSession) doMaybeDeleteTOTP(t *testing.T, page *rod.Page, username string) {
	ctx := page.GetContext()

	require.NoError(t, ctx.Err())

	require.NoError(t, page.WaitStable(time.Millisecond*50))

	has, _, err := page.Has("#one-time-password-delete")
	require.NoError(t, err)

	if !has {
		return
	}

	rs.doMustDeleteTOTP(t, page, username)
}

func (rs *RodSession) doMustDeleteTOTP(t *testing.T, page *rod.Page, username string) {
	rs.ClickElementLocatedByID(t, page, "one-time-password-delete")

	rs.doMaybeVerifyIdentity(t, page, "#dialog-delete")

	rs.ClickElementLocatedByID(t, page, "dialog-delete")

	rs.verifyNotificationDisplayed(t, page, "Successfully deleted the One-Time Password")

	rs.DeleteOneTimePassword(username)

	rs.WaitElementLocatedBySelector(t, page, "#one-time-password-add:not([disabled])")
}

func (rs *RodSession) doRegisterTOTPStart(t *testing.T, page *rod.Page, username string) {
	rs.doMaybeDeleteTOTP(t, page, username)

	rs.ClickElementLocatedByID(t, page, "one-time-password-add")

	rs.doMaybeVerifyIdentity(t, page, "#dialog-next")
}

func (rs *RodSession) doRegisterTOTPStartBadCode(t *testing.T, page *rod.Page, username string) {
	rs.doMaybeDeleteTOTP(t, page, username)

	rs.ClickElementLocatedByID(t, page, "one-time-password-add")

	if rs.isVerifyIdentityShowing(t, page, "#dialog-next") {
		rs.doMustVerifyIdentityBadCode(t, page)
		rs.doMustVerifyIdentity(t, page)
	}
}

func (rs *RodSession) doRegisterTOTPFinish(t *testing.T, page *rod.Page, username string, credential RodSuiteCredentialOneTimePassword) {
	passcode, err := credential.Generate(time.Now())
	require.NoError(t, err)

	rs.doEnterOTP(t, page, passcode)
	rs.verifyNotificationDisplayed(t, page, "Successfully added the One-Time Password")

	rs.WaitElementLocatedByID(t, page, "one-time-password-delete")

	rs.SetOneTimePassword(username, credential)
}

func (rs *RodSession) doWaitSecretURL(t *testing.T, page *rod.Page) *url.URL {
	element, err := page.ElementR("#secret-url", "otpauth://")
	require.NoError(t, err)

	raw, err := element.Text()
	require.NoError(t, err)

	secretURL, err := url.Parse(raw)
	require.NoError(t, err)

	return secretURL
}

func (rs *RodSession) doRegisterTOTPAdvanced(t *testing.T, page *rod.Page, invalid bool, username string, algorithm string, digits, period int) {
	if invalid {
		rs.doRegisterTOTPStartBadCode(t, page, username)
	} else {
		rs.doRegisterTOTPStart(t, page, username)
	}

	rs.ClickElementLocatedByID(t, page, "one-time-password-advanced")
	rs.ClickElementLocatedByID(t, page, "one-time-password-algorithm-"+algorithm)
	rs.ClickElementLocatedByID(t, page, "one-time-password-length-"+strconv.Itoa(digits))
	rs.ClickElementLocatedByID(t, page, "one-time-password-period-"+strconv.Itoa(period))
	rs.ClickElementLocatedByID(t, page, "dialog-next")
	rs.ClickElementLocatedByID(t, page, "qr-toggle")

	values := rs.doWaitSecretURL(t, page).Query()

	credential := RodSuiteCredentialOneTimePassword{
		Secret: values.Get("secret"),
	}

	ualgorithm := values.Get("algorithm")

	uperiod, err := strconv.Atoi(values.Get("period"))
	require.NoError(t, err)

	udigits, err := strconv.Atoi(values.Get("digits"))
	require.NoError(t, err)

	require.Equal(t, algorithm, ualgorithm)
	require.Equal(t, period, uperiod)
	require.Equal(t, digits, udigits)

	var alg otp.Algorithm

	switch strings.ToUpper(ualgorithm) {
	case SHA1:
		alg = otp.AlgorithmSHA1
	case SHA256:
		alg = otp.AlgorithmSHA256
	case SHA512:
		alg = otp.AlgorithmSHA512
	}

	credential.ValidationOptions = totp.ValidateOpts{
		Period:    uint(uperiod),
		Skew:      1,
		Digits:    otp.Digits(udigits),
		Algorithm: alg,
	}

	rs.ClickElementLocatedByID(t, page, "dialog-next")

	rs.doRegisterTOTPFinish(t, page, username, credential)
}

func (rs *RodSession) doOpenSettingsAndRegisterTOTP(t *testing.T, page *rod.Page, username string) {
	credential := rs.GetOneTimePassword(username)

	if credential.Valid() {
		return
	}

	rs.doOpenSettings(t, page)
	rs.doOpenSettingsMenuClickTwoFactor(t, page)
	rs.doRegisterTOTPStart(t, page, username)

	rs.ClickElementLocatedByID(t, page, "dialog-next")
	rs.ClickElementLocatedByID(t, page, "qr-toggle")

	values := rs.doWaitSecretURL(t, page).Query()

	credential.Secret = values.Get("secret")

	algorithm := otp.AlgorithmSHA1

	switch strings.ToUpper(values.Get("algorithm")) {
	case SHA1:
		algorithm = otp.AlgorithmSHA1
	case SHA256:
		algorithm = otp.AlgorithmSHA256
	case SHA512:
		algorithm = otp.AlgorithmSHA512
	}

	period, err := strconv.ParseUint(values.Get("period"), 10, 32)
	require.NoError(t, err)

	digits, err := strconv.ParseInt(values.Get("digits"), 10, 32)
	require.NoError(t, err)

	credential.ValidationOptions = totp.ValidateOpts{
		Period:    uint(period),
		Skew:      1,
		Digits:    otp.Digits(digits),
		Algorithm: algorithm,
	}

	rs.ClickElementLocatedByID(t, page, "dialog-next")

	rs.doRegisterTOTPFinish(t, page, username, credential)

	rs.doDismissTooltips(t, page)

	rs.doOpenSettingsMenuClickClose(t, page)
}

func (rs *RodSession) doEnterOTP(t *testing.T, page *rod.Page, passcode string) {
	inputs := rs.WaitElementsLocatedByID(t, page, "otp-input input")

	require.Greater(t, len(inputs), 0)

	for i := 0; i < len(passcode); i++ {
		err := inputs[i].Type(input.Key(passcode[i]))
		require.NoError(t, err)
	}
}

func (rs *RodSession) doValidateTOTP(t *testing.T, page *rod.Page, username string) {
	credential := rs.GetOneTimePassword(username)

	require.True(t, credential.Valid())

	passcode, err := credential.Generate(time.Now())
	assert.NoError(t, err)
	rs.doEnterOTP(t, page, passcode)
}
