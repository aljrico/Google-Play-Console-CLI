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
	cmd.AddCommand(newNotificationsRTDNCommand(out, options))
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
