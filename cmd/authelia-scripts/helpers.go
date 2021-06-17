package main

import (
	"fmt"
	"time"

	"github.com/authelia/authelia/internal/utils"
)

func getXFlags(arch, branch, build, extra string) (flags []string, err error) {
	if branch == "" {
		out, _, err := utils.RunCommandAndReturnOutput("git rev-parse --abbrev-ref HEAD")
		if err != nil {
			return flags, err
		}

		if out == "" {
			branch = "master"
		} else {
			branch = out
		}
	}

	gitTagCommit, _, err := utils.RunCommandAndReturnOutput("git rev-list --tags --max-count=1")
	if err != nil {
		return flags, err
	}

	tag, _, err := utils.RunCommandAndReturnOutput("git describe --tags --abbrev=0 " + gitTagCommit)
	if err != nil {
		return flags, err
	}

	commit, _, err := utils.RunCommandAndReturnOutput("git rev-parse HEAD")
	if err != nil {
		return flags, err
	}

	stateTag := "untagged"
	if gitTagCommit == commit {
		stateTag = "tagged"
	}

	if extra == "" {
		if _, exitCode, err := utils.RunCommandAndReturnOutput("git diff --quiet"); err != nil {
			return flags, err
		} else if exitCode != 0 {
			extra = "dirty"
		}
	}

	if build == "" {
		build = "manual"
	}

	return []string{
		fmt.Sprintf(fmtLDFLAGSX, "BuildBranch", branch),
		fmt.Sprintf(fmtLDFLAGSX, "BuildTag", tag),
		fmt.Sprintf(fmtLDFLAGSX, "BuildCommit", commit),
		fmt.Sprintf(fmtLDFLAGSX, "BuildDate", time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700")),
		fmt.Sprintf(fmtLDFLAGSX, "BuildStateTag", stateTag),
		fmt.Sprintf(fmtLDFLAGSX, "BuildStateExtra", extra),
		fmt.Sprintf(fmtLDFLAGSX, "BuildNumber", build),
		fmt.Sprintf(fmtLDFLAGSX, "BuildArch", arch),
	}, nil
}
