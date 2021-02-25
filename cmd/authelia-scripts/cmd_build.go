package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/utils"
)

func buildAutheliaBinary() {
	cmd := utils.CommandWithStdout("go", "build", "-o", "../../"+OutputDir+"/authelia")
	cmd.Dir = "cmd/authelia"

	cmd.Env = append(os.Environ(),
		"GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=1")

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

	cmd = utils.CommandWithStdout("rm", "-rf", "internal/server/public_html")

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	err = os.Rename("web/build", "internal/server/public_html")
	if err != nil {
		log.Fatal(err)
	}
}

func buildSwagger() {
	swaggerVer := "3.43.0"
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

	log.Debug("Creating `" + OutputDir + "` directory")
	err := os.MkdirAll(OutputDir, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Building Authelia frontend...")
	buildFrontend()

	log.Debug("Building swagger-ui frontend...")
	buildSwagger()

	log.Debug("Building Authelia Go binary...")
	buildAutheliaBinary()
	cleanAssets()
}
