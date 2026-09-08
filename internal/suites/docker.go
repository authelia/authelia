package suites

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/utils"
)

// DockerEnvironment represent a docker environment.
type DockerEnvironment struct {
	dockerComposeFiles []string
}

func composeProjectName() string {
	if name := os.Getenv("COMPOSE_PROJECT_NAME"); name != "" {
		return name
	}

	return composeProjectDefault
}

// SuiteSubnet returns the first three octets of the suite network, matching the SUITE_SUBNET variable the compose files
// interpolate. Reading the same variable here means the addresses baked into Go and into /etc/hosts cannot drift from
// the ones compose assigns.
func SuiteSubnet() string {
	if subnet := os.Getenv("SUITE_SUBNET"); subnet != "" {
		return subnet
	}

	return suiteSubnetDefault
}

// SuiteAddress returns the address of the given host octet on the suite network.
func SuiteAddress(octet int) string {
	return fmt.Sprintf("%s.%d", SuiteSubnet(), octet)
}

// SuiteTmpPath joins elem onto the directory this process exchanges files with the suite containers through, matching
// the SUITE_TMP_PATH variable. The containers always see that directory at /tmp; SUITE_TMP is the host directory bound
// there, and SUITE_TMP_PATH is where this process finds the same content. The three coincide by default and in CI, and
// differ locally so that concurrent runs on one machine do not write over each other.
func SuiteTmpPath(elem ...string) string {
	path := os.Getenv("SUITE_TMP_PATH")
	if path == "" {
		path = suiteTmpPathDefault
	}

	return filepath.Join(append([]string{path}, elem...)...)
}

func proxyAccessLog() string {
	return fmt.Sprintf("traefik-access-%s.log", composeProjectName())
}

func removeProxyAccessLog() {
	// The proxy appends, and the directory it writes into outlives the containers, so a run would
	// otherwise collect every run before it.
	if err := os.Remove(SuiteTmpPath(proxyAccessLog())); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Debugf("Error removing the previous access log: %v", err)
	}
}

func agentContainer() string {
	return os.Getenv("AGENT_CONTAINER")
}

func autheliaNetworkName() string {
	return fmt.Sprintf("%s_%s", composeProjectName(), networkAuthelia)
}

func connectAgentNetwork() error {
	container := agentContainer()
	if container == "" {
		return nil
	}

	network := autheliaNetworkName()

	output, err := utils.Command("docker", "network", "connect", "--ip", SuiteAddress(agentAddressOctet), network, container).CombinedOutput()
	if err == nil {
		return nil
	}

	// A job killed mid-run leaves its attachment behind and the next connect reports it as an error.
	if strings.Contains(string(output), "already exists in network") {
		return nil
	}

	return fmt.Errorf("error connecting container '%s' to network '%s': %w: %s", container, network, err, output)
}

func disconnectAgentNetwork() {
	container := agentContainer()
	if container == "" {
		return
	}

	network := autheliaNetworkName()

	if output, err := utils.Command("docker", "network", "disconnect", "-f", network, container).CombinedOutput(); err != nil {
		log.Debugf("Error disconnecting container '%s' from network '%s': %v: %s", container, network, err, output)
	}
}

// NewDockerEnvironment create a new docker environment.
func NewDockerEnvironment(files []string) *DockerEnvironment {
	if os.Getenv("CI") == t {
		for i := range files {
			files[i] = strings.ReplaceAll(files[i], "{}", "dist")
		}
	} else {
		for i := range files {
			files[i] = strings.ReplaceAll(files[i], "{}", "dev")
		}
	}

	return &DockerEnvironment{dockerComposeFiles: files}
}

func (de *DockerEnvironment) createCommandWithStdout(cmd string) *exec.Cmd {
	dockerCmdLine := fmt.Sprintf("docker compose -p %s -f %s %s", composeProjectName(), strings.Join(de.dockerComposeFiles, " -f "), cmd)
	log.Trace(dockerCmdLine)

	return utils.CommandWithStdout("bash", "-c", dockerCmdLine)
}

func (de *DockerEnvironment) createCommand(cmd string) *exec.Cmd {
	dockerCmdLine := fmt.Sprintf("docker compose -p %s -f %s %s", composeProjectName(), strings.Join(de.dockerComposeFiles, " -f "), cmd)
	log.Trace(dockerCmdLine)

	return utils.Command("bash", "-c", dockerCmdLine)
}

// Pull pull all images of needed in the environment.
func (de *DockerEnvironment) Pull(images ...string) error {
	return de.createCommandWithStdout(fmt.Sprintf("pull %s", strings.Join(images, " "))).Run()
}

// Up spawn a docker environment.
func (de *DockerEnvironment) Up() error {
	removeProxyAccessLog()

	command := "up --build -d"

	if os.Getenv("CI") == t {
		command = "up --build --quiet-pull -d --wait --wait-timeout 300"
	}

	if err := de.createCommandWithStdout(command).Run(); err != nil {
		return err
	}

	// Chrome and the test process both run inside the agent container. On a shared daemon the suite network belongs
	// to the host namespace rather than the agent's, so the agent has to join it to reach the portal at all.
	return connectAgentNetwork()
}

// Restart restarts a service.
func (de *DockerEnvironment) Restart(service string) error {
	return de.createCommandWithStdout(fmt.Sprintf("restart %s", service)).Run()
}

// Stop a docker service.
func (de *DockerEnvironment) Stop(service string) error {
	return de.createCommandWithStdout(fmt.Sprintf("stop %s", service)).Run()
}

// Start a docker service.
func (de *DockerEnvironment) Start(service string) error {
	return de.createCommandWithStdout(fmt.Sprintf("start %s", service)).Run()
}

// Down destroy a docker environment.
func (de *DockerEnvironment) Down() error {
	disconnectAgentNetwork()

	return de.createCommandWithStdout("down -v").Run()
}

// Exec execute a command within a given service of the environment.
func (de *DockerEnvironment) Exec(service string, command []string) (string, error) {
	cmd := de.createCommand(fmt.Sprintf("exec -T %s %s", service, strings.Join(command, " ")))
	content, err := cmd.CombinedOutput()

	return string(content), err
}

// ExecWithEnv executes the given command against the given service with the given environment.
func (de *DockerEnvironment) ExecWithEnv(service string, command []string, env map[string]string) (string, error) {
	envs := make([]string, 0, len(env))

	for k, v := range env {
		envs = append(envs, fmt.Sprintf("-e %s=%s", k, v))
	}

	cmd := de.createCommand(fmt.Sprintf("exec %s -T %s %s", strings.Join(envs, " "), service, strings.Join(command, " ")))
	content, err := cmd.CombinedOutput()

	return string(content), err
}

// Logs get logs of a given service of the environment.
func (de *DockerEnvironment) Logs(service string, flags []string) (string, error) {
	cmd := de.createCommand(fmt.Sprintf("logs %s %s", strings.Join(flags, " "), service))
	content, err := cmd.Output()

	return string(content), err
}

// PrintLogs for the given service names.
func (de *DockerEnvironment) PrintLogs(services ...string) (err error) {
	var logs string

	for _, service := range services {
		if service == "authelia-frontend" && os.Getenv("CI") == t {
			continue
		}

		if logs, err = de.Logs(service, nil); err != nil {
			return err
		}

		fmt.Println(logs) //nolint:forbidigo
	}

	return nil
}
