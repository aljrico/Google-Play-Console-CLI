package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/notifications"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

func newNotificationsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notifications",
		Short: "Inspect Google Play notification payloads",
	}
	cmd.AddCommand(newNotificationsPubSubCommand(out, options), newNotificationsRTDNCommand(out, options))
	return cmd
}

func newNotificationsPubSubCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pubsub",
		Short: "Set up Google Cloud Pub/Sub for Play notifications",
	}
	cmd.AddCommand(newNotificationsPubSubSetupCommand(out, options))
	return cmd
}

func newNotificationsPubSubSetupCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var setupOptions notifications.PubSubSetupOptions
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Create Pub/Sub resources for Play real-time developer notifications",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setupOptions.Validate(); err != nil {
				return err
			}
			if setupOptions.DryRun {
				result, err := notifications.SetupPubSub(cmd.Context(), nil, setupOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			client, err := notifications.NewPubSubClientFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := notifications.SetupPubSub(cmd.Context(), client, setupOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&setupOptions.ProjectID, "project", "", "Google Cloud project ID that owns the Pub/Sub resources")
	cmd.Flags().StringVar(&setupOptions.TopicID, "topic", "", "Pub/Sub topic ID to create")
	cmd.Flags().StringVar(&setupOptions.SubscriptionID, "subscription", "", "Pub/Sub subscription ID to create")
	cmd.Flags().StringVar(&setupOptions.PushEndpoint, "push-endpoint", "", "Optional HTTPS push endpoint; omit for pull subscriptions")
	cmd.Flags().Int64Var(&setupOptions.AckDeadlineSeconds, "ack-deadline", 10, "Subscription acknowledgement deadline in seconds")
	cmd.Flags().BoolVar(&setupOptions.Confirm, "confirm", false, "Create Pub/Sub resources and grant the Google Play publisher role")
	cmd.Flags().BoolVar(&setupOptions.DryRun, "dry-run", false, "Print the planned Pub/Sub setup without calling Google Cloud")
	return cmd
}

func newNotificationsRTDNCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rtdn",
		Short: "Inspect real-time developer notifications",
	}
	cmd.AddCommand(newNotificationsRTDNDecodeCommand(out, options))
	return cmd
}

func newNotificationsRTDNDecodeCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var decodeOptions notifications.RTDNDecodeOptions
	cmd := &cobra.Command{
		Use:   "decode",
		Short: "Decode a Pub/Sub RTDN push payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := notifications.DecodeRTDN(decodeOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&decodeOptions.File, "file", "", "RTDN JSON payload file; required unless --data is set")
	cmd.Flags().StringVar(&decodeOptions.Data, "data", "", "Inline RTDN JSON payload; required unless --file is set")
	cmd.Flags().BoolVar(&decodeOptions.Unwrapped, "unwrapped", false, "Decode an unwrapped push payload containing the developer notification directly")
	return cmd
}
