package commands

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
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
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}

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

	return runAccessControlCheck(cmd.OutOrStdout(), object, subject, results, ctx.config.AccessControl.DefaultPolicy, verbose)
}

func runAccessControlCheck(w io.Writer, object authorization.Object, subject authorization.Subject, results []authorization.RuleMatchResult, defaultPolicy string, verbose bool) (err error) {
	if len(results) == 0 {
		_, _ = fmt.Fprintf(w, "\nThe default policy '%s' will be applied to ALL requests as no rules are configured.\n\n", defaultPolicy)

		return nil
	}

	tw := tabwriter.NewWriter(w, 1, 1, 4, ' ', 0)

	accessControlCheckWriteOutput(tw, object, subject, results, defaultPolicy, verbose)

	return tw.Flush()
}

func accessControlCheckWriteOutput(w io.Writer, object authorization.Object, subject authorization.Subject, results []authorization.RuleMatchResult, defaultPolicy string, verbose bool) {
	accessControlCheckWriteObjectSubject(w, object, subject)

	_, _ = fmt.Fprintln(w, "  #\tDomain\tResource\tQuery\tMethod\tNetwork\tSubject")

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
			_, _ = fmt.Fprintf(w, "* %d\t%s\t%s\t%s\t%s\t%s\t%s\n", i+1, hitMissMay(result.MatchDomain), hitMissMay(result.MatchResources), hitMissMay(result.MatchQuery), hitMissMay(result.MatchMethods), hitMissMay(result.MatchNetworks), hitMissMay(result.MatchSubjects, result.MatchSubjectsExact))
		case result.IsPotentialMatch() && !result.Skipped:
			if potentialPos == 0 {
				potentialPos, potential = i+1, result
			}

			_, _ = fmt.Fprintf(w, "~ %d\t%s\t%s\t%s\t%s\t%s\t%s\n", i+1, hitMissMay(result.MatchDomain), hitMissMay(result.MatchResources), hitMissMay(result.MatchQuery), hitMissMay(result.MatchMethods), hitMissMay(result.MatchNetworks), hitMissMay(result.MatchSubjects, result.MatchSubjectsExact))
		default:
			_, _ = fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%s\t%s\t%s\n", i+1, hitMissMay(result.MatchDomain), hitMissMay(result.MatchResources), hitMissMay(result.MatchQuery), hitMissMay(result.MatchMethods), hitMissMay(result.MatchNetworks), hitMissMay(result.MatchSubjects, result.MatchSubjectsExact))
		}
	}

	switch {
	case appliedPos != 0 && (potentialPos == 0 || (potentialPos > appliedPos)):
		_, _ = fmt.Fprintf(w, "\nThe policy '%s' from rule #%d will be applied to this request.\n\n", applied.Rule.Policy, appliedPos)
	case potentialPos != 0 && appliedPos != 0:
		_, _ = fmt.Fprintf(w, "\nThe policy '%s' from rule #%d will potentially be applied to this request. If not policy '%s' from rule #%d will be.\n\n", potential.Rule.Policy, potentialPos, applied.Rule.Policy, appliedPos)
	case potentialPos != 0:
		_, _ = fmt.Fprintf(w, "\nThe policy '%s' from rule #%d will potentially be applied to this request. Otherwise the policy '%s' from the default policy will be.\n\n", potential.Rule.Policy, potentialPos, defaultPolicy)
	default:
		_, _ = fmt.Fprintf(w, "\nThe policy '%s' from the default policy will be applied to this request as no rules matched the request.\n\n", defaultPolicy)
	}
}

func accessControlCheckWriteObjectSubject(w io.Writer, object authorization.Object, subject authorization.Subject) {
	_, _ = fmt.Fprintf(w, "Performing policy check for request to '%s'", object.String())

	if object.Method != "" {
		_, _ = fmt.Fprintf(w, " method '%s'", object.Method)
	}

	if subject.Username != "" {
		_, _ = fmt.Fprintf(w, " username '%s'", subject.Username)
	}

	if len(subject.Groups) != 0 {
		_, _ = fmt.Fprintf(w, " groups '%s'", strings.Join(subject.Groups, ","))
	}

	if subject.IP != nil {
		_, _ = fmt.Fprintf(w, " from IP '%s'", subject.IP.String())
	}

	_, _ = fmt.Fprintf(w, ".\n\n")
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
