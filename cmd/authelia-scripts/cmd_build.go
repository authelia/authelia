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

	err = os.Rename("web/build", "./public_html")
	if err != nil {
		log.Fatal(err)
	}
}

func buildSwagger() {
	swaggerVer := "3.38.0"
	cmd := utils.CommandWithStdout("bash", "-c", "wget -q https://github.com/swagger-api/swagger-ui/archive/v"+swaggerVer+".tar.gz -O ./v"+swaggerVer+".tar.gz")

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(swaggerDirectory, 0775)
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("tar", "-C", swaggerDirectory, "--exclude=index.html", "--strip-components=2", "-xf", "v"+swaggerVer+".tar.gz", "swagger-ui-"+swaggerVer+"/dist")

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("rm", "./v"+swaggerVer+".tar.gz")

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("cp", "swagger/index.html", swaggerDirectory)

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("cp", "swagger/authelia-api.yml", swaggerDirectory)

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func generateEmbeddedAssets() {
	cmd := utils.CommandWithStdout("go", "get", "-u", "aletheia.icu/broccoli")

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("go", "generate", ".")
	cmd.Dir = "internal/configuration"

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("go", "generate", ".")
	cmd.Dir = "internal/server"

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	err = os.Rename("./public_html", OutputDir+"/public_html")
	if err != nil {
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
	generateEmbeddedAssets()
	buildAutheliaBinary()
}
