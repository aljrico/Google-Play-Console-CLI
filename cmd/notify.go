package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/notify"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

func newNotifyCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Send release workflow notifications",
	}
	cmd.AddCommand(newNotifySendCommand(out, options), newNotifyDiscordCommand(out, options), newNotifyGitHubCommand(out, options), newNotifySlackCommand(out, options), newNotifyTeamsCommand(out, options))
	return cmd
}

func newNotifySendCommand(out io.Writer, options *globalOptions) *cobra.Command {
	sendOptions := newNotifySendOptions("notify send")
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a JSON notification webhook",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := notify.Send(cmd.Context(), nil, sendOptions)
			if result.Webhook == "" {
				return err
			}
			if writeErr := output.Write(out, options.output, options.pretty, result); writeErr != nil {
				return writeErr
			}
			return err
		},
	}
	addNotifyFlags(cmd, &sendOptions, notifyFlagHelp{
		WebhookURL:     "HTTPS webhook URL; http is allowed only for loopback hosts",
		WebhookURLEnv:  "Environment variable containing the webhook URL",
		WebhookURLFile: "File containing the webhook URL",
		Confirm:        "Send the notification webhook",
		DryRun:         "Print the notification payload without sending",
	})
	return cmd
}

func newNotifySlackCommand(out io.Writer, options *globalOptions) *cobra.Command {
	sendOptions := newNotifySendOptions("notify slack")
	cmd := &cobra.Command{
		Use:   "slack",
		Short: "Send a Slack incoming webhook notification",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := notify.SendSlack(cmd.Context(), nil, sendOptions)
			if result.Webhook == "" {
				return err
			}
			if writeErr := output.Write(out, options.output, options.pretty, result); writeErr != nil {
				return writeErr
			}
			return err
		},
	}
	addNotifyFlags(cmd, &sendOptions, notifyFlagHelp{
		WebhookURL:     "HTTPS Slack incoming webhook URL; http is allowed only for loopback hosts",
		WebhookURLEnv:  "Environment variable containing the Slack incoming webhook URL",
		WebhookURLFile: "File containing the Slack incoming webhook URL",
		Confirm:        "Send the Slack webhook",
		DryRun:         "Print the Slack payload without sending",
	})
	return cmd
}

func newNotifyDiscordCommand(out io.Writer, options *globalOptions) *cobra.Command {
	sendOptions := newNotifySendOptions("notify discord")
	cmd := &cobra.Command{
		Use:   "discord",
		Short: "Send a Discord incoming webhook notification",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := notify.SendDiscord(cmd.Context(), nil, sendOptions)
			if result.Webhook == "" {
				return err
			}
			if writeErr := output.Write(out, options.output, options.pretty, result); writeErr != nil {
				return writeErr
			}
			return err
		},
	}
	addNotifyFlags(cmd, &sendOptions, notifyFlagHelp{
		WebhookURL:     "HTTPS Discord incoming webhook URL; http is allowed only for loopback hosts",
		WebhookURLEnv:  "Environment variable containing the Discord incoming webhook URL",
		WebhookURLFile: "File containing the Discord incoming webhook URL",
		Confirm:        "Send the Discord webhook",
		DryRun:         "Print the Discord payload without sending",
	})
	return cmd
}

func newNotifyGitHubCommand(out io.Writer, options *globalOptions) *cobra.Command {
	sendOptions := newNotifySendOptions("notify github")
	cmd := &cobra.Command{
		Use:   "github",
		Short: "Send a GitHub repository dispatch-shaped webhook notification",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := notify.SendGitHub(cmd.Context(), nil, sendOptions)
			if result.Webhook == "" {
				return err
			}
			if writeErr := output.Write(out, options.output, options.pretty, result); writeErr != nil {
				return writeErr
			}
			return err
		},
	}
	addNotifyFlags(cmd, &sendOptions, notifyFlagHelp{
		WebhookURL:     "HTTPS GitHub repository dispatch webhook URL; http is allowed only for loopback hosts",
		WebhookURLEnv:  "Environment variable containing the GitHub repository dispatch webhook URL",
		WebhookURLFile: "File containing the GitHub repository dispatch webhook URL",
		Confirm:        "Send the GitHub webhook",
		DryRun:         "Print the GitHub payload without sending",
	})
	cmd.Flags().StringVar(&sendOptions.EventType, "event-type", "gpc.notify", "GitHub repository dispatch event_type")
	return cmd
}

func newNotifyTeamsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	sendOptions := newNotifySendOptions("notify teams")
	cmd := &cobra.Command{
		Use:   "teams",
		Short: "Send a Microsoft Teams Workflows webhook notification",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := notify.SendTeams(cmd.Context(), nil, sendOptions)
			if result.Webhook == "" {
				return err
			}
			if writeErr := output.Write(out, options.output, options.pretty, result); writeErr != nil {
				return writeErr
			}
			return err
		},
	}
	addNotifyFlags(cmd, &sendOptions, notifyFlagHelp{
		WebhookURL:     "HTTPS Microsoft Teams Workflows webhook URL; legacy incoming connector URLs are also supported; http is allowed only for loopback hosts",
		WebhookURLEnv:  "Environment variable containing the Microsoft Teams Workflows webhook URL",
		WebhookURLFile: "File containing the Microsoft Teams Workflows webhook URL",
		Confirm:        "Send the Microsoft Teams webhook",
		DryRun:         "Print the Microsoft Teams payload without sending",
	})
	return cmd
}

type notifyFlagHelp struct {
	WebhookURL     string
	WebhookURLEnv  string
	WebhookURLFile string
	Confirm        string
	DryRun         string
}

func newNotifySendOptions(commandPath string) notify.SendOptions {
	return notify.SendOptions{
		CommandPath:   commandPath,
		WebhookURLEnv: notify.DefaultWebhookURLEnv,
	}
}

func addNotifyFlags(cmd *cobra.Command, sendOptions *notify.SendOptions, help notifyFlagHelp) {
	cmd.Flags().StringVar(&sendOptions.WebhookURL, "webhook-url", "", help.WebhookURL)
	cmd.Flags().StringVar(&sendOptions.WebhookURLEnv, "webhook-url-env", notify.DefaultWebhookURLEnv, help.WebhookURLEnv)
	cmd.Flags().StringVar(&sendOptions.WebhookURLFile, "webhook-url-file", "", help.WebhookURLFile)
	cmd.Flags().StringVar(&sendOptions.Title, "title", "", "Notification title")
	cmd.Flags().StringVar(&sendOptions.Message, "message", "", "Notification message")
	cmd.Flags().StringVar(&sendOptions.Severity, "severity", "", "Notification severity label")
	cmd.Flags().StringArrayVar(&sendOptions.Fields, "field", nil, "Notification field as name=value; repeatable")
	cmd.Flags().BoolVar(&sendOptions.Confirm, "confirm", false, help.Confirm)
	cmd.Flags().BoolVar(&sendOptions.DryRun, "dry-run", false, help.DryRun)
}
