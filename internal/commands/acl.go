package commands

import (
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newAccessControlCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "access-control",
		Short: "Helpers for the access control system",
	}

	cmd.AddCommand(
		newAccessControlCheckCommand(),
	)

	return cmd
}

func newAccessControlCheckCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "check-policy",
		Short: "Checks a request against the access control rules to determine what policy would be applied",
		RunE:  accessControlCheckRunE,
	}

	cmdWithConfigFlags(cmd, false, []string{"config.yml"})

	cmd.Flags().String("url", "", "the url of the object")
	cmd.Flags().String("method", "GET", "the HTTP method of the object")
	cmd.Flags().String("username", "", "the username of the subject")
	cmd.Flags().StringSlice("groups", nil, "the groups of the subject")
	cmd.Flags().String("ip", "", "the ip of the subject")
	cmd.Flags().Bool("verbose", false, "enables verbose output")

	return cmd
}

func accessControlCheckRunE(cmd *cobra.Command, _ []string) (err error) {
	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	sources := make([]configuration.Source, len(configs)+2)

	for i, path := range configs {
		sources[i] = configuration.NewYAMLFileSource(path)
	}

	sources[0+len(configs)] = configuration.NewEnvironmentSource(configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter)
	sources[1+len(configs)] = configuration.NewSecretsSource(configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter)

	val := schema.NewStructValidator()

	accessControlConfig := &schema.Configuration{}

	_, err = configuration.LoadAdvanced(val, "access_control", &accessControlConfig.AccessControl, sources...)

	if err != nil {
		return err
	}

	authorizer := authorization.NewAuthorizer(accessControlConfig)

	subject, object, err := getSubjectAndObjectFromFlags(cmd)
	if err != nil {
		return err
	}

	results := authorizer.GetRuleMatchResults(subject, object)

	if len(results) == 0 {
		return errors.New("no rules to check")
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}

	fmt.Printf("  Pos\tDomain\tResource\tMethod\tNetwork\tSubject\n")

	var (
		appliedPos int
		applied    authorization.RuleMatchResult
	)

	for i, result := range results {
		if result.Skipped && !verbose {
			break
		}

		if result.IsMatch() && !result.Skipped {
			appliedPos, applied = i+1, result

			fmt.Printf("* %d\t%s\t%s\t\t%s\t%s\t%s\n", i+1, hitMiss(result.MatchDomain), hitMiss(result.MatchResources), hitMiss(result.MatchMethods), hitMiss(result.MatchNetworks), hitMiss(result.MatchSubjects))
		} else {
			fmt.Printf("  %d\t%s\t%s\t\t%s\t%s\t%s\n", i+1, hitMiss(result.MatchDomain), hitMiss(result.MatchResources), hitMiss(result.MatchMethods), hitMiss(result.MatchNetworks), hitMiss(result.MatchSubjects))
		}
	}

	fmt.Printf("\nRule %d with policy %s will be applied to this request.\n\n", appliedPos, authorization.LevelToPolicy(applied.Rule.Policy))

	return nil
}

func hitMiss(in bool) (out string) {
	if in {
		return "hit"
	}

	return "miss"
}

func getSubjectAndObjectFromFlags(cmd *cobra.Command) (subject authorization.Subject, object authorization.Object, err error) {
	requestURL, err := cmd.Flags().GetString("url")
	if err != nil {
		return subject, object, err
	}

	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return subject, object, err
	}

	method, err := cmd.Flags().GetString("method")
	if err != nil {
		return subject, object, err
	}

	username, err := cmd.Flags().GetString("username")
	if err != nil {
		return subject, object, err
	}

	groups, err := cmd.Flags().GetStringSlice("groups")
	if err != nil {
		return subject, object, err
	}

	remoteIP, err := cmd.Flags().GetString("ip")
	if err != nil {
		return subject, object, err
	}

	parsedIP := net.ParseIP(remoteIP)

	subject = authorization.Subject{
		Username: username,
		Groups:   groups,
		IP:       parsedIP,
	}

	object = authorization.NewObject(parsedURL, method)

	return subject, object, nil
}
