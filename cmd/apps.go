package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newAppsCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "Inspect Google Play apps",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List apps visible to the active service account",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("apps list is not implemented yet; Google Play has limited app discovery APIs")
		},
	})

	return cmd
}
