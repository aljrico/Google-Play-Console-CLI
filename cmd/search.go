package cmd

import (
	"io"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/commandsearch"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newSearchCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "search QUERY...",
		Short: "Search playpub commands and flags",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := commandsearch.Search(commandSearchDocuments(cmd.Root()), commandsearch.Options{
				Query: strings.Join(args, " "),
				Limit: limit,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", commandsearch.DefaultLimit, "Maximum number of matches; 0 returns all matches")
	return cmd
}

func commandSearchDocuments(root *cobra.Command) []commandsearch.Document {
	documents := []commandsearch.Document{}
	collectCommandSearchDocuments(root, &documents)
	return documents
}

func collectCommandSearchDocuments(cmd *cobra.Command, documents *[]commandsearch.Document) {
	if !shouldDocumentCommand(cmd) {
		return
	}
	*documents = append(*documents, commandsearch.Document{
		Path:  cmd.CommandPath(),
		Use:   cmd.UseLine(),
		Short: cmd.Short,
		Flags: commandSearchFlagNames(cmd),
	})
	for _, child := range sortedCommands(cmd.Commands()) {
		collectCommandSearchDocuments(child, documents)
	}
}

func commandSearchFlagNames(cmd *cobra.Command) []string {
	flags := []string{}
	seen := map[string]bool{}
	visitCommandSearchFlags(cmd.NonInheritedFlags(), seen, &flags)
	visitCommandSearchFlags(cmd.PersistentFlags(), seen, &flags)
	return flags
}

func visitCommandSearchFlags(flagSet *pflag.FlagSet, seen map[string]bool, flags *[]string) {
	flagSet.VisitAll(func(flag *pflag.Flag) {
		if shouldDocumentFlag(flag) {
			appendCommandSearchFlag(flags, seen, flag.Name)
			appendCommandSearchFlag(flags, seen, "--"+flag.Name)
			if flag.Shorthand != "" {
				appendCommandSearchFlag(flags, seen, flag.Shorthand)
				appendCommandSearchFlag(flags, seen, "-"+flag.Shorthand)
			}
			if flag.Usage != "" {
				appendCommandSearchFlag(flags, seen, strings.ToLower(flag.Usage))
			}
		}
	})
}

func appendCommandSearchFlag(flags *[]string, seen map[string]bool, value string) {
	if seen[value] {
		return
	}
	seen[value] = true
	*flags = append(*flags, value)
}
