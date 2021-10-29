package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestNetworkACLSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Spec("TestShouldAccessSecretUpon2FA", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		s := setupTest(t, "", false)

		defer func() {
			cancel()
			s.collectScreenshot(ctx.Err(), s.Page)
			teardownTest(s)
		}()

		targetURL := fmt.Sprintf("%s/secret.html", SecureBaseURL)
		s.doVisit(s.Context(ctx), targetURL)

		s.verifyIsFirstFactorPage(t, s.Context(ctx))
		s.doRegisterAndLogin2FA(t, s.Context(ctx), testUsername, testPassword, false, targetURL)
		s.verifySecretAuthorized(t, s.Context(ctx))
	})

	// from network 192.168.240.201/32.
	o.Spec("TestShouldAccessSecretUpon1FA", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		s := setupTest(t, "http://proxy-client1.example.com:3128", false)

		defer func() {
			cancel()
			s.collectScreenshot(ctx.Err(), s.Page)
			teardownTest(s)
		}()

		targetURL := fmt.Sprintf("%s/secret.html", SecureBaseURL)
		s.Page = s.doCreateTab(targetURL)

		s.verifyIsFirstFactorPage(t, s.Context(ctx))
		s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword,
			false, targetURL)
		s.verifySecretAuthorized(t, s.Context(ctx))
	})

	// from network 192.168.240.202/32.
	o.Spec("TestShouldAccessSecretUpon0FA", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		s := setupTest(t, "http://proxy-client2.example.com:3128", false)

		defer func() {
			cancel()
			s.collectScreenshot(ctx.Err(), s.Page)
			teardownTest(s)
		}()

		s.Page = s.doCreateTab(fmt.Sprintf("%s/secret.html", SecureBaseURL))
		s.verifySecretAuthorized(t, s.Context(ctx))
	})
}
