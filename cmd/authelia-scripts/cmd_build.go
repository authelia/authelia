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
		panic(err)
	}
}

func buildFrontend() {
	// Install npm dependencies.
	cmd := utils.CommandWithStdout("yarn", "install")
	cmd.Dir = webDirectory

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	// Then build the frontend.
	cmd = utils.CommandWithStdout("yarn", "build")
	cmd.Dir = webDirectory

	cmd.Env = append(os.Environ(), "INLINE_RUNTIME_CHUNK=false")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	if err := os.Rename("web/build", "./public_html"); err != nil {
		log.Fatal(err)
	}
}

func generateEmbeddedAssets() {
	cmd := utils.CommandWithStdout("go", "get", "-u", "aletheia.icu/broccoli")

	err := cmd.Run()

	if err != nil {
		panic(err)
	}

	cmd = utils.CommandWithStdout("go", "generate", ".")
	cmd.Dir = "internal/server"

	err = cmd.Run()

	if err != nil {
		panic(err)
	}

	if err := os.Rename("./public_html", OutputDir+"/public_html"); err != nil {
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
		panic(err)
	}

	log.Debug("Building Authelia frontend...")
	buildFrontend()

	log.Debug("Building Authelia Go binary...")
	generateEmbeddedAssets()
	buildAutheliaBinary()
}
