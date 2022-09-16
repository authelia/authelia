package main

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

var rootCmd *cobra.Command

func init() {
	rootCmd = newRootCmd()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseRoot,
		Short: "Authelia's generator tooling",
		RunE:  rootSubCommandsRunE,

		DisableAutoGenTag: true,
	}

	cmd.PersistentFlags().StringP(cmdFlagCwd, "C", "", "Sets the CWD for git commands")
	cmd.PersistentFlags().StringP(cmdFlagRoot, "d", dirCurrent, "The repository root")
	cmd.PersistentFlags().StringSliceP(cmdFlagExclude, "X", nil, "Sets the names of excluded generators")
	cmd.PersistentFlags().String(cmdFlagFeatureRequest, fileGitHubIssueTemplateFR, "Sets the path of the feature request issue template file")
	cmd.PersistentFlags().String(cmdFlagBugReport, fileGitHubIssueTemplateBR, "Sets the path of the bug report issue template file")
	cmd.PersistentFlags().Int(cmdFlagVersions, 5, "the maximum number of minor versions to list in output templates")
	cmd.PersistentFlags().String(cmdFlagDirLocales, dirLocales, "The locales directory in relation to the root")
	cmd.PersistentFlags().String(cmdFlagFileWebI18N, fileWebI18NIndex, "The i18n typescript configuration file in relation to the root")
	cmd.PersistentFlags().String(cmdFlagDocsDataLanguages, fileDocsDataLanguages, "The languages docs data file in relation to the docs data folder")
	cmd.PersistentFlags().String(cmdFlagDocsCLIReference, dirDocsCLIReference, "The directory to store the markdown in")
	cmd.PersistentFlags().String(cmdFlagDocsContent, dirDocsContent, "The directory with the docs content")
	cmd.PersistentFlags().String(cmdFlagFileConfigKeys, fileCodeConfigKeys, "Sets the path of the keys file")
	cmd.PersistentFlags().String(cmdFlagFileDocsKeys, fileDocsConfigKeys, "Sets the path of the docs keys file")
	cmd.PersistentFlags().String(cmdFlagPackageConfigKeys, pkgConfigSchema, "Sets the package name of the keys file")
	cmd.PersistentFlags().String(cmdFlagFileScriptsGen, fileScriptsGen, "Sets the path of the authelia-scripts gen file")
	cmd.PersistentFlags().String(cmdFlagPackageScriptsGen, pkgScriptsGen, "Sets the package name of the authelia-scripts gen file")
	cmd.PersistentFlags().String(cmdFlagFileConfigCommitLint, fileCICommitLintConfig, "The commit lint javascript configuration file in relation to the root")
	cmd.PersistentFlags().String(cmdFlagFileDocsCommitMsgGuidelines, fileDocsCommitMessageGuidelines, "The commit message guidelines documentation file in relation to the root")

	cmd.AddCommand(newCodeCmd(), newDocsCmd(), newGitHubCmd(), newLocalesCmd(), newCommitLintCmd())

	return cmd
}

func rootSubCommandsRunE(cmd *cobra.Command, args []string) (err error) {
	var exclude []string

	if exclude, err = cmd.Flags().GetStringSlice(cmdFlagExclude); err != nil {
		return err
	}

	subCmds := cmd.Commands()

	switch cmd.Use {
	case cmdUseRoot:
		sort.Slice(subCmds, func(i, j int) bool {
			switch subCmds[j].Use {
			case cmdUseDocs:
				// Ensure `docs` subCmd is last.
				return true
			default:
				return subCmds[i].Use < subCmds[j].Use
			}
		})
	case cmdUseDocs:
		sort.Slice(subCmds, func(i, j int) bool {
			switch subCmds[j].Use {
			case cmdUseDocsDate:
				// Ensure `date` subCmd is last.
				return true
			default:
				return subCmds[i].Use < subCmds[j].Use
			}
		})
	default:
		sort.Slice(subCmds, func(i, j int) bool {
			return subCmds[i].Use < subCmds[j].Use
		})
	}

	for _, subCmd := range subCmds {
		if subCmd.Use == cmdUseCompletion || strings.HasPrefix(subCmd.Use, "help ") || utils.IsStringInSlice(subCmd.Use, exclude) {
			continue
		}

		rootCmd.SetArgs(rootCmdGetArgs(subCmd, args))

		if err = rootCmd.Execute(); err != nil {
			return err
		}
	}

	return nil
}

func rootCmdGetArgs(cmd *cobra.Command, args []string) []string {
	for {
		if cmd == rootCmd {
			break
		}

		args = append([]string{cmd.Use}, args...)

		cmd = cmd.Parent()
	}

	return args
}
