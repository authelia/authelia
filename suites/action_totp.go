package suites

import (
	"context"
	"time"

	"github.com/pquerna/otp/totp"
)

func doRegisterTOTP(ctx context.Context, s *SeleniumSuite) string {
	WaitElementLocatedByClassName(ctx, s, "register-totp").Click()
	verifyBodyContains(ctx, s, "Please check your e-mails")
	link := doGetLinkFromLastMail(s)
	doVisit(s, link)
	secret, err := WaitElementLocatedByClassName(ctx, s, "base32-secret").Text()
	s.Assert().NoError(err)
	return secret
}

func doValidateTOTP(ctx context.Context, s *SeleniumSuite, secret string) {
	code, err := totp.GenerateCode(secret, time.Now())
	s.Assert().NoError(err)
	WaitElementLocatedByID(ctx, s, "totp-token").SendKeys(code)
	WaitElementLocatedByID(ctx, s, "totp-button").Click()
}
