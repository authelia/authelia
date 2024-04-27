package commands

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
)

func newAccessControlCommand(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "access-control",
		Short:   cmdAutheliaAccessControlShort,
		Long:    cmdAutheliaAccessControlLong,
		Example: cmdAutheliaAccessControlExample,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newAccessControlCheckCommand(ctx),
	)

	return cmd
}

func newAccessControlCheckCommand(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "check-policy",
		Short:   cmdAutheliaAccessControlCheckPolicyShort,
		Long:    cmdAutheliaAccessControlCheckPolicyLong,
		Example: cmdAutheliaAccessControlCheckPolicyExample,
		PreRunE: ctx.ChainRunE(
			ctx.HelperConfigLoadRunE,
		),
		RunE: ctx.AccessControlCheckRunE,

		DisableAutoGenTag: true,
	}

	cmd.Flags().String("url", "", "the url of the object")
	cmd.Flags().String("method", fasthttp.MethodGet, "the HTTP method of the object")
	cmd.Flags().String("username", "", "the username of the subject")
	cmd.Flags().StringSlice("groups", nil, "the groups of the subject")
	cmd.Flags().String("ip", "", "the ip of the subject")
	cmd.Flags().Bool("verbose", false, "enables verbose output")

	return cmd
}

func (ctx *CmdCtx) AccessControlCheckRunE(cmd *cobra.Command, _ []string) (err error) {
	validator.ValidateAccessControl(ctx.config, ctx.cconfig.validator)

	if ctx.cconfig.validator.HasErrors() {
		return errors.New("failed to execute command due to errors in the configuration")
	}

	authorizer := authorization.NewAuthorizer(ctx.config)

	subject, object, err := getSubjectAndObjectFromFlags(cmd)
	if err != nil {
		return err
	}

	results := authorizer.GetRuleMatchResults(subject, object)

	if len(results) == 0 {
		fmt.Printf("\nThe default policy '%s' will be applied to ALL requests as no rules are configured.\n\n", ctx.config.AccessControl.DefaultPolicy)

		return nil
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}

	accessControlCheckWriteOutput(object, subject, results, ctx.config.AccessControl.DefaultPolicy, verbose)

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

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)

	_, _ = fmt.Fprintln(w, "  #\tDomain\tResource\tMethod\tNetwork\tSubject")

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

			_, _ = fmt.Fprintf(w, "* %d\t%s\t%s\t%s\t%s\t%s\n", i+1, hitMissMay(result.MatchDomain), hitMissMay(result.MatchResources), hitMissMay(result.MatchMethods), hitMissMay(result.MatchNetworks), hitMissMay(result.MatchSubjects, result.MatchSubjectsExact))
		case result.IsPotentialMatch() && !result.Skipped:
			if potentialPos == 0 {
				potentialPos, potential = i+1, result
			}

			_, _ = fmt.Fprintf(w, "~ %d\t%s\t%s\t%s\t%s\t%s\n", i+1, hitMissMay(result.MatchDomain), hitMissMay(result.MatchResources), hitMissMay(result.MatchMethods), hitMissMay(result.MatchNetworks), hitMissMay(result.MatchSubjects, result.MatchSubjectsExact))
		default:
			_, _ = fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%s\t%s\n", i+1, hitMissMay(result.MatchDomain), hitMissMay(result.MatchResources), hitMissMay(result.MatchMethods), hitMissMay(result.MatchNetworks), hitMissMay(result.MatchSubjects, result.MatchSubjectsExact))
		}
	}

	_ = w.Flush()

	switch {
	case appliedPos != 0 && (potentialPos == 0 || (potentialPos > appliedPos)):
		fmt.Printf("\nThe policy '%s' from rule #%d will be applied to this request.\n\n", applied.Rule.Policy, appliedPos)
	case potentialPos != 0 && appliedPos != 0:
		fmt.Printf("\nThe policy '%s' from rule #%d will potentially be applied to this request. If not policy '%s' from rule #%d will be.\n\n", potential.Rule.Policy, potentialPos, applied.Rule.Policy, appliedPos)
	case potentialPos != 0:
		fmt.Printf("\nThe policy '%s' from rule #%d will potentially be applied to this request. Otherwise the policy '%s' from the default policy will be.\n\n", potential.Rule.Policy, potentialPos, defaultPolicy)
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

	parsedURL, err := url.ParseRequestURI(requestURL)
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
