package suites

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/utils"
)

var (
	k3dImageName  = "k3d"
	dockerCmdLine = fmt.Sprintf("docker-compose -p authelia -f internal/suites/docker-compose.yml -f internal/suites/example/compose/k3d/docker-compose.yml exec -T %s", k3dImageName)
)

// K3D used for running kind commands.
type K3D struct{}

func k3dCommand(cmdline string) *exec.Cmd {
	cmd := fmt.Sprintf("%s %s", dockerCmdLine, cmdline)
	return utils.Shell(cmd)
}

// CreateCluster create a new Kubernetes cluster.
func (k K3D) CreateCluster() error {
	cmd := k3dCommand("k3d cluster create --registry-config /authelia/registry.yml -v /authelia:/var/lib/rancher/k3s/server/manifests/custom -v /configmaps:/configmaps -p 8080:443")
	err := cmd.Run()

	return err
}

// DeleteCluster delete a Kubernetes cluster.
func (k K3D) DeleteCluster() error {
	cmd := k3dCommand("k3d cluster delete")
	return cmd.Run()
}

// ClusterExists check whether a cluster exists.
func (k K3D) ClusterExists() (bool, error) {
	cmd := k3dCommand("k3d cluster list")
	cmd.Stdout = nil
	cmd.Stderr = nil
	output, err := cmd.Output()

	if err != nil {
		return false, err
	}

	return strings.Contains(string(output), "k3s-default"), nil
}

// LoadImage load an image in the Kubernetes container.
func (k K3D) LoadImage(imageName string) error {
	cmd := k3dCommand(fmt.Sprintf("k3d image import %s", imageName))
	return cmd.Run()
}

// Kubectl used for running kubectl commands.
type Kubectl struct{}

// GetDashboardToken generates bearer token for Kube Dashboard.
func (k Kubectl) GetDashboardToken() error {
	return k3dCommand("kubectl -n kubernetes-dashboard create token admin-user;echo ''").Run()
}

// WaitPodsReady wait for all pods to be ready.
func (k Kubectl) WaitPodsReady(namespace string, timeout time.Duration) error {
	return utils.CheckUntil(5*time.Second, timeout, func() (bool, error) {
		cmd := k3dCommand(fmt.Sprintf("kubectl get -n %s pods --no-headers --field-selector=status.phase!=Succeeded", namespace))
		cmd.Stdout = nil
		cmd.Stderr = nil
		output, _ := cmd.Output()

		lines := strings.Split(string(output), "\n")

		nonEmptyLines := make([]string, 0)

		for _, line := range lines {
			if line != "" {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}

		for _, line := range nonEmptyLines {
			re := regexp.MustCompile(`1/1|2/2`)
			if !re.MatchString(line) {
				return false, nil
			}
		}

		return true, nil
	})
}
