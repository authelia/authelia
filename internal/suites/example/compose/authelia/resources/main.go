package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	log.SetFlags(0)

	tools := []string{
		"github.com/cespare/reflex",
		"github.com/go-delve/delve/cmd/dlv",
	}

	for _, tool := range tools {
		log.Printf("Installing %s", tool)

		cmd := exec.Command("go", "install", tool)
		cmd.Env = os.Environ()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalf("failed to install %s: %v", tool, err)
		}
	}
}
