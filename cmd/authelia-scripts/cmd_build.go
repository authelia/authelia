package main

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func buildAutheliaBinary(xflags []string, buildkite bool) {
	cmd := utils.CommandWithStdout("go", "build", "-tags", "netgo", "-trimpath", "-o", OutputDir+"/authelia", "-ldflags", "-s -w "+strings.Join(xflags, " "), "./cmd/authelia/")

	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0")

	if buildkite {
		cmd = utils.CommandWithStdout("gox", "-tags=netgo", "-output={{.Dir}}-{{.OS}}-{{.Arch}}", "-ldflags=-s -w "+strings.Join(xflags, " "), "-osarch=linux/amd64 linux/arm linux/arm64 freebsd/amd64", "./cmd/authelia/")

		cmd.Env = append(os.Environ(),
			"GOFLAGS=-trimpath", "CGO_ENABLED=0")
	}

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func buildFrontend() {
	cmd := utils.CommandWithStdout("yarn", "install")
	cmd.Dir = webDirectory

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("yarn", "build")
	cmd.Dir = webDirectory

	cmd.Env = append(os.Environ(), "INLINE_RUNTIME_CHUNK=false")

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func buildSwagger() {
	swaggerVer := "3.52.2"
	cmd := utils.CommandWithStdout("bash", "-c", "wget -q https://github.com/swagger-api/swagger-ui/archive/v"+swaggerVer+".tar.gz -O ./v"+swaggerVer+".tar.gz")

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("cp", "-r", "api", "internal/server/public_html")

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("tar", "-C", "internal/server/public_html/api", "--exclude=index.html", "--strip-components=2", "-xf", "v"+swaggerVer+".tar.gz", "swagger-ui-"+swaggerVer+"/dist")

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("rm", "./v"+swaggerVer+".tar.gz")

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

// Build build Authelia.
func Build(cobraCmd *cobra.Command, args []string) {
	log.Info("Building Authelia...")

	Clean(cobraCmd, args)

	xflags, err := getXFlags(os.Getenv("BUILDKITE_BRANCH"), os.Getenv("BUILDKITE_BUILD_NUMBER"), "")
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Creating `" + OutputDir + "` directory")
	err = os.MkdirAll(OutputDir, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Building Authelia frontend...")
	buildFrontend()

	log.Debug("Building swagger-ui frontend...")
	buildSwagger()

	buildkite, _ := cobraCmd.Flags().GetBool("buildkite")
	if buildkite {
		log.Debug("Building Authelia Go binaries with gox...")
	} else {
		log.Debug("Building Authelia Go binary...")
	}

	buildAutheliaBinary(xflags, buildkite)

	cleanAssets()
}
