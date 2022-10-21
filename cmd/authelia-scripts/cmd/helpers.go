package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/authelia/authelia/v4/internal/utils"
)

func getBuild(branch, buildNumber, extra string) (b *Build, err error) {
	var out string

	b = &Build{
		Branch: branch,
		Extra:  extra,
	}

	if buildNumber != "" {
		if b.Number, err = strconv.Atoi(buildNumber); err != nil {
			return nil, fmt.Errorf("error parsing provided build number: %w", err)
		}
	}

	if b.Branch == "" {
		if out, _, err = utils.RunCommandAndReturnOutput("git rev-parse --abbrev-ref HEAD"); err != nil {
			return nil, fmt.Errorf("error getting branch with git rev-parse: %w", err)
		}

		if out == "" {
			b.Branch = "master"
		} else {
			b.Branch = out
		}
	}

	var (
		gitTagCommit string
	)

	if gitTagCommit, _, err = utils.RunCommandAndReturnOutput("git rev-list --tags --max-count=1"); err != nil {
		return nil, fmt.Errorf("error getting tag commit with git rev-list: %w", err)
	}

	if b.Tag, _, err = utils.RunCommandAndReturnOutput(fmt.Sprintf("git describe --tags --abbrev=0 %s", gitTagCommit)); err != nil {
		return nil, fmt.Errorf("error getting tag with git describe: %w", err)
	}

	if b.Commit, _, err = utils.RunCommandAndReturnOutput("git rev-parse HEAD"); err != nil {
		return nil, fmt.Errorf("error getting commit with git rev-parse: %w", err)
	}

	if gitTagCommit == b.Commit {
		b.Tagged = true
	}

	if _, exitCode, _ := utils.RunCommandAndReturnOutput("git diff --quiet"); exitCode == 0 {
		b.Clean = true
	}

	b.Date = time.Now()

	return b, nil
}
