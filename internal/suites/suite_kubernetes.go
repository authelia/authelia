package suites

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/utils"
)

var kubernetesSuiteName = "Kubernetes"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/example/compose/k3d/docker-compose.yml",
	})

	k3d := K3D{}
	kubectl := Kubectl{}

	setup := func(suitePath string) error {
		err := dockerEnvironment.Up()
		if err != nil {
			return err
		}

		err = waitUntilK3DIsReady(dockerEnvironment)
		if err != nil {
			return err
		}

		exists, err := k3d.ClusterExists()
		if err != nil {
			return err
		}

		if exists {
			log.Info("Kubernetes cluster already exists")
		} else {
			err = k3d.CreateCluster()
			if err != nil {
				return err
			}
		}

		log.Info("Building authelia:dist image or use cache if already built...")

		if os.Getenv("CI") != t {
			if err := utils.Shell("authelia-scripts docker build").Run(); err != nil {
				return err
			}

			if err := utils.Shell("docker save authelia:dist -o internal/suites/example/kube/authelia-image-dev.tar").Run(); err != nil {
				return err
			}
		}

		log.Info("Loading images into Kubernetes container...")

		if err := loadDockerImages(); err != nil {
			return err
		}

		log.Info("Waiting for cluster to be ready...")

		if err := waitAllPodsAreReady(namespaceKube, 5*time.Minute); err != nil {
			return err
		}

		log.Info("Waiting for dashboard to be ready...")

		err = waitAllPodsAreReady(namespaceDashboard, 2*time.Minute)

		log.Info("Bearer token for UI user:")

		if err := kubectl.GetDashboardToken(); err != nil {
			return err
		}

		log.Info("Waiting for services to be ready...")

		if err := waitAllPodsAreReady(namespaceAuthelia, 5*time.Minute); err != nil {
			return err
		}

		return err
	}

	teardown := func(suitePath string) error {
		if err := k3d.DeleteCluster(); err != nil {
			return err
		}

		return dockerEnvironment.Down()
	}

	GlobalRegistry.Register(kubernetesSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    12 * time.Minute,
		TestTimeout:     2 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
		Description:     "This suite has been created to test Authelia in a Kubernetes context and using Traefik as the ingress controller.",
	})
}

func loadDockerImages() error {
	k3d := K3D{}
	images := []string{"/authelia/authelia-image-coverage.tar"}

	if os.Getenv("CI") != t {
		images = []string{"/authelia/authelia-image-dev.tar"}
	}

	for _, image := range images {
		err := k3d.LoadImage(image)

		if err != nil {
			return err
		}
	}

	return nil
}

func waitAllPodsAreReady(namespace string, timeout time.Duration) error {
	kubectl := Kubectl{}

	log.Infof("Checking services in %s namespace are running...", namespace)

	if err := kubectl.WaitPodsReady(namespace, timeout); err != nil {
		return err
	}

	log.Info("All pods are ready")

	return nil
}
