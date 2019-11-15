package suites

import (
	"fmt"
	"strings"
	"time"

	"github.com/clems4ever/authelia/utils"
	log "github.com/sirupsen/logrus"
)

func waitUntilServiceLogDetected(
	interval time.Duration,
	timeout time.Duration,
	dockerEnvironment *DockerEnvironment,
	service string,
	logPattern string) error {
	log.Debug("Waiting for service " + service + " to be ready...")
	err := utils.CheckUntil(5*time.Second, 1*time.Minute, func() (bool, error) {
		logs, err := dockerEnvironment.Logs(service, []string{"--tail", "20"})
		fmt.Printf(".")

		if err != nil {
			return false, err
		}
		return strings.Contains(logs, logPattern), nil
	})

	fmt.Print("\n")
	return err
}

func waitUntilAutheliaIsReady(dockerEnvironment *DockerEnvironment) error {
	log.Info("Waiting for Authelia to be ready...")

	err := waitUntilServiceLogDetected(
		5*time.Second,
		90*time.Second,
		dockerEnvironment,
		"authelia-backend",
		"Authelia is listening on")

	if err != nil {
		return err
	}

	err = waitUntilServiceLogDetected(
		5*time.Second,
		90*time.Second,
		dockerEnvironment,
		"authelia-frontend",
		"You can now view authelia-portal in the browser.")

	if err != nil {
		return err
	}
	log.Info("Authelia is now ready!")

	return nil
}
