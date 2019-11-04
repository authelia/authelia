package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func buildAutheliaBinary() {
	cmd := CommandWithStdout("go", "build", "-o", "../../"+OutputDir+"/authelia")
	cmd.Dir = "cmd/authelia"
	cmd.Env = append(os.Environ(),
		"GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")

	err := cmd.Run()

	if err != nil {
		panic(err)
	}
}

func buildFrontend() {
	cmd := CommandWithStdout("npm", "run", "build")
	cmd.Dir = "client"
	err := cmd.Run()

	if err != nil {
		panic(err)
	}

	err = os.Rename("client/build", OutputDir+"/public_html")

	if err != nil {
		panic(err)
	}
}

// Build build Authelia
func Build(cobraCmd *cobra.Command, args []string) {
	Clean(cobraCmd, args)

	fmt.Println("Creating `" + OutputDir + "` directory")
	err := os.MkdirAll(OutputDir, os.ModePerm)

	if err != nil {
		panic(err)
	}

	fmt.Println("Building Authelia Go binary...")
	buildAutheliaBinary()

	fmt.Println("Building Authelia frontend...")
	buildFrontend()
}
