package suites

import (
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/utils"
)

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

func waitUntilAutheliaFrontendIsReady(dockerEnvironment *DockerEnvironment) error {
	return waitUntilServiceLogDetected(
		5*time.Second,
		180*time.Second,
		dockerEnvironment,
		"authelia-frontend",
		time.Time{},
		[]string{"dev server running at", "ready in", "server restarted"})
}

func waitUntilAutheliaFrontendRestarted(dockerEnvironment *DockerEnvironment) error {
	return waitUntilServiceLogDetected(
		5*time.Second,
		180*time.Second,
		dockerEnvironment,
		"authelia-frontend",
		time.Time{},
		[]string{"Watching for file changes"})
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
