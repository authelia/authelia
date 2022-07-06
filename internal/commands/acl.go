package commands

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
)

func newAccessControlCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "access-control",
		Short:   cmdAutheliaAccessControlShort,
		Long:    cmdAutheliaAccessControlLong,
		Example: cmdAutheliaAccessControlExample,
	}

	cmd.AddCommand(
		newAccessControlCheckCommand(),
	)

	return cmd
}

func newAccessControlCheckCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "check-policy",
		Short:   cmdAutheliaAccessControlCheckPolicyShort,
		Long:    cmdAutheliaAccessControlCheckPolicyLong,
		Example: cmdAutheliaAccessControlCheckPolicyExample,
		RunE:    accessControlCheckRunE,
	}

	cmdWithConfigFlags(cmd, false, []string{"configuration.yml"})

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

	if _, err = configuration.LoadAdvanced(val, "access_control", &accessControlConfig.AccessControl, sources...); err != nil {
		return err
	}

	v := schema.NewStructValidator()

	validator.ValidateAccessControl(accessControlConfig, v)

	if v.HasErrors() || v.HasWarnings() {
		return errors.New("your configuration has errors")
	}

	authorizer := authorization.NewAuthorizer(accessControlConfig)

	subject, object, err := getSubjectAndObjectFromFlags(cmd)
	if err != nil {
		return err
	}

	results := authorizer.GetRuleMatchResults(subject, object)

	if len(results) == 0 {
		fmt.Printf("\nThe default policy '%s' will be applied to ALL requests as no rules are configured.\n\n", accessControlConfig.AccessControl.DefaultPolicy)

		return nil
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}

	accessControlCheckWriteOutput(object, subject, results, accessControlConfig.AccessControl.DefaultPolicy, verbose)

	return nil
}

func accessControlCheckWriteObjectSubject(object authorization.Object, subject authorization.Subject) {
	output := strings.Builder{}

	output.WriteString(fmt.Sprintf("Performing policy check for request to '%s'", object.String()))

	if object.Method != "" {
		output.WriteString(fmt.Sprintf(" method '%s'", object.Method))
	}

	if subject.Username != "" {
		output.WriteString(fmt.Sprintf(" username '%s'", subject.Username))
	}

	if len(subject.Groups) != 0 {
		output.WriteString(fmt.Sprintf(" groups '%s'", strings.Join(subject.Groups, ",")))
	}

	if subject.IP != nil {
		output.WriteString(fmt.Sprintf(" from IP '%s'", subject.IP.String()))
	}

	output.WriteString(".\n")

	fmt.Println(output.String())
}

func accessControlCheckWriteOutput(object authorization.Object, subject authorization.Subject, results []authorization.RuleMatchResult, defaultPolicy string, verbose bool) {
	accessControlCheckWriteObjectSubject(object, subject)

	fmt.Printf("  #\tDomain\tResource\tMethod\tNetwork\tSubject\n")

	var (
		appliedPos int
		applied    authorization.RuleMatchResult

		potentialPos int
		potential    authorization.RuleMatchResult
	)

	for i, result := range results {
		if result.Skipped && !verbose {
			break
		}

		switch {
		case result.IsMatch() && !result.Skipped:
			appliedPos, applied = i+1, result

			fmt.Printf("* %d\t%s\t%s\t\t%s\t%s\t%s\n", i+1, hitMissMay(result.MatchDomain), hitMissMay(result.MatchResources), hitMissMay(result.MatchMethods), hitMissMay(result.MatchNetworks), hitMissMay(result.MatchSubjects, result.MatchSubjectsExact))
		case result.IsPotentialMatch() && !result.Skipped:
			if potentialPos == 0 {
				potentialPos, potential = i+1, result
			}

			fmt.Printf("~ %d\t%s\t%s\t\t%s\t%s\t%s\n", i+1, hitMissMay(result.MatchDomain), hitMissMay(result.MatchResources), hitMissMay(result.MatchMethods), hitMissMay(result.MatchNetworks), hitMissMay(result.MatchSubjects, result.MatchSubjectsExact))
		default:
			fmt.Printf("  %d\t%s\t%s\t\t%s\t%s\t%s\n", i+1, hitMissMay(result.MatchDomain), hitMissMay(result.MatchResources), hitMissMay(result.MatchMethods), hitMissMay(result.MatchNetworks), hitMissMay(result.MatchSubjects, result.MatchSubjectsExact))
		}
	}

	switch {
	case appliedPos != 0 && (potentialPos == 0 || (potentialPos > appliedPos)):
		fmt.Printf("\nThe policy '%s' from rule #%d will be applied to this request.\n\n", authorization.LevelToString(applied.Rule.Policy), appliedPos)
	case potentialPos != 0 && appliedPos != 0:
		fmt.Printf("\nThe policy '%s' from rule #%d will potentially be applied to this request. If not policy '%s' from rule #%d will be.\n\n", authorization.LevelToString(potential.Rule.Policy), potentialPos, authorization.LevelToString(applied.Rule.Policy), appliedPos)
	case potentialPos != 0:
		fmt.Printf("\nThe policy '%s' from rule #%d will potentially be applied to this request. Otherwise the policy '%s' from the default policy will be.\n\n", authorization.LevelToString(potential.Rule.Policy), potentialPos, defaultPolicy)
	default:
		fmt.Printf("\nThe policy '%s' from the default policy will be applied to this request as no rules matched the request.\n\n", defaultPolicy)
	}
}

func hitMissMay(in ...bool) (out string) {
	var hit, miss bool

	for _, x := range in {
		if x {
			hit = true
		} else {
			miss = true
		}
	}

	switch {
	case hit && miss:
		return "may"
	case hit:
		return "hit"
	default:
		return "miss"
	}
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
