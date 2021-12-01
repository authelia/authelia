package suites

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/utils"
)

var kubernetesSuiteName = "Kubernetes"

func init() {
	kind := Kind{}
	kubectl := Kubectl{}

	setup := func(suitePath string) error {
		cmd := utils.Shell("docker-compose -p authelia -f internal/suites/docker-compose.yml -f internal/suites/example/compose/kind/docker-compose.yml build")
		if err := cmd.Run(); err != nil {
			return err
		}

		exists, err := kind.ClusterExists()

		if err != nil {
			return err
		}

		if exists {
			log.Debug("Kubernetes cluster already exists")
		} else {
			err = kind.CreateCluster()

			if err != nil {
				return err
			}
		}

		log.Debug("Building authelia:dist image or use cache if already built...")

		if os.Getenv("CI") != t {
			if err := utils.Shell("authelia-scripts docker build").Run(); err != nil {
				return err
			}
		}

		log.Debug("Loading images into Kubernetes container...")

		if err := loadDockerImages(); err != nil {
			return err
		}

		log.Debug("Starting Kubernetes dashboard...")

		if err := kubectl.StartDashboard(); err != nil {
			return err
		}

		log.Debug("Deploying thirdparties...")

		if err := kubectl.DeployThirdparties(); err != nil {
			return err
		}

		log.Debug("Waiting for services to be ready...")

		if err := waitAllPodsAreReady(5 * time.Minute); err != nil {
			return err
		}

		log.Debug("Deploying Authelia...")

		if err = kubectl.DeployAuthelia(); err != nil {
			return err
		}

		log.Debug("Waiting for services to be ready...")

		if err := waitAllPodsAreReady(2 * time.Minute); err != nil {
			return err
		}

		log.Debug("Starting proxy...")

		err = kubectl.StartProxy()

		return err
	}

	teardown := func(suitePath string) error {
		err := kubectl.StopDashboard()
		if err != nil {
			log.Errorf("Unable to stop Kubernetes dashboard: %s", err)
		}

		err = kubectl.StopProxy()
		if err != nil {
			log.Errorf("Unable to stop Kind proxy: %s", err)
		}

		return kind.DeleteCluster()
	}

	GlobalRegistry.Register(kubernetesSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    12 * time.Minute,
		TestTimeout:     2 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
		Description:     "This suite has been created to test Authelia in a Kubernetes context and using nginx as the ingress controller.",
	})
}

func loadDockerImages() error {
	kind := Kind{}
	images := []string{"authelia:dist"}

	for _, image := range images {
		err := kind.LoadImage(image)

		if err != nil {
			return err
		}
	}

	return nil
}

func waitAllPodsAreReady(timeout time.Duration) error {
	kubectl := Kubectl{}
	// Wait in case the deployment has just been done and some services do not appear in kubectl logs.
	time.Sleep(1 * time.Second)
	fmt.Println("Check services are running")

	if err := kubectl.WaitPodsReady(timeout); err != nil {
		return err
	}

	fmt.Println("All pods are ready")

	return nil
}
