package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type globalOptions struct {
	output output.Format
	pretty bool
}

func newRootCommand(out io.Writer) *cobra.Command {
	options := &globalOptions{}
	cmd := &cobra.Command{
		Use:           "gpc",
		Short:         "Fast, scriptable CLI for the Google Play Developer API",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().VarP(&options.output, "output", "o", "Output format: json, table, markdown")
	cmd.PersistentFlags().BoolVar(&options.pretty, "pretty", false, "Pretty-print JSON output")
	cmd.SetOut(out)
	cmd.SetErr(out)

	cmd.AddCommand(
		newVersionCommand(out, options),
		newAccountCommand(out, options),
		newAuthCommand(out, options),
		newAppsCommand(out),
		newAppRecoveryCommand(out, options),
		newCapabilitiesCommand(out, options),
		newDataSafetyCommand(out, options),
		newDiffCommand(out, options),
		newDocsCommand(out, options),
		newDeviceTierConfigsCommand(out, options),
		newFinanceCommand(out, options),
		newInitCommand(out, options),
		newInsightsCommand(out, options),
		newNotifyCommand(out, options),
		newNotificationsCommand(out, options),
		newStatusCommand(out, options),
		newTestersCommand(out, options),
		newTracksCommand(out, options),
		newValidateCommand(out, options),
		newVitalsCommand(out, options),
		newReleasesCommand(out, options),
		newPublishCommand(out, options),
		newListingsCommand(out, options),
		newMigrateCommand(out, options),
		newDetailsCommand(out, options),
		newReviewsCommand(out, options),
		newSchemaCommand(out, options),
		newSearchCommand(out, options),
		newSnitchCommand(out, options),
		newInternalSharingCommand(out, options),
		newGeneratedAPKsCommand(out, options),
		newSystemAPKsCommand(out, options),
		newGrantsCommand(out, options),
		newInAppProductsCommand(out, options),
		newImagesCommand(out, options),
		newInstallSkillsCommand(out, options),
		newOneTimeProductsCommand(out, options),
		newOneTimeProductOffersCommand(out, options),
		newSubscriptionsCommand(out, options),
		newSubscriptionOffersCommand(out, options),
		newPurchasesCommand(out, options),
		newOrdersCommand(out, options),
		newPricingCommand(out, options),
		newUsersCommand(out, options),
		newWorkflowCommand(out, options),
	)

	return cmd
}

func Execute() error {
	cmd := newRootCommand(os.Stdout)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
