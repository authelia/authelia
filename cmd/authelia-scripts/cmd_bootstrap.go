package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// HostEntry represents an entry in /etc/hosts
type HostEntry struct {
	Domain string
	IP     string
}

var hostEntries = []HostEntry{
	// For common tests
	HostEntry{Domain: "login.example.com", IP: "192.168.240.100"},
	HostEntry{Domain: "admin.example.com", IP: "192.168.240.100"},
	HostEntry{Domain: "singlefactor.example.com", IP: "192.168.240.100"},
	HostEntry{Domain: "dev.example.com", IP: "192.168.240.100"},
	HostEntry{Domain: "home.example.com", IP: "192.168.240.100"},
	HostEntry{Domain: "mx1.mail.example.com", IP: "192.168.240.100"},
	HostEntry{Domain: "mx2.mail.example.com", IP: "192.168.240.100"},
	HostEntry{Domain: "public.example.com", IP: "192.168.240.100"},
	HostEntry{Domain: "secure.example.com", IP: "192.168.240.100"},
	HostEntry{Domain: "mail.example.com", IP: "192.168.240.100"},
	HostEntry{Domain: "duo.example.com", IP: "192.168.240.100"},

	// For Traefik suite
	HostEntry{Domain: "traefik.example.com", IP: "192.168.240.100"},

	// For testing network ACLs
	HostEntry{Domain: "proxy-client1.example.com", IP: "192.168.240.201"},
	HostEntry{Domain: "proxy-client2.example.com", IP: "192.168.240.202"},
	HostEntry{Domain: "proxy-client3.example.com", IP: "192.168.240.203"},
}

func runCommand(cmd string, args ...string) {
	command := CommandWithStdout(cmd, args...)
	err := command.Run()

	if err != nil {
		panic(err)
	}
}

func installNpmPackages() {
	runCommand("npm", "ci")
}

func checkCommandExist(cmd string) {
	fmt.Print("Checking if '" + cmd + "' command is installed...")
	command := exec.Command("bash", "-c", "command -v "+cmd)
	err := command.Run()

	if err != nil {
		log.Fatal("[ERROR] You must install " + cmd + " on your machine.")
	}

	fmt.Println("		OK")
}

func installClientNpmPackages() {
	command := CommandWithStdout("npm", "ci")
	command.Dir = "client"
	err := command.Run()

	if err != nil {
		panic(err)
	}
}

func createTemporaryDirectory() {
	err := os.MkdirAll("/tmp/authelia", 0755)

	if err != nil {
		panic(err)
	}
}

func bootstrapPrintln(args ...interface{}) {
	a := make([]interface{}, 0)
	a = append(a, "[BOOTSTRAP]")
	a = append(a, args...)
	fmt.Println(a...)
}

func shell(cmd string) {
	runCommand("bash", "-c", cmd)
}

func buildHelperDockerImages() {
	shell("docker build -t authelia-example-backend example/compose/nginx/backend")
	shell("docker build -t authelia-duo-api example/compose/duo-api")
}

func installKubernetesDependencies() {
	if exist, err := FileExists("/tmp/kind"); err == nil && !exist {
		shell("wget -nv https://github.com/kubernetes-sigs/kind/releases/download/v0.5.1/kind-linux-amd64 -O /tmp/kind && chmod +x /tmp/kind")
	} else {
		bootstrapPrintln("Skip installing Kind since it's already installed")
	}

	if exist, err := FileExists("/tmp/kubectl"); err == nil && !exist {
		shell("wget -nv https://storage.googleapis.com/kubernetes-release/release/v1.13.0/bin/linux/amd64/kubectl -O /tmp/kubectl && chmod +x /tmp/kubectl")
	} else {
		bootstrapPrintln("Skip installing Kubectl since it's already installed")
	}
}

func prepareHostsFile() {
	contentBytes, err := readHostsFile()

	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(contentBytes), "\n")
	toBeAddedLine := make([]string, 0)
	modified := false

	for _, entry := range hostEntries {
		domainInHostFile := false
		for i, line := range lines {
			domainFound := strings.Contains(line, entry.Domain)
			ipFound := strings.Contains(line, entry.IP)

			if domainFound {
				domainInHostFile = true

				// The IP is not up to date.
				if ipFound {
					break
				} else {
					lines[i] = entry.IP + " " + entry.Domain
					modified = true
					break
				}
			}
		}

		if !domainInHostFile {
			toBeAddedLine = append(toBeAddedLine, entry.IP+" "+entry.Domain)
		}
	}

	if len(toBeAddedLine) > 0 {
		lines = append(lines, toBeAddedLine...)
		modified = true
	}

	err = ioutil.WriteFile("/tmp/authelia/hosts", []byte(strings.Join(lines, "\n")), 0644)

	if err != nil {
		panic(err)
	}

	if modified {
		bootstrapPrintln("/etc/hosts needs to be updated")
		shell("/usr/bin/sudo mv /tmp/authelia/hosts /etc/hosts")
	}
}

// ReadHostsFile reads the hosts file.
func readHostsFile() ([]byte, error) {
	bs, err := ioutil.ReadFile("/etc/hosts")
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func readVersion(cmd string, args ...string) {
	command := exec.Command(cmd, args...)
	b, err := command.Output()

	if err != nil {
		panic(err)
	}

	fmt.Print(cmd + " => " + string(b))
}

func readVersions() {
	readVersion("go", "version")
	readVersion("node", "--version")
	readVersion("docker", "--version")
	readVersion("docker-compose", "--version")
}

// Bootstrap bootstrap authelia dev environment
func Bootstrap(cobraCmd *cobra.Command, args []string) {
	bootstrapPrintln("Checking command installation...")
	checkCommandExist("node")
	checkCommandExist("docker")
	checkCommandExist("docker-compose")

	bootstrapPrintln("Getting versions of tools")
	readVersions()

	bootstrapPrintln("Installing NPM packages for development...")
	installNpmPackages()

	bootstrapPrintln("Install NPM packages for frontend...")
	installClientNpmPackages()

	bootstrapPrintln("Building development Docker images...")
	buildHelperDockerImages()

	bootstrapPrintln("Installing Kubernetes dependencies for testing in /tmp... (no junk installed on host)")
	installKubernetesDependencies()

	createTemporaryDirectory()

	bootstrapPrintln("Preparing /etc/hosts to serve subdomains of example.com...")
	prepareHostsFile()

	bootstrapPrintln("Run 'authelia-scripts suites start docker-image' to start Authelia and visit https://home.example.com:8080.")
	bootstrapPrintln("More details at https://github.com/clems4ever/authelia/blob/master/docs/getting-started.md")
}
