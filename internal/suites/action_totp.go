package suites

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/authelia/otp"
	"github.com/authelia/otp/totp"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "one-time-password-delete").Click("left", 1))

	rs.doMaybeVerifyIdentity(t, page)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "dialog-delete").Click("left", 1))

	rs.verifyNotificationDisplayed(t, page, "Successfully deleted the One-Time Password")

	rs.DeleteOneTimePassword(username)

	has, _, err := page.Has("#one-time-password-add")

	require.NoError(t, err)
	require.True(t, has)
}

func (rs *RodSession) doRegisterTOTPStart(t *testing.T, page *rod.Page, username string) {
	rs.doMaybeDeleteTOTP(t, page, username)

	elementAdd := rs.WaitElementLocatedByID(t, page, "one-time-password-add")

	require.NoError(t, elementAdd.Click("left", 1))

	rs.doMaybeVerifyIdentity(t, page)
}

func (rs *RodSession) doRegisterTOTPStartBadCode(t *testing.T, page *rod.Page, username string) {
	rs.doMaybeDeleteTOTP(t, page, username)

	elementAdd := rs.WaitElementLocatedByID(t, page, "one-time-password-add")

	require.NoError(t, elementAdd.Click("left", 1))

	if rs.isVerifyIdentityShowing(t, page) {
		rs.doMustVerifyIdentityBadCode(t, page)
		rs.doMustVerifyIdentity(t, page)
	}
}

func (rs *RodSession) doRegisterTOTPFinish(t *testing.T, page *rod.Page, username string, credential RodSuiteCredentialOneTimePassword) {
	passcode, err := credential.Generate(time.Now())
	require.NoError(t, err)

	rs.doEnterOTP(t, page, passcode)
	rs.verifyNotificationDisplayed(t, page, "Successfully added the One-Time Password")

	rs.SetOneTimePassword(username, credential)
}

func (rs *RodSession) doRegisterTOTPAdvanced(t *testing.T, page *rod.Page, invalid bool, username string, algorithm string, digits, period int) {
	if invalid {
		rs.doRegisterTOTPStartBadCode(t, page, username)
	} else {
		rs.doRegisterTOTPStart(t, page, username)
	}

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "one-time-password-advanced").Click("left", 1))
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "one-time-password-algorithm-"+algorithm).Click("left", 1))
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "one-time-password-length-"+strconv.Itoa(digits)).Click("left", 1))
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "one-time-password-period-"+strconv.Itoa(period)).Click("left", 1))
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "dialog-next").Click("left", 1))
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "qr-toggle").Click("left", 1))

	element := rs.WaitElementLocatedByID(t, page, "secret-url")

	raw, err := element.Text()
	require.NoError(t, err)

	secretURL, err := url.Parse(raw)
	require.NoError(t, err)

	values := secretURL.Query()

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
		Period:    uint(uperiod), //nolint:gosec // This is a test function.
		Skew:      1,
		Digits:    otp.Digits(udigits),
		Algorithm: alg,
	}

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "dialog-next").Click("left", 1))

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

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "dialog-next").Click("left", 1))
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "qr-toggle").Click("left", 1))

	secretURLElement := rs.WaitElementLocatedByID(t, page, "secret-url")

	secretURLRaw, err := secretURLElement.Text()
	require.NoError(t, err)

	secretURL, err := url.Parse(secretURLRaw)
	require.NoError(t, err)

	values := secretURL.Query()

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

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "dialog-next").Click("left", 1))

	rs.doRegisterTOTPFinish(t, page, username, credential)

	require.NoError(t, page.WaitStable(time.Millisecond*50))
	rs.doHoverAllMuiTooltip(t, page)
	require.NoError(t, page.WaitStable(time.Millisecond*50))

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
