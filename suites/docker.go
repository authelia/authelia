package suites

import (
	"os"
	"os/exec"
	"strings"

	"github.com/clems4ever/authelia/utils"
	log "github.com/sirupsen/logrus"
)

// DockerEnvironment represent a docker environment
type DockerEnvironment struct {
	dockerComposeFiles []string
}

// NewDockerEnvironment create a new docker environment
func NewDockerEnvironment(files []string) *DockerEnvironment {
	return &DockerEnvironment{dockerComposeFiles: files}
}

func (de *DockerEnvironment) createCommandWithStdout(cmd string) *exec.Cmd {
	dockerCmdLine := "docker-compose -f " + strings.Join(de.dockerComposeFiles, " -f ") + " " + cmd
	log.Trace(dockerCmdLine)
	return utils.CommandWithStdout("bash", "-c", dockerCmdLine)
}

func (de *DockerEnvironment) createCommand(cmd string) *exec.Cmd {
	dockerCmdLine := "docker-compose -f " + strings.Join(de.dockerComposeFiles, " -f ") + " " + cmd
	log.Trace(dockerCmdLine)
	return exec.Command("bash", "-c", dockerCmdLine)
}

// Up spawn a docker environment
func (de *DockerEnvironment) Up(suitePath string) error {
	cmd := de.createCommandWithStdout("up -d")
	cmd.Env = append(os.Environ(), "SUITE_PATH="+suitePath)
	return cmd.Run()
}

// Down spawn a docker environment
func (de *DockerEnvironment) Down(suitePath string) error {
	cmd := de.createCommandWithStdout("down -v")
	cmd.Env = append(os.Environ(), "SUITE_PATH="+suitePath)
	return cmd.Run()
}

// Logs get logs of a given service of the environment
func (de *DockerEnvironment) Logs(service string, flags []string) (string, error) {
	cmd := de.createCommand("logs " + strings.Join(flags, " ") + " " + service)
	content, err := cmd.Output()
	return string(content), err
}
