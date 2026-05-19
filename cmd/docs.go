package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/docs"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type docsDocument struct {
	Name    string `json:"name"`
	Format  string `json:"format"`
	Content string `json:"content"`
}

type commandReference struct {
	Name     string                  `json:"name"`
	Path     string                  `json:"path"`
	Use      string                  `json:"use"`
	Short    string                  `json:"short,omitempty"`
	Flags    []commandReferenceFlag  `json:"flags,omitempty"`
	Commands []commandReferenceEntry `json:"commands"`
}

type commandReferenceEntry struct {
	Name     string                  `json:"name"`
	Path     string                  `json:"path"`
	Use      string                  `json:"use"`
	Short    string                  `json:"short,omitempty"`
	Flags    []commandReferenceFlag  `json:"flags,omitempty"`
	Commands []commandReferenceEntry `json:"commands,omitempty"`
}

type commandReferenceFlag struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Usage     string `json:"usage,omitempty"`
	Default   string `json:"default,omitempty"`
}

func newDocsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Print embedded playpub documentation",
	}
	cmd.AddCommand(
		newDocsParityCommand(out, options),
		newDocsCommandsCommand(out, options),
	)
	return cmd
}

func newDocsParityCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parity",
		Short: "Print the asc-to-playpub parity matrix",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.output == output.Markdown {
				_, err := fmt.Fprint(out, docs.ParityMatrix)
				return err
			}
			return output.Write(out, options.output, options.pretty, docsDocument{
				Name:    "parity",
				Format:  "markdown",
				Content: docs.ParityMatrix,
			})
		},
	}
	return cmd
}

func newDocsCommandsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commands",
		Short: "Print generated command reference",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			reference := buildCommandReference(cmd.Root())
			if options.output == output.Markdown {
				_, err := fmt.Fprint(out, renderCommandReferenceMarkdown(reference))
				return err
			}
			return output.Write(out, options.output, options.pretty, reference)
		},
	}
	return cmd
}

func buildCommandReference(root *cobra.Command) commandReference {
	return commandReference{
		Name:     root.Name(),
		Path:     root.CommandPath(),
		Use:      root.UseLine(),
		Short:    root.Short,
		Flags:    commandFlags(root),
		Commands: childCommandReferences(root),
	}
}

func renderCommandReferenceMarkdown(reference commandReference) string {
	var builder strings.Builder
	builder.WriteString("# Command Reference\n\n")
	builder.WriteString("## ")
	builder.WriteString(reference.Path)
	builder.WriteString("\n\n")
	if reference.Short != "" {
		builder.WriteString(reference.Short)
		builder.WriteString("\n\n")
	}
	builder.WriteString("```sh\n")
	builder.WriteString(reference.Use)
	builder.WriteString("\n```\n\n")
	if len(reference.Flags) > 0 {
		builder.WriteString("### Global Flags\n\n")
		for _, flag := range reference.Flags {
			builder.WriteString("- `--")
			builder.WriteString(flag.Name)
			builder.WriteString("`")
			if flag.Shorthand != "" {
				builder.WriteString(" / `-")
				builder.WriteString(flag.Shorthand)
				builder.WriteString("`")
			}
			if flag.Usage != "" {
				builder.WriteString(": ")
				builder.WriteString(flag.Usage)
			}
			if flag.Default != "" {
				builder.WriteString(" (default `")
				builder.WriteString(flag.Default)
				builder.WriteString("`)")
			}
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}
	renderCommandReferenceEntries(&builder, reference.Commands, 3)
	return builder.String()
}

func renderCommandReferenceEntries(builder *strings.Builder, entries []commandReferenceEntry, level int) {
	if len(entries) == 0 {
		return
	}
	builder.WriteString(strings.Repeat("#", level))
	builder.WriteString(" Commands\n\n")
	for _, entry := range entries {
		builder.WriteString("- `")
		builder.WriteString(entry.Path)
		builder.WriteString("`: ")
		builder.WriteString(entry.Short)
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
	for _, entry := range entries {
		if len(entry.Commands) == 0 && len(entry.Flags) == 0 {
			continue
		}
		builder.WriteString(strings.Repeat("#", level))
		builder.WriteString(" ")
		builder.WriteString(entry.Path)
		builder.WriteString("\n\n")
		if entry.Short != "" {
			builder.WriteString(entry.Short)
			builder.WriteString("\n\n")
		}
		builder.WriteString("```sh\n")
		builder.WriteString(entry.Use)
		builder.WriteString("\n```\n\n")
		renderCommandFlags(builder, entry.Flags, "Flags", level+1)
		renderCommandReferenceEntries(builder, entry.Commands, level+1)
	}
}

func renderCommandFlags(builder *strings.Builder, flags []commandReferenceFlag, title string, level int) {
	if len(flags) == 0 {
		return
	}
	builder.WriteString(strings.Repeat("#", min(level, 6)))
	builder.WriteString(" ")
	builder.WriteString(title)
	builder.WriteString("\n\n")
	for _, flag := range flags {
		builder.WriteString("- `--")
		builder.WriteString(flag.Name)
		builder.WriteString("`")
		if flag.Shorthand != "" {
			builder.WriteString(" / `-")
			builder.WriteString(flag.Shorthand)
			builder.WriteString("`")
		}
		if flag.Usage != "" {
			builder.WriteString(": ")
			builder.WriteString(flag.Usage)
		}
		if flag.Default != "" {
			builder.WriteString(" (default `")
			builder.WriteString(flag.Default)
			builder.WriteString("`)")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
}

func childCommandReferences(parent *cobra.Command) []commandReferenceEntry {
	entries := make([]commandReferenceEntry, 0, len(parent.Commands()))
	for _, child := range sortedCommands(parent.Commands()) {
		if !shouldDocumentCommand(child) {
			continue
		}
		entries = append(entries, commandReferenceEntry{
			Name:     child.Name(),
			Path:     child.CommandPath(),
			Use:      child.UseLine(),
			Short:    child.Short,
			Flags:    commandFlags(child),
			Commands: childCommandReferences(child),
		})
	}
	return entries
}

func sortedCommands(commands []*cobra.Command) []*cobra.Command {
	sorted := make([]*cobra.Command, 0, len(commands))
	sorted = append(sorted, commands...)
	sort.Slice(sorted, func(i int, j int) bool {
		return sorted[i].Name() < sorted[j].Name()
	})
	return sorted
}

func shouldDocumentCommand(cmd *cobra.Command) bool {
	return cmd.Name() != "help" && !cmd.Hidden
}

func commandFlags(cmd *cobra.Command) []commandReferenceFlag {
	flags := []commandReferenceFlag{}
	cmd.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if !shouldDocumentFlag(flag) {
			return
		}
		flags = append(flags, commandFlag(flag))
	})
	return flags
}

func shouldDocumentFlag(flag *pflag.Flag) bool {
	return flag.Name != "help" && !flag.Hidden
}

func commandFlag(flag *pflag.Flag) commandReferenceFlag {
	return commandReferenceFlag{
		Name:      flag.Name,
		Shorthand: flag.Shorthand,
		Usage:     flag.Usage,
		Default:   flag.DefValue,
	}
}
