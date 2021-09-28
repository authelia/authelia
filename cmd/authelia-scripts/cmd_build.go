package main

import (
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func buildAutheliaBinary(xflags []string, buildkite bool) {
	if buildkite {
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
	} else {
		cmd := utils.CommandWithStdout("go", "build", "-buildmode=pie", "-trimpath", "-o", OutputDir+"/authelia", "-ldflags", "-linkmode=external -s -w "+strings.Join(xflags, " "), "./cmd/authelia/")

		cmd.Env = append(os.Environ(),
			"CGO_CPPFLAGS=-D_FORTIFY_SOURCE=2 -fstack-protector-strong", "CGO_LDFLAGS=-Wl,-z,relro,-z,now")

		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func buildFrontend() {
	cmd := utils.CommandWithStdout("pnpm", "install", "--shamefully-hoist")
	cmd.Dir = webDirectory

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = utils.CommandWithStdout("pnpm", "build")
	cmd.Dir = webDirectory

	cmd.Env = append(os.Environ(), "GENERATE_SOURCEMAP=false", "INLINE_RUNTIME_CHUNK=false")

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func buildSwagger() {
	swaggerVer := "3.52.3"
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

	buildkite, _ := cobraCmd.Flags().GetBool("buildkite")

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

	if buildkite {
		log.Debug("Building Authelia Go binaries with gox...")
	} else {
		log.Debug("Building Authelia Go binary...")
	}

	buildAutheliaBinary(xflags, buildkite)

	cleanAssets()
}
