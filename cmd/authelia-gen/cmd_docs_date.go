package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v3"
)

func newDocsDateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseDocsDate,
		Short: "Generate doc dates",
		RunE:  docsDateRunE,

		DisableAutoGenTag: true,
	}

	cmd.Flags().String("commit-until", fasthttp.MethodHead, "The commit to check the logs until")
	cmd.Flags().String("commit-since", "", "The commit to check the logs since")

	return cmd
}

func docsDateRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		pathDocsContent, cwd, commitUtil, commitSince, commitFilter string
	)

	if pathDocsContent, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsContent); err != nil {
		return err
	}

	if cwd, err = cmd.Flags().GetString(cmdFlagCwd); err != nil {
		return err
	}

	if cmd.Flags().Changed("commit-since") {
		if commitUtil, err = cmd.Flags().GetString("commit-util"); err != nil {
			return err
		}

		if commitSince, err = cmd.Flags().GetString("commit-since"); err != nil {
			return err
		}

		commitFilter = fmt.Sprintf("%s...%s", commitUtil, commitSince)
	}

	return filepath.Walk(pathDocsContent, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return nil
		}

		frontmatterBytes := getFrontmatter(abs)

		if frontmatterBytes == nil {
			return nil
		}

		frontmatter := map[string]any{}

		if err = yaml.Unmarshal(frontmatterBytes, frontmatter); err != nil {
			return err
		}

		var (
			date time.Time
		)

		if value, ok := frontmatter["date"]; ok {
			date, ok = value.(time.Time)

			if !ok {
				var abspath string

				if abspath, err = filepath.Abs(path); err != nil {
					abspath = path
				}

				return fmt.Errorf("frontmatter for %s has an invalid date value: is %T with a value of %s", abspath, value, value)
			}
		}

		dateGit := getDateFromGit(cwd, abs, commitFilter)

		replaceDates(abs, date, dateGit)

		return nil
	})
}

var newline = []byte("\n")

func getDateFromGit(cwd, path, commitFilter string) *time.Time {
	var args []string

	if len(cwd) != 0 {
		args = append(args, "-C", cwd)
	}

	args = append(args, "log")

	if len(commitFilter) != 0 {
		args = append(args, commitFilter)
	}

	args = append(args, "-1", "--follow", "--diff-filter=A", "--pretty=format:%cD", "--", path)

	return getTimeFromGitCmd(exec.Command("git", args...))
}

func getTimeFromGitCmd(cmd *exec.Cmd) *time.Time {
	var (
		output []byte
		err    error
		t      time.Time
	)

	if output, err = cmd.Output(); err != nil {
		return nil
	}

	if t, err = time.Parse(dateFmtRFC2822, string(output)); err != nil {
		return nil
	}

	return &t
}

func replaceDates(path string, date time.Time, dateGit *time.Time) {
	f, err := os.Open(path)
	if err != nil {
		return
	}

	buf := bytes.Buffer{}

	scanner := bufio.NewScanner(f)

	var dateGitLine string

	dateLine := fmt.Sprintf("date: %s", date.Format(dateFmtYAML))

	if dateGit != nil {
		dateGitLine = fmt.Sprintf("date: %s", dateGit.Format(dateFmtYAML))
	} else {
		dateGitLine = dateLine
	}

	found := 0

	frontmatter := 0

	for scanner.Scan() {
		if found < 2 && frontmatter < 2 {
			switch {
			case scanner.Text() == delimiterLineFrontMatter:
				buf.Write(scanner.Bytes())

				frontmatter++
			case frontmatter != 0 && strings.HasPrefix(scanner.Text(), "date: "):
				buf.WriteString(dateGitLine)

				found++
			default:
				buf.Write(scanner.Bytes())
			}
		} else {
			buf.Write(scanner.Bytes())
		}

		buf.Write(newline)
	}

	f.Close()

	newF, err := os.Create(path)
	if err != nil {
		return
	}

	_, _ = buf.WriteTo(newF)

	newF.Close()
}

func getFrontmatter(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	var start bool

	buf := bytes.Buffer{}

	for scanner.Scan() {
		if start {
			if scanner.Text() == delimiterLineFrontMatter {
				break
			}

			buf.Write(scanner.Bytes())
			buf.Write(newline)
		} else if scanner.Text() == delimiterLineFrontMatter {
			start = true
		}
	}

	return buf.Bytes()
}
