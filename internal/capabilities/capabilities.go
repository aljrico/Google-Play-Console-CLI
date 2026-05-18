package capabilities

import (
	"fmt"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/docs"
)

type Status string

const (
	StatusPlanned       Status = "planned"
	StatusImplemented   Status = "implemented"
	StatusTested        Status = "tested"
	StatusDocumented    Status = "documented"
	StatusBlocked       Status = "blocked"
	StatusNotApplicable Status = "not applicable"
)

type Capability struct {
	Section   string `json:"section"`
	ASCFamily string `json:"ascFamily"`
	GPCFamily string `json:"gpcFamily"`
	GoogleAPI string `json:"googleApiCoverage"`
	Status    Status `json:"status"`
	Notes     string `json:"notes"`
}

type ListOptions struct {
	Status  Status `json:"status,omitempty"`
	Section string `json:"section,omitempty"`
}

func (o ListOptions) Validate() error {
	if o.Status == "" {
		return nil
	}
	switch o.Status {
	case StatusPlanned, StatusImplemented, StatusTested, StatusDocumented, StatusBlocked, StatusNotApplicable:
		return nil
	default:
		return fmt.Errorf("unsupported capability status %q", o.Status)
	}
}

func List(options ListOptions) ([]Capability, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	capabilities, err := Parse(docs.ParityMatrix)
	if err != nil {
		return nil, err
	}
	return filter(capabilities, options), nil
}

func Parse(markdown string) ([]Capability, error) {
	var (
		section       string
		inParityTable bool
		capabilities  []Capability
	)

	for _, line := range strings.Split(markdown, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") {
			section = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			inParityTable = false
			continue
		}
		if strings.HasPrefix(line, "| `asc` family |") {
			inParityTable = true
			continue
		}
		if !inParityTable || line == "" || strings.HasPrefix(line, "| ---") {
			continue
		}
		if !strings.HasPrefix(line, "|") {
			inParityTable = false
			continue
		}
		capability, err := capabilityFromMarkdownRow(section, line)
		if err != nil {
			return nil, err
		}
		capabilities = append(capabilities, capability)
	}
	return capabilities, nil
}

func capabilityFromMarkdownRow(section string, row string) (Capability, error) {
	columns := splitMarkdownRow(row)
	if len(columns) != 5 {
		return Capability{}, fmt.Errorf("capability row has %d columns, want 5: %s", len(columns), row)
	}
	status := Status(stripCode(columns[3]))
	if err := (ListOptions{Status: status}).Validate(); err != nil {
		return Capability{}, err
	}
	return Capability{
		Section:   section,
		ASCFamily: stripCode(columns[0]),
		GPCFamily: stripCode(columns[1]),
		GoogleAPI: stripCode(columns[2]),
		Status:    status,
		Notes:     columns[4],
	}, nil
}

func splitMarkdownRow(row string) []string {
	trimmed := strings.Trim(row, "|")
	var (
		columns []string
		current strings.Builder
		escaped bool
	)
	for _, character := range trimmed {
		if escaped {
			if character == '|' {
				current.WriteRune('|')
			} else {
				current.WriteRune('\\')
				current.WriteRune(character)
			}
			escaped = false
			continue
		}
		if character == '\\' {
			escaped = true
			continue
		}
		if character == '|' {
			columns = append(columns, strings.TrimSpace(current.String()))
			current.Reset()
			continue
		}
		current.WriteRune(character)
	}
	if escaped {
		current.WriteRune('\\')
	}
	columns = append(columns, strings.TrimSpace(current.String()))
	return columns
}

func stripCode(value string) string {
	value = strings.TrimSpace(value)
	if value == "`N/A`" {
		return "N/A"
	}
	return strings.ReplaceAll(value, "`", "")
}

func filter(items []Capability, options ListOptions) []Capability {
	filtered := make([]Capability, 0, len(items))
	for _, item := range items {
		if options.Status != "" && item.Status != options.Status {
			continue
		}
		if options.Section != "" && !strings.EqualFold(item.Section, options.Section) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}
