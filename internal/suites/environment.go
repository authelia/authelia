package suites

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/utils"
)

// suitesWithTraefikProxy matches suite names that run behind Traefik, which is what writes the access log.
var suitesWithTraefikProxy = regexp.MustCompile(`^(OIDCTraefik|PathPrefix|Traefik2|Traefik3)$`)

// suitesWithoutFrontend matches suite names that do not include the authelia-frontend service.
var suitesWithoutFrontend = regexp.MustCompile(`^CLI$`)

func waitUntilServiceLogDetected(
	interval time.Duration,
	timeout time.Duration,
	dockerEnvironment *DockerEnvironment,
	service string,
	since time.Time,
	logPatterns []string) error {
	log.Debug("Waiting for service " + service + " to be ready...")

	flags := []string{"--tail", "200"}

	if !since.IsZero() {
		// The log is cumulative, so without this an entry produced before the action being waited on
		// satisfies the wait immediately. Rewound slightly to keep an entry from the same instant.
		flags = []string{"--since", since.Add(-time.Second).Format(time.RFC3339Nano)}
	}

	err := utils.CheckUntil(interval, timeout, func() (bool, error) {
		logs, err := dockerEnvironment.Logs(service, flags)
		if err != nil {
			return false, err
		}

		for _, pattern := range logPatterns {
			if strings.Contains(logs, pattern) {
				log.Debug("Service " + service + " is ready!")
				return true, nil
			}
		}

		return false, nil
	})

	return err
}

func waitUntilAutheliaBackendIsReady(dockerEnvironment *DockerEnvironment, since time.Time) error {
	return waitUntilServiceLogDetected(
		5*time.Second,
		180*time.Second,
		dockerEnvironment,
		"authelia-backend",
		since,
		[]string{"Startup complete"})
}

func waitUntilAutheliaUpstreamBackendIsReady(dockerEnvironment *DockerEnvironment) error {
	if os.Getenv("CI") == t {
		return nil
	}

	log.Info("Waiting for Authelia (Upstream Provider) to be ready...")

	if err := waitUntilServiceLogDetected(
		5*time.Second,
		180*time.Second,
		dockerEnvironment,
		"authelia-upstream-backend",
		time.Time{},
		[]string{"Startup complete"}); err != nil {
		return err
	}

	log.Info("Authelia (Upstream Provider) is ready!")

	return nil
}

func waitUntilAutheliaFrontendIsReady(dockerEnvironment *DockerEnvironment) error {
	return waitUntilServiceLogDetected(
		5*time.Second,
		180*time.Second,
		dockerEnvironment,
		"authelia-frontend",
		time.Time{},
		[]string{"ready in"})
}

func waitUntilAutheliaFrontendRestarted(dockerEnvironment *DockerEnvironment, since time.Time) error {
	return waitUntilServiceLogDetected(
		5*time.Second,
		180*time.Second,
		dockerEnvironment,
		"authelia-frontend",
		since,
		[]string{"server restarted"})
}

func waitUntilK3DIsReady(dockerEnvironment *DockerEnvironment) error {
	return waitUntilServiceLogDetected(
		5*time.Second,
		180*time.Second,
		dockerEnvironment,
		"k3d",
		time.Time{},
		[]string{"API listen on [::]:2376"})
}

func waitUntilSambaIsReady(dockerEnvironment *DockerEnvironment) error {
	return waitUntilServiceLogDetected(
		5*time.Second,
		180*time.Second,
		dockerEnvironment,
		"sambaldap",
		time.Time{},
		[]string{"samba entered RUNNING state"})
}

func waitUntilServiceLog(dockerEnvironment *DockerEnvironment, service, log string, since time.Time) error {
	return waitUntilServiceLogDetected(
		time.Second,
		30*time.Second,
		dockerEnvironment,
		service,
		since,
		[]string{log})
}

func waitUntilAutheliaIsReady(dockerEnvironment *DockerEnvironment, suite string) error {
	if os.Getenv("CI") == t {
		return nil
	}

	if !suitesWithoutFrontend.MatchString(suite) {
		log.Info("Waiting for Authelia (Frontend) to be ready...")

		if err := waitUntilAutheliaFrontendIsReady(dockerEnvironment); err != nil {
			return err
		}

		log.Info("Authelia (Frontend) is ready!")
	}

	log.Info("Waiting for Authelia (Backend) to be ready...")

	if err := waitUntilAutheliaBackendIsReady(dockerEnvironment, time.Time{}); err != nil {
		return err
	}

	log.Info("Authelia (Backend) is ready!")

	if suite == "ActiveDirectory" {
		log.Info("Waiting for Samba to be ready...")

		if err := waitUntilSambaIsReady(dockerEnvironment); err != nil {
			return err
		}

		log.Info("Samba is ready!")
	}

	return nil
}

func waitUntilProxyRoutesPortal(baseDomain string) error {
	// The proxy discovers the portal from the daemon, and `up --wait` returns once the portal's healthcheck
	// passes, which is before the proxy has enumerated the daemon and published a route to it. Until that
	// route exists the proxy answers the portal's host itself, and no wait recovers a test that visits the
	// error page it serves, because the document loads and the portal is never fetched. The path is
	// deliberately unprefixed: it is the one route the portal serves outside its own base URL, so it answers
	// for a suite with a path prefix as well as one without.
	log.Info("Waiting for the proxy to route the portal...")

	client, target := NewHTTPClient(), LoginBaseURLFmt(baseDomain)+"/api/health"

	if err := utils.CheckUntil(time.Second, 30*time.Second, func() (bool, error) {
		response, err := client.Get(target)
		if err != nil {
			return false, nil
		}

		defer response.Body.Close()

		_, _ = io.Copy(io.Discard, response.Body)

		return response.StatusCode == http.StatusOK, nil
	}); err != nil {
		return fmt.Errorf("the proxy did not route '%s' to the portal: %w", target, err)
	}

	log.Info("The proxy routes the portal!")

	return nil
}
