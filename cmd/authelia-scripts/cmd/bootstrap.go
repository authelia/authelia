package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/suites"
	"github.com/authelia/authelia/v4/internal/utils"
)

func newBootstrapCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "bootstrap",
		Short:   cmdBootstrapShort,
		Long:    cmdBootstrapLong,
		Example: cmdBootstrapExample,
		Args:    cobra.NoArgs,
		Run:     cmdBootstrapRun,

		DisableAutoGenTag: true,
	}

	return cmd
}

func cmdBootstrapRun(_ *cobra.Command, _ []string) {
	bootstrapPrintln("Checking command installation...")
	checkCommandExist("node", "Follow installation guidelines from https://nodejs.org/en/download")
	checkCommandExist("pnpm", "Follow installation guidelines from https://pnpm.io/installation")
	checkCommandExist("docker", "Follow installation guidelines from https://docs.docker.com/get-docker/")
	checkCommandExist("docker compose", "Follow installation guidelines from https://docs.docker.com/compose/install/")

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

	if os.Getenv("CI") != stringTrue {
		createPNPMDirectory()
		pnpmInstall()
	}

	bootstrapPrintln("Preparing /etc/hosts to serve subdomains of example.com...")
	prepareHostsFile()

	fmt.Println()
	bootstrapPrintln("Run 'authelia-scripts suites setup Standalone' to start Authelia and visit https://home.example.com:8080.")
	bootstrapPrintln("More details at https://www.authelia.com/contributing/development/build-and-test/")
}

func hostEntries() []HostEntry {
	backend := suites.SuiteAddress(50)
	portal := suites.SuiteAddress(100)
	ssh := suites.SuiteAddress(130)

	entries := []HostEntry{
		// For unit tests.
		{Domain: "local.example.com", IP: "127.0.0.1"},

		// For authelia backend.
		{Domain: "authelia.example.com", IP: backend},

		// For common tests.
		{Domain: "login.example.com", IP: portal},
		{Domain: "admin.example.com", IP: portal},
		{Domain: "singlefactor.example.com", IP: portal},
		{Domain: "deny.example.com", IP: portal},
		{Domain: "dev.example.com", IP: portal},
		{Domain: "home.example.com", IP: portal},
		{Domain: "mx1.mail.example.com", IP: portal},
		{Domain: "mx2.mail.example.com", IP: portal},
		{Domain: "public.example.com", IP: portal},
		{Domain: "secure.example.com", IP: portal},
		{Domain: "mail.example.com", IP: portal},
		{Domain: "duo.example.com", IP: portal},

		// For HAProxy suite.
		{Domain: "haproxy.example.com", IP: portal},

		// Kubernetes dashboard.
		{Domain: "kubernetes.example.com", IP: portal},

		// OIDC tester app.
		{Domain: "oidc.example.com", IP: portal},
		{Domain: "oidc-public.example.com", IP: portal},

		// For Traefik suite.
		{Domain: "traefik.example.com", IP: portal},

		// For testing network ACLs.
		{Domain: "proxy-client1.example.com", IP: suites.SuiteAddress(201)},
		{Domain: "proxy-client2.example.com", IP: suites.SuiteAddress(202)},
		{Domain: "proxy-client3.example.com", IP: suites.SuiteAddress(203)},

		// Redis Replicas.
		{Domain: "redis-node-0.example.com", IP: suites.SuiteAddress(110)},
		{Domain: "redis-node-1.example.com", IP: suites.SuiteAddress(111)},
		{Domain: "redis-node-2.example.com", IP: suites.SuiteAddress(112)},

		// Redis Sentinel Replicas.
		{Domain: "redis-sentinel-0.example.com", IP: suites.SuiteAddress(120)},
		{Domain: "redis-sentinel-1.example.com", IP: suites.SuiteAddress(121)},
		{Domain: "redis-sentinel-2.example.com", IP: suites.SuiteAddress(122)},

		// For PAM suite.
		{Domain: "ssh.example.com", IP: ssh},
	}

	return slices.Concat(entries, hostEntriesCookieDomains(portal))
}

func hostEntriesCookieDomains(ip string) []HostEntry {
	domains := []string{"example2.com", "example3.com"}
	subdomains := []string{"login", "admin", "singlefactor", "dev", "home", "mx1.mail", "mx2.mail", "public", "secure", "mail", "duo"}

	entries := make([]HostEntry, 0, len(domains)*len(subdomains))

	for _, domain := range domains {
		for _, subdomain := range subdomains {
			entries = append(entries, HostEntry{Domain: subdomain + "." + domain, IP: ip})
		}
	}

	return entries
}

func runCommand(cmd string, args ...string) {
	command := utils.CommandWithStdout(cmd, args...)

	err := command.Run()
	if err != nil {
		panic(err)
	}
}

func checkCommandExist(cmd string, resolutionHint string) {
	fmt.Print("Checking if '" + cmd + "' command is installed...")
	command := exec.Command("bash", "-c", "command -v "+cmd) //nolint:gosec // Used only in development.

	if command.Run() != nil {
		msg := "[ERROR] You must install " + cmd + " on your machine."
		if resolutionHint != "" {
			msg += fmt.Sprintf(" %s", resolutionHint)
		}

		log.Fatal(msg)
	}

	fmt.Println("		OK")
}

func createTemporaryDirectory() {
	err := os.MkdirAll("/tmp/authelia", 0755)
	if err != nil {
		panic(err)
	}
}

func createPNPMDirectory() {
	if _, ok := os.LookupEnv("PNPM_HOME"); !ok {
		home := os.Getenv("HOME")
		if home != "" {
			if _, err := os.Stat(home + pathPNPMStore); os.IsNotExist(err) { //nolint:gosec // TODO: Run this line through taint analysis.
				bootstrapPrintln("Creating ", home+pathPNPMStore)

				if err = os.MkdirAll(home+pathPNPMStore, 0755); err != nil { //nolint:gosec // TODO: Run this line through taint analysis.
					panic(err)
				}
			}
		}
	}
}

func pnpmInstall() {
	bootstrapPrintln("Installing web dependencies ")

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if _, err = os.Stat(cwd + pathPNPMModule); err == nil {
		if err = os.Remove(cwd + pathPNPMModule); err != nil {
			panic(err)
		}
	}

	shell(fmt.Sprintf("cd %s/web && pnpm install", cwd))
}

func bootstrapPrintln(args ...any) {
	a := make([]any, 0, 1+len(args))
	a = append(a, "[BOOTSTRAP]")
	a = append(a, args...)
	fmt.Println(a...)
}

func shell(cmd string) {
	runCommand("bash", "-c", cmd)
}

func hostsLineFields(line string) (ip string, domains []string) {
	if i := strings.IndexByte(line, '#'); i >= 0 {
		line = line[:i]
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", nil
	}

	return fields[0], fields[1:]
}

func prepareHostsFile() {
	contentBytes, err := readHostsFile()
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(contentBytes), "\n")
	toBeAddedLine := make([]string, 0)
	modified := false

	for _, entry := range hostEntries() {
		domainInHostFile := false

		for i, line := range lines {
			ip, domains := hostsLineFields(line)

			if !slices.Contains(domains, entry.Domain) {
				continue
			}

			domainInHostFile = true

			// The IP is not up to date. Only the address is rewritten so any other names shared with this line, and
			// any trailing comment, survive.
			if ip != entry.IP {
				lines[i] = strings.Replace(line, ip, entry.IP, 1)
				modified = true
			}

			break
		}

		if !domainInHostFile {
			toBeAddedLine = append(toBeAddedLine, entry.IP+" "+entry.Domain)
		}
	}

	if len(toBeAddedLine) > 0 {
		lines = append(lines, toBeAddedLine...)
		modified = true
	}

	fd, err := os.CreateTemp("/tmp/authelia/", "hosts")
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
	bs, err := os.ReadFile("/etc/hosts")
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
	readVersion("pnpm", "--version")
	readVersion("docker", "--version")
	readVersion("docker", "compose", "version")
}
