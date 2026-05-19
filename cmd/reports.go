package cmd

import (
	"github.com/aljrico/Google-Play-Console-CLI/internal/gcsreports"
	"github.com/spf13/cobra"
)

func addReportDownloadFlags(cmd *cobra.Command, options *gcsreports.DownloadOptions) {
	cmd.Flags().StringVar(&options.Bucket, "bucket", "", "Google Play reports bucket, for example pubsite_prod_rev_0123456789")
	cmd.Flags().StringVar(&options.Object, "object", "", "Cloud Storage object path for the report")
	cmd.Flags().StringVar(&options.OutputPath, "file", "", "Destination .csv or .zip path")
	cmd.Flags().BoolVar(&options.Force, "force", false, "Overwrite the destination file")
	cmd.Flags().BoolVar(&options.DryRun, "dry-run", false, "Print the planned report download without calling Google Cloud Storage")
	_ = cmd.MarkFlagRequired("bucket")
	_ = cmd.MarkFlagRequired("object")
	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagFilename("file", "csv", "zip")
}
