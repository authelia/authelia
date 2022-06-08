package suites

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/utils"
)

var kindImageName = "authelia-kind-proxy"
var dockerCmdLine = fmt.Sprintf("docker-compose -p authelia -f internal/suites/docker-compose.yml -f internal/suites/example/compose/kind/docker-compose.yml run -T --rm %s", kindImageName)

// Kind used for running kind commands.
type Kind struct{}

func kindCommand(cmdline string) *exec.Cmd {
	cmd := fmt.Sprintf("%s %s", dockerCmdLine, cmdline)
	return utils.Shell(cmd)
}

// CreateCluster create a new Kubernetes cluster.
func (k Kind) CreateCluster() error {
	cmd := kindCommand("kind create cluster --config /etc/kind/config.yml")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = kindCommand("patch-kubeconfig.sh")
	if err := cmd.Run(); err != nil {
		return err
	}

	// This command is necessary to fix the coredns loop detected when using user-defined docker network.
	// In that case /etc/resolv.conf use 127.0.0.11 as DNS and CoreDNS thinks it is talking to itself which is wrong.
	// This IP is the docker internal DNS so it is safe to disable the loop check.
	cmd = kindCommand("sh -c 'kubectl -n kube-system get configmap/coredns -o yaml | grep -v loop | kubectl replace -f -'")
	err := cmd.Run()

	return err
}

// DeleteCluster delete a Kubernetes cluster.
func (k Kind) DeleteCluster() error {
	cmd := kindCommand("kind delete cluster")
	return cmd.Run()
}

// ClusterExists check whether a cluster exists.
func (k Kind) ClusterExists() (bool, error) {
	cmd := kindCommand("kind get clusters")
	cmd.Stdout = nil
	cmd.Stderr = nil
	output, err := cmd.Output()

	if err != nil {
		return false, err
	}

	return strings.Contains(string(output), "kind"), nil
}

// LoadImage load an image in the Kubernetes container.
func (k Kind) LoadImage(imageName string) error {
	cmd := kindCommand(fmt.Sprintf("kind load docker-image %s", imageName))
	return cmd.Run()
}

// Kubectl used for running kubectl commands.
type Kubectl struct{}

// StartProxy start a proxy.
func (k Kubectl) StartProxy() error {
	cmd := utils.Shell("docker-compose -p authelia -f internal/suites/docker-compose.yml -f internal/suites/example/compose/kind/docker-compose.yml up -d authelia-kind-proxy")
	return cmd.Run()
}

// StopProxy stop a proxy.
func (k Kubectl) StopProxy() error {
	cmd := utils.Shell("docker-compose -p authelia -f internal/suites/docker-compose.yml -f internal/suites/example/compose/kind/docker-compose.yml rm -s -f authelia-kind-proxy")
	return cmd.Run()
}

// StartDashboard start Kube dashboard.
func (k Kubectl) StartDashboard() error {
	if err := kindCommand("sh -c 'cd /authelia && ./bootstrap-dashboard.sh'").Run(); err != nil {
		return err
	}

	err := utils.Shell("docker-compose -p authelia -f internal/suites/docker-compose.yml -f internal/suites/example/compose/kind/docker-compose.yml up -d kube-dashboard").Run()

	return err
}

// StopDashboard stop kube dashboard.
func (k Kubectl) StopDashboard() error {
	cmd := utils.Shell("docker-compose -p authelia -f internal/suites/docker-compose.yml -f internal/suites/example/compose/kind/docker-compose.yml rm -s -f kube-dashboard")
	return cmd.Run()
}

// DeployThirdparties deploy thirdparty services (ldap, db, ingress controllers, etc...).
func (k Kubectl) DeployThirdparties() error {
	cmd := kindCommand("sh -c 'cd /authelia && ./bootstrap.sh'")
	return cmd.Run()
}

// DeployAuthelia deploy Authelia application.
func (k Kubectl) DeployAuthelia() error {
	cmd := kindCommand("sh -c 'cd /authelia && ./bootstrap-authelia.sh'")
	return cmd.Run()
}

// WaitPodsReady wait for all pods to be ready.
func (k Kubectl) WaitPodsReady(timeout time.Duration) error {
	return utils.CheckUntil(5*time.Second, timeout, func() (bool, error) {
		cmd := kindCommand("kubectl get -n authelia pods --no-headers")
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
			if !strings.Contains(line, "1/1") {
				return false, nil
			}
		}
		return true, nil
	})
}
