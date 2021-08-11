package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

// HostEntry represents an entry in /etc/hosts.
type HostEntry struct {
	Domain string
	IP     string
}

var hostEntries = []HostEntry{
	// For authelia backend.
	{Domain: "authelia.example.com", IP: "192.168.240.50"},

	// For common tests.
	{Domain: "login.example.com", IP: "192.168.240.100"},
	{Domain: "admin.example.com", IP: "192.168.240.100"},
	{Domain: "singlefactor.example.com", IP: "192.168.240.100"},
	{Domain: "dev.example.com", IP: "192.168.240.100"},
	{Domain: "home.example.com", IP: "192.168.240.100"},
	{Domain: "mx1.mail.example.com", IP: "192.168.240.100"},
	{Domain: "mx2.mail.example.com", IP: "192.168.240.100"},
	{Domain: "public.example.com", IP: "192.168.240.100"},
	{Domain: "secure.example.com", IP: "192.168.240.100"},
	{Domain: "mail.example.com", IP: "192.168.240.100"},
	{Domain: "duo.example.com", IP: "192.168.240.100"},

	// For Traefik suite.
	{Domain: "traefik.example.com", IP: "192.168.240.100"},

	// For HAProxy suite.
	{Domain: "haproxy.example.com", IP: "192.168.240.100"},

	// For testing network ACLs.
	{Domain: "proxy-client1.example.com", IP: "192.168.240.201"},
	{Domain: "proxy-client2.example.com", IP: "192.168.240.202"},
	{Domain: "proxy-client3.example.com", IP: "192.168.240.203"},

	// Redis Replicas
	{Domain: "redis-node-0.example.com", IP: "192.168.240.110"},
	{Domain: "redis-node-1.example.com", IP: "192.168.240.111"},
	{Domain: "redis-node-2.example.com", IP: "192.168.240.112"},

	// Redis Sentinel Replicas
	{Domain: "redis-sentinel-0.example.com", IP: "192.168.240.120"},
	{Domain: "redis-sentinel-1.example.com", IP: "192.168.240.121"},
	{Domain: "redis-sentinel-2.example.com", IP: "192.168.240.122"},

	// Kubernetes dashboard.
	{Domain: "kubernetes.example.com", IP: "192.168.240.110"},
	// OIDC tester app
	{Domain: "oidc.example.com", IP: "192.168.240.100"},
	{Domain: "oidc-public.example.com", IP: "192.168.240.100"},
}

func runCommand(cmd string, args ...string) {
	command := utils.CommandWithStdout(cmd, args...)
	err := command.Run()

	if err != nil {
		panic(err)
	}
}

func checkCommandExist(cmd string) {
	fmt.Print("Checking if '" + cmd + "' command is installed...")
	command := exec.Command("bash", "-c", "command -v "+cmd) //nolint:gosec // Used only in development.
	err := command.Run()

	if err != nil {
		log.Fatal("[ERROR] You must install " + cmd + " on your machine.")
	}

	fmt.Println("		OK")
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

	fd, err := ioutil.TempFile("/tmp/authelia/", "hosts")
	if err != nil {
		panic(err)
	}

	_, err = fd.Write([]byte(strings.Join(lines, "\n")))
	if err != nil {
		panic(err)
	}

	if modified {
		bootstrapPrintln("/etc/hosts needs to be updated")
		shell(fmt.Sprintf("cat %s | sudo tee /etc/hosts > /dev/null", fd.Name()))
	}

	err = fd.Close()
	if err != nil {
		panic(err)
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

// Bootstrap bootstrap authelia dev environment.
func Bootstrap(cobraCmd *cobra.Command, args []string) {
	bootstrapPrintln("Checking command installation...")
	checkCommandExist("node")
	checkCommandExist("docker")
	checkCommandExist("docker-compose")

	bootstrapPrintln("Getting versions of tools")
	readVersions()

	bootstrapPrintln("Checking if GOPATH is set")

	goPathFound := false

	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "GOPATH=") {
			goPathFound = true
			break
		}
	}

	if !goPathFound {
		log.Fatal("GOPATH is not set")
	}

	createTemporaryDirectory()

	bootstrapPrintln("Preparing /etc/hosts to serve subdomains of example.com...")
	prepareHostsFile()

	fmt.Println()
	bootstrapPrintln("Run 'authelia-scripts suites setup Standalone' to start Authelia and visit https://home.example.com:8080.")
	bootstrapPrintln("More details at https://github.com/authelia/authelia/blob/master/docs/getting-started.md")
}
