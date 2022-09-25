package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newXFlagsCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "xflags",
		Short:   cmdXFlagsShort,
		Long:    cmdXFlagsLong,
		Example: cmdXFlagsExample,
		Args:    cobra.NoArgs,
		Run:     cmdXFlagsRun,

		DisableAutoGenTag: true,
	}

	cmd.Flags().StringP("build", "b", "0", "Sets the BuildNumber flag value")
	cmd.Flags().StringP("extra", "e", "", "Sets the BuildExtra flag value")

	return cmd
}

func cmdXFlagsRun(cobraCmd *cobra.Command, _ []string) {
	build, err := cobraCmd.Flags().GetString("build")
	if err != nil {
		log.Fatal(err)
	}

	extra, err := cobraCmd.Flags().GetString("extra")
	if err != nil {
		log.Fatal(err)
	}

	buildMetaData, err := getBuild("", build, extra)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(strings.Join(buildMetaData.XFlags(), " "))
}
