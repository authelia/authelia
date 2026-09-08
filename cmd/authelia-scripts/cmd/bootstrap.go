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

	// /etc/hosts is machine wide but the addresses behind those names are per slot, so a slotted shell must not claim
	// it: the tests resolve the suite domains themselves through suites.HostResolverRules and suites.DialContext, and
	// leaving the file alone is what keeps an unslotted working tree browsable while a slotted one runs.
	if slot := os.Getenv("SUITE_SLOT"); slot != "" {
		bootstrapPrintln("Leaving /etc/hosts alone because this shell holds suite slot " + slot)
	} else {
		bootstrapPrintln("Preparing /etc/hosts to serve subdomains of example.com...")
		prepareHostsFile()
	}

	fmt.Println()
	bootstrapPrintln("Run 'authelia-scripts suites setup Standalone' to start Authelia and visit https://home.example.com:8080.")
	bootstrapPrintln("More details at https://www.authelia.com/contributing/development/build-and-test/")
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
	err := os.MkdirAll(suites.SuiteTmpPath("authelia"), 0755)
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

func hostsLineOnlyWants(domains []string, ip string, addresses map[string]string) bool {
	for _, domain := range domains {
		if addresses[domain] != ip {
			return false
		}
	}

	return true
}

func hostsLineWithout(line, domain string) string {
	comment := ""

	if i := strings.IndexByte(line, '#'); i >= 0 {
		comment, line = " "+strings.TrimSpace(line[i:]), line[:i]
	}

	fields := strings.Fields(line)
	kept := make([]string, 0, len(fields))

	for _, field := range fields {
		if field != domain {
			kept = append(kept, field)
		}
	}

	return strings.Join(kept, " ") + comment
}

func prepareHostsFile() {
	contentBytes, err := readHostsFile()
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(contentBytes), "\n")
	toBeAddedLine := make([]string, 0)
	modified := false

	entries := suites.HostEntries()

	addresses := make(map[string]string, len(entries))

	for _, entry := range entries {
		addresses[entry.Domain] = entry.IP
	}

	for _, entry := range entries {
		domainInHostFile := false

		for i, line := range lines {
			ip, domains := hostsLineFields(line)

			if !slices.Contains(domains, entry.Domain) {
				continue
			}

			domainInHostFile = true

			if ip == entry.IP {
				break
			}

			// A line carries one address for every name on it, so rewriting it is only safe when all of those names
			// are ours and all of them want this address. Otherwise the domain moves to a line of its own and the
			// rest of the line, including any name we do not manage and any trailing comment, is left alone.
			if hostsLineOnlyWants(domains, entry.IP, addresses) {
				lines[i] = strings.Replace(line, ip, entry.IP, 1)
			} else {
				lines[i] = hostsLineWithout(line, entry.Domain)
				toBeAddedLine = append(toBeAddedLine, entry.IP+" "+entry.Domain)
			}

			modified = true

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

	fd, err := os.CreateTemp(suites.SuiteTmpPath("authelia"), "hosts")
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
