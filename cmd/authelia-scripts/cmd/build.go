package cmd

import (
	"os"
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

	return cmd
}

func cmdBuildRun(cobraCmd *cobra.Command, args []string) {
	branch := os.Getenv("BUILDKITE_BRANCH")

	if strings.HasPrefix(branch, "renovate/") {
		buildFrontend(branch)
		log.Info("Skip building Authelia for deps...")
		os.Exit(0)
	}

	log.Info("Building Authelia...")

	cmdCleanRun(cobraCmd, args)

	buildMetaData, err := getBuild(branch, os.Getenv("BUILDKITE_BUILD_NUMBER"), "")
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Creating `" + OutputDir + txtDirectoryTidle)

	if err = os.MkdirAll(OutputDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	log.Debug("Building Authelia frontend...")
	buildFrontend(branch)

	log.Debug("Building swagger-ui frontend...")
	buildSwagger()

	buildkite, _ := cobraCmd.Flags().GetBool("buildkite")

	if buildkite {
		log.Info("Building Authelia Go binaries with gox...")

		buildAutheliaBinaryGOX(buildMetaData.XFlags())
	} else {
		log.Info("Building Authelia Go binary...")

		buildAutheliaBinaryGO(buildMetaData.XFlags())
	}

	cleanAssets()
}

func buildAutheliaBinaryGOX(xflags []string) {
	var wg sync.WaitGroup

	s := time.Now()

	wg.Add(2)

	go func() {
		defer wg.Done()

		cmd := utils.CommandWithStdout("gox", "-output={{.Dir}}-{{.OS}}-{{.Arch}}-musl", "-buildmode=pie", "-trimpath", "-cgo", "-ldflags=-linkmode=external -s -w "+strings.Join(xflags, " "), "-osarch=linux/amd64 linux/arm linux/arm64", "./cmd/authelia/")

		cmd.Env = append(os.Environ(),
			"CGO_CPPFLAGS=-D_FORTIFY_SOURCE=2 -fstack-protector-strong", "CGO_LDFLAGS=-Wl,-z,relro,-z,now",
			"GOX_LINUX_ARM_CC=arm-linux-musleabihf-gcc", "GOX_LINUX_ARM64_CC=aarch64-linux-musl-gcc")

		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()

		cmd := utils.CommandWithStdout("bash", "-c", "docker run --rm -e GOX_LINUX_ARM_CC=arm-linux-gnueabihf-gcc -e GOX_LINUX_ARM64_CC=aarch64-linux-gnu-gcc -e GOX_FREEBSD_AMD64_CC=x86_64-pc-freebsd13-gcc -v ${PWD}:/workdir -v /buildkite/.go:/root/go authelia/crossbuild "+
			"gox -output={{.Dir}}-{{.OS}}-{{.Arch}} -buildmode=pie -trimpath -cgo -ldflags=\"-linkmode=external -s -w "+strings.Join(xflags, " ")+"\" -osarch=\"linux/amd64 linux/arm linux/arm64 freebsd/amd64\" ./cmd/authelia/")

		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()

	e := time.Since(s)

	log.Debugf("Binary compilation completed in %s.", e)
}

func buildAutheliaBinaryGO(xflags []string) {
	cmd := utils.CommandWithStdout("go", "build", "-buildmode=pie", "-trimpath", "-o", OutputDir+pathAuthelia, "-ldflags", "-linkmode=external -s -w "+strings.Join(xflags, " "), "./cmd/authelia/")

	cmd.Env = append(os.Environ(),
		"CGO_CPPFLAGS=-D_FORTIFY_SOURCE=2 -fstack-protector-strong", "CGO_LDFLAGS=-Wl,-z,relro,-z,now")

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func buildFrontend(branch string) {
	cmd := utils.CommandWithStdout("pnpm", "install")
	cmd.Dir = webDirectory

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	if !strings.HasPrefix(branch, "renovate/") {
		cmd = utils.CommandWithStdout("pnpm", "build")
		cmd.Dir = webDirectory

		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func buildSwagger() {
	cmd := utils.CommandWithStdout("bash", "-c", "wget -q https://github.com/swagger-api/swagger-ui/archive/v"+versionSwaggerUI+".tar.gz -O ./v"+versionSwaggerUI+extTarballGzip)

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("cp", "-r", "api", "internal/server/public_html")

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("tar", "-C", "internal/server/public_html/api", "--exclude=index.html", "--strip-components=2", "-xf", "v"+versionSwaggerUI+extTarballGzip, "swagger-ui-"+versionSwaggerUI+"/dist")

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("rm", "./v"+versionSwaggerUI+extTarballGzip)

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func cleanAssets() {
	if err := os.Rename("internal/server/public_html", OutputDir+"/public_html"); err != nil {
		log.Fatal(err)
	}

	cmd := utils.CommandWithStdout("mkdir", "-p", "internal/server/public_html/api")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("bash", "-c", "touch internal/server/public_html/{index.html,api/index.html,api/openapi.yml}")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
