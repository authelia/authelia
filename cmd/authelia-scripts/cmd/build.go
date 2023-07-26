package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func newBuildCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "build",
		Short:   cmdBuildShort,
		Long:    cmdBuildLong,
		Example: cmdBuildExample,
		Args:    cobra.NoArgs,
		Run:     cmdBuildRun,

		DisableAutoGenTag: true,
	}

	cmd.Flags().Bool("print", false, "Prints the command instead of running it, useful for reproducible builds")
	cmd.Flags().Int("build-number", 0, "Forcefully sets the build number, useful for reproducible builds")

	return cmd
}

func cmdBuildRun(cmd *cobra.Command, args []string) {
	branch := os.Getenv("BUILDKITE_BRANCH")

	var (
		buildPrint bool
		err        error
	)

	if buildPrint, err = cmd.Flags().GetBool("print"); err != nil {
		log.Fatal(err)
	}

	if strings.HasPrefix(branch, "renovate/") {
		buildFrontend(false, branch)
		log.Info("Skip building Authelia for deps...")
		os.Exit(0)
	}

	switch {
	case buildPrint:
		log.Info("Printing Build Authelia Commands...")
	default:
		log.Info("Building Authelia...")

		cmdCleanRun(cmd, args)

		log.Debug("Creating `" + OutputDir + "` directory")

		if err = os.MkdirAll(OutputDir, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	buildMetaData, err := getBuild(branch, os.Getenv("BUILDKITE_BUILD_NUMBER"), "")
	if err != nil {
		log.Fatal(err)
	}

	if cmd.Flags().Changed("build-number") {
		buildMetaData.Number, _ = cmd.Flags().GetInt("build-number")
	}

	log.Debug("Building Authelia frontend...")
	buildFrontend(buildPrint, branch)

	log.Debug("Building swagger-ui frontend...")
	buildSwagger(buildPrint)

	buildkite, _ := cmd.Flags().GetBool("buildkite")

	if buildkite {
		buildAutheliaBinaryGOX(buildPrint, buildMetaData)
	} else {
		buildAutheliaBinaryGO(buildPrint, buildMetaData)
	}

	if !buildPrint {
		cleanAssets()
	}
}

func buildAutheliaBinaryGOX(buildPrint bool, buildMetaData *Build) {
	var wg sync.WaitGroup

	started := time.Now()

	xflags := buildMetaData.XFlags()

	cmds := make([]*exec.Cmd, 2)

	cmds[0] = utils.CommandWithStdout("gox", "-output={{.Dir}}-{{.OS}}-{{.Arch}}-musl", "-buildmode=pie", "-buildvcs=false", "-trimpath", "-cgo", "-ldflags=-linkmode=external -buildid= -s -w "+strings.Join(xflags, " "), "-osarch=linux/amd64 linux/arm linux/arm64", "./cmd/authelia/")

	cmds[0].Env = append(cmds[0].Env,
		"CGO_CPPFLAGS=-D_FORTIFY_SOURCE=2 -fstack-protector-strong", "CGO_LDFLAGS=-Wl,-z,relro,-z,now",
		"GOX_LINUX_ARM_CC=arm-linux-musleabihf-gcc", "GOX_LINUX_ARM64_CC=aarch64-linux-musl-gcc")

	cmds[1] = utils.CommandWithStdout("bash", "-c", "docker run --rm -e GOX_LINUX_ARM_CC=arm-linux-gnueabihf-gcc -e GOX_LINUX_ARM64_CC=aarch64-linux-gnu-gcc -e GOX_FREEBSD_AMD64_CC=x86_64-pc-freebsd13-gcc -v ${PWD}:/workdir -v /buildkite/.go:/root/go authelia/crossbuild "+
		"gox -output={{.Dir}}-{{.OS}}-{{.Arch}} -buildmode=pie -buildvcs=false -trimpath -cgo -ldflags=\"-linkmode=external -buildid= -s -w "+strings.Join(xflags, " ")+"\" -osarch=\"linux/amd64 linux/arm linux/arm64 freebsd/amd64\" ./cmd/authelia/")

	if buildPrint {
		for _, cmd := range cmds {
			buildCmdPrint(cmd)
		}

		return
	}

	log.Info("Building Authelia Go binaries with gox...")

	wg.Add(len(cmds))

	go func() {
		defer wg.Done()

		buildCmdRun(cmds[0])
	}()

	go func() {
		defer wg.Done()

		buildCmdRun(cmds[1])
	}()

	wg.Wait()

	log.Debugf("Binary compilation completed in %s.", time.Since(started))
}

func buildAutheliaBinaryGO(buildPrint bool, buildMetaData *Build) {
	cmd := utils.CommandWithStdout("go", "build", "-buildmode=pie", "-buildvcs=false", "-trimpath", "-o", OutputDir+"/authelia", "-ldflags", "-linkmode=external -buildid= -s -w "+strings.Join(buildMetaData.XFlags(), " "), "./cmd/authelia/")

	cmd.Env = append(cmd.Env,
		"CGO_CPPFLAGS=-D_FORTIFY_SOURCE=2 -fstack-protector-strong", "CGO_LDFLAGS=-Wl,-z,relro,-z,now")

	if buildPrint {
		buildCmdPrint(cmd)

		return
	}

	log.Info("Building Authelia Go binary...")

	buildCmdRun(cmd)
}

func buildFrontend(buildPrint bool, branch string) {
	var (
		cmds []*exec.Cmd
		cmd  *exec.Cmd
	)

	cmd = utils.CommandWithStdout("pnpm", "install")
	cmd.Dir = filepath.Join(cmd.Dir, webDirectory)

	cmds = append(cmds, cmd)

	if !strings.HasPrefix(branch, "renovate/") {
		cmd = utils.CommandWithStdout("pnpm", "build")
		cmd.Dir = filepath.Join(cmd.Dir, webDirectory)

		cmds = append(cmds, cmd)
	}

	for _, cmd = range cmds {
		if buildPrint {
			buildCmdPrint(cmd)

			continue
		}

		buildCmdRun(cmd)
	}
}

func buildSwagger(buildPrint bool) {
	var (
		cmds []*exec.Cmd
		cmd  *exec.Cmd
	)

	cmds = append(cmds, utils.CommandWithStdout("bash", "-c", "wget -q https://github.com/swagger-api/swagger-ui/archive/v"+versionSwaggerUI+".tar.gz -O ./v"+versionSwaggerUI+".tar.gz"))
	cmds = append(cmds, utils.CommandWithStdout("cp", "-r", "api", "internal/server/public_html"))
	cmds = append(cmds, utils.CommandWithStdout("tar", "-C", "internal/server/public_html/api", "--exclude=index.html", "--strip-components=2", "-xf", "v"+versionSwaggerUI+".tar.gz", "swagger-ui-"+versionSwaggerUI+"/dist"))
	cmds = append(cmds, utils.CommandWithStdout("rm", "./v"+versionSwaggerUI+".tar.gz"))

	for _, cmd = range cmds {
		if buildPrint {
			buildCmdPrint(cmd)

			continue
		}

		buildCmdRun(cmd)
	}
}

func cleanAssets() {
	if err := os.Rename("internal/server/public_html", OutputDir+"/public_html"); err != nil {
		log.Fatal(err)
	}

	cmd := utils.CommandWithStdout("mkdir", "-p", "internal/server/public_html/api")

	buildCmdRun(cmd)

	cmd = utils.CommandWithStdout("bash", "-c", "touch internal/server/public_html/{index.html,api/index.html,api/openapi.yml}")

	buildCmdRun(cmd)
}

func buildCmdRun(cmd *exec.Cmd) {
	if len(cmd.Env) != 0 {
		cmd.Env = append(os.Environ(), cmd.Env...)
	}

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func buildCmdPrint(cmd *exec.Cmd) {
	b := &strings.Builder{}

	if cmd.Dir != "" {
		b.WriteString(fmt.Sprintf("cd %s\n", cmd.Dir))
	}

	buildCmdWriteCmd(b, cmd)

	fmt.Println(b.String())
}

func buildCmdWriteCmd(wr io.StringWriter, cmd *exec.Cmd) {
	buildCmdWriteEnv(wr, cmd)

	_, _ = wr.WriteString(cmd.Path)

	for _, arg := range cmd.Args[1:] {
		_, _ = wr.WriteString(" ")
		if strings.Contains(arg, " ") {
			_, _ = wr.WriteString(fmt.Sprintf(`"%s"`, arg))
		} else {
			_, _ = wr.WriteString(arg)
		}
	}
}

func buildCmdWriteEnv(wr io.StringWriter, cmd *exec.Cmd) {
	n := len(cmd.Env)

	if n == 0 {
		return
	}

	envs := make([]string, n)

	for i, env := range cmd.Env {
		parts := strings.SplitN(env, "=", 2)

		envs[i] = fmt.Sprintf(`%s="%s"`, parts[0], parts[1])
	}

	_, _ = wr.WriteString(strings.Join(envs, " "))
	_, _ = wr.WriteString(" ")
}
