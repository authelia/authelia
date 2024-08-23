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
	cmd.PersistentFlags().String(cmdFlagWeb, dirWeb, "The repository web directory in relation to the root directory")
	cmd.PersistentFlags().StringSliceP(cmdFlagExclude, "X", nil, "Sets the names of excluded generators")
	cmd.PersistentFlags().String(cmdFlagFeatureRequest, fileGitHubIssueTemplateFR, "Sets the path of the feature request issue template file")
	cmd.PersistentFlags().String(cmdFlagBugReport, fileGitHubIssueTemplateBR, "Sets the path of the bug report issue template file")
	cmd.PersistentFlags().Int(cmdFlagVersionCount, 5, "the maximum number of minor versions to list in output templates")
	cmd.PersistentFlags().String(cmdFlagDirLocales, dirLocales, "The locales directory in relation to the root")
	cmd.PersistentFlags().String(cmdFlagDirSchema, "internal/configuration/schema", "The schema directory in relation to the root")
	cmd.PersistentFlags().String(cmdFlagDirAuthentication, "internal/authentication", "The authentication directory in relation to the root")
	cmd.PersistentFlags().String(cmdFlagFileWebI18N, fileWebI18NIndex, "The i18n typescript configuration file in relation to the web directory")
	cmd.PersistentFlags().String(cmdFlagFileWebPackage, fileWebPackage, "The node package configuration file in relation to the web directory")
	cmd.PersistentFlags().String(cmdFlagDocsDataLanguages, fileDocsDataLanguages, "The languages docs data file in relation to the docs data folder")
	cmd.PersistentFlags().String(cmdFlagDocsDataMisc, fileDocsDataMisc, "The misc docs data file in relation to the docs data folder")
	cmd.PersistentFlags().String(cmdFlagDocsCLIReference, dirDocsCLIReference, "The directory to store the markdown in")
	cmd.PersistentFlags().String(cmdFlagDocs, dirDocs, "The directory with the docs")
	cmd.PersistentFlags().String(cmdFlagDocsContent, dirDocsContent, "The directory with the docs content")
	cmd.PersistentFlags().String(cmdFlagDocsStatic, dirDocsStatic, "The directory with the docs static files")
	cmd.PersistentFlags().String(cmdFlagDocsStaticJSONSchemas, dirDocsStaticJSONSchemas, "The directory with the docs static JSONSchema files")
	cmd.PersistentFlags().String(cmdFlagDocsData, dirDocsData, "The directory with the docs data")
	cmd.PersistentFlags().String(cmdFlagDocsADR, dirDocsADR, "The directory with the ADR data")
	cmd.PersistentFlags().String(cmdFlagFileConfigKeys, fileCodeConfigKeys, "Sets the path of the keys file")
	cmd.PersistentFlags().String(cmdFlagDocsDataKeys, fileDocsDataConfigKeys, "Sets the path of the docs keys file")
	cmd.PersistentFlags().String(cmdFlagPackageConfigKeys, pkgConfigSchema, "Sets the package name of the keys file")
	cmd.PersistentFlags().String(cmdFlagFileScriptsGen, fileScriptsGen, "Sets the path of the authelia-scripts gen file")
	cmd.PersistentFlags().String(cmdFlagDocsStaticJSONSchemaConfiguration, fileDocsStaticJSONSchemasConfiguration, "Sets the path of the configuration JSONSchema")
	cmd.PersistentFlags().String(cmdFlagDocsStaticJSONSchemaUserDatabase, fileDocsStaticJSONSchemasUserDatabase, "Sets the path of the user database JSONSchema")
	cmd.PersistentFlags().String(cmdFlagDocsStaticJSONSchemaExportsTOTP, fileDocsStaticJSONSchemasExportsTOTP, "Sets the path of the TOTP export JSONSchema")
	cmd.PersistentFlags().String(cmdFlagDocsStaticJSONSchemaExportsWebAuthn, fileDocsStaticJSONSchemasExportsWebAuthn, "Sets the path of the WebAuthn export JSONSchema")
	cmd.PersistentFlags().String(cmdFlagDocsStaticJSONSchemaExportsIdentifiers, fileDocsStaticJSONSchemasExportsIdentifiers, "Sets the path of the identifiers export JSONSchema")
	cmd.PersistentFlags().String(cmdFlagFileServerGenerated, fileServerGenerated, "Sets the path of the server generated file")
	cmd.PersistentFlags().String(cmdFlagPackageScriptsGen, pkgScriptsGen, "Sets the package name of the authelia-scripts gen file")
	cmd.PersistentFlags().String(cmdFlagFileConfigCommitLint, fileCICommitLintConfig, "The commit lint javascript configuration file in relation to the root")
	cmd.PersistentFlags().String(cmdFlagFileDocsCommitMsgGuidelines, fileDocsCommitMessageGuidelines, "The commit message guidelines documentation file in relation to the root")
	cmd.PersistentFlags().Bool("latest", false, "Enables latest functionality with several generators like the JSON Schema generator")
	cmd.PersistentFlags().Bool("next", false, "Enables next functionality with several generators like the JSON Schema generator")
	cmd.PersistentFlags().StringSlice(cmdFlagVersions, []string{}, "The versions to run the generator for, the special versions current and next are mutually exclusive")
	cmd.AddCommand(newCodeCmd(), newDocsCmd(), newGitHubCmd(), newLocalesCmd(), newCommitLintCmd())

	return cmd
}

func rootSubCommandsRunE(cmd *cobra.Command, args []string) (err error) {
	var exclude []string

	if exclude, err = cmd.Flags().GetStringSlice(cmdFlagExclude); err != nil {
		return err
	}

	subCmds := sortCmds(cmd)

	for _, subCmd := range subCmds {
		if subCmd.Use == cmdUseCompletion || strings.HasPrefix(subCmd.Use, "help ") || utils.IsStringSliceContainsAny([]string{resolveCmdName(subCmd), subCmd.Use}, exclude) {
			continue
		}

		if cmd.Use == cmdUseDocs && subCmd.Use == cmdUseManage {
			continue
		}

		rootCmd.SetArgs(rootCmdGetArgs(subCmd, args))

		if err = rootCmd.Execute(); err != nil {
			return err
		}
	}

	return nil
}

func sortCmds(cmd *cobra.Command) []*cobra.Command {
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

	return subCmds
}

func resolveCmdName(cmd *cobra.Command) string {
	parent := cmd.Parent()

	if parent != nil && parent.Use != cmd.Use && parent.Use != cmdUseRoot {
		return resolveCmdName(parent) + "." + cmd.Use
	}

	return cmd.Use
}

func rootCmdGetArgs(cmd *cobra.Command, args []string) []string {
	for {
		if cmd == nil || cmd == rootCmd {
			break
		}

		args = append([]string{cmd.Use}, args...)

		cmd = cmd.Parent()
	}

	return args
}
