package cmd

import (
	"os"
	"strings"
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
		log.Info("Building Authelia Go binaries with GoReleaser...")

		buildAutheliaBinaryCI(buildMetaData.XFlags())
	} else {
		log.Info("Building Authelia Go binary...")

		buildAutheliaBinaryGO(buildMetaData.XFlags())
	}

	cleanAssets()
}

func buildAutheliaBinaryCI(xflags []string) {
	s := time.Now()

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	args := []string{
		"run", "--rm",
		"--name", "authelia-crossbuild",
		"--user", "1000:1000",
		"-e", "BUILDKITE_TAG=" + os.Getenv("BUILDKITE_TAG"),
		"-e", "GOPATH=/tmp/go",
		"-e", "GOCACHE=/tmp/go-build",
		"-e", "GPG_PASSWORD=" + os.Getenv("GPG_PASSWORD"),
		"-e", "GPG_KEY_PATH=" + os.Getenv("GPG_KEY_PATH"),
		"-e", "HOME=/tmp",
		"-e", "NFPM_DEBIAN_PASSPHRASE=" + os.Getenv("GPG_PASSWORD"),
		"-e", "XFLAGS=" + strings.Join(xflags, " "),
		"-v", pwd + ":/workdir",
		"-v", "/buildkite/.gnupg:/tmp/.gnupg",
		"-v", "/buildkite/.go:/tmp/go",
		"-v", "/buildkite/.sign:/tmp/sign",
		"-v", "/usr/lib/go:/usr/local/go",
		"-v", "/usr/local/include:/usr/local/include",
		"-v", "/usr/bin/goreleaser:/usr/local/bin/goreleaser",
		"-v", "/usr/local/bin/grype:/usr/local/bin/grype",
		"-v", "/usr/local/bin/syft:/usr/local/bin/syft",
		"authelia/crossbuild",
		"goreleaser", "release", "--skip=publish,validate",
	}

	cmd := utils.CommandWithStdout("docker", args...)

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("Binary compilation completed in %s.", time.Since(s))
}

func buildAutheliaBinaryGO(xflags []string) {
	cmd := utils.CommandWithStdout("go", "build", "-buildmode=pie", "-trimpath", "-o", OutputDir+pathAuthelia, "-ldflags", "-linkmode=external -s -w "+strings.Join(xflags, " "), "./cmd/authelia/")

	cmd.Env = append(os.Environ(),
		"GOEXPERIMENT=nosynchashtriemap", "CGO_CPPFLAGS=-D_FORTIFY_SOURCE=2 -fstack-protector-strong", "CGO_LDFLAGS=-Wl,-z,relro,-z,now")

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func buildFrontend(branch string) {
	cmd := utils.CommandWithStdout("pnpm", "install", "--ignore-scripts")
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

	cmd = utils.CommandWithStdout("tar", "-C", "internal/server/public_html/api", "--exclude=index.html", "--exclude=*.map", "--exclude=*-es-*", "--exclude=swagger-{ui,initializer}.js", "--strip-components=2", "-xf", "v"+versionSwaggerUI+extTarballGzip, "swagger-ui-"+versionSwaggerUI+"/dist")

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
