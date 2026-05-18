package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newTracksCommand(out io.Writer) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "tracks",
		Short: "Manage Google Play release tracks",
	}

	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List release tracks",
		RunE: func(cmd *cobra.Command, args []string) error {
			if packageName == "" {
				return fmt.Errorf("--package is required")
			}
			return fmt.Errorf("tracks list is not implemented yet")
		},
	})

	return cmd
}
