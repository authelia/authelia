package suites

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/internal/utils"
)

func waitUntilServiceLogDetected(
	interval time.Duration,
	timeout time.Duration,
	dockerEnvironment *DockerEnvironment,
	service string,
	logPatterns []string) error {
	log.Debug("Waiting for service " + service + " to be ready...")
	err := utils.CheckUntil(5*time.Second, 1*time.Minute, func() (bool, error) {
		logs, err := dockerEnvironment.Logs(service, []string{"--tail", "20"})
		fmt.Printf(".")

		if err != nil {
			return false, err
		}
		for _, pattern := range logPatterns {
			if strings.Contains(logs, pattern) {
				return true, nil
			}
		}
		return false, nil
	})

	fmt.Print("\n")
	return err
}

func waitUntilAutheliaBackendIsReady(dockerEnvironment *DockerEnvironment) error {
	return waitUntilServiceLogDetected(
		5*time.Second,
		90*time.Second,
		dockerEnvironment,
		"authelia-backend",
		[]string{"Authelia is listening for"})
}

func waitUntilAutheliaFrontendIsReady(dockerEnvironment *DockerEnvironment) error {
	return waitUntilServiceLogDetected(
		5*time.Second,
		90*time.Second,
		dockerEnvironment,
		"authelia-frontend",
		[]string{"You can now view web in the browser.", "Compiled with warnings", "Compiled successfully!"})
}

func waitUntilAutheliaIsReady(dockerEnvironment *DockerEnvironment) error {
	log.Info("Waiting for Authelia to be ready...")

	if err := waitUntilAutheliaBackendIsReady(dockerEnvironment); err != nil {
		return err
	}

	if err := waitUntilAutheliaFrontendIsReady(dockerEnvironment); err != nil {
		return err
	}
	log.Info("Authelia is now ready!")
	return nil
}
