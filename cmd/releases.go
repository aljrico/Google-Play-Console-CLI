package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newReleasesCommand(out io.Writer) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "releases",
		Short: "Upload and manage Google Play releases",
	}

	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(&cobra.Command{
		Use:   "upload",
		Short: "Upload an Android App Bundle to a track",
		RunE: func(cmd *cobra.Command, args []string) error {
			if packageName == "" {
				return fmt.Errorf("--package is required")
			}
			return fmt.Errorf("releases upload is not implemented yet")
		},
	})

	return cmd
}
