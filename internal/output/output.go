package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

type Format string

const (
	JSON     Format = "json"
	Table    Format = "table"
	Markdown Format = "markdown"
)

func (f *Format) String() string {
	if f == nil || *f == "" {
		return string(defaultFormat())
	}
	return string(*f)
}

func (f *Format) Set(value string) error {
	switch Format(value) {
	case JSON, Table, Markdown:
		*f = Format(value)
		return nil
	default:
		return fmt.Errorf("unsupported output format %q", value)
	}
}

func (f *Format) Type() string {
	return "format"
}

func Write(w io.Writer, format Format, pretty bool, value any) error {
	if format == "" {
		format = defaultFormat()
	}

	switch format {
	case JSON:
		enc := json.NewEncoder(w)
		if pretty {
			enc.SetIndent("", "  ")
		}
		return enc.Encode(value)
	case Table:
		return writeTable(w, value, false)
	case Markdown:
		return writeTable(w, value, true)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func defaultFormat() Format {
	if info, err := os.Stdout.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
		return Table
	}
	return JSON
}

func writeTable(w io.Writer, value any, markdown bool) error {
	rows, err := rowsFromValue(value)
	if err != nil {
		return err
	}

	if markdown {
		writeMarkdownTable(w, rows)
		return nil
	}
	writePlainTable(w, rows)
	return nil
}

type tableRows struct {
	header []string
	values [][]string
}

func rowsFromValue(value any) (tableRows, error) {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() == reflect.Map {
		header := []string{"key", "value"}
		values := make([][]string, 0, v.Len())
		for _, key := range v.MapKeys() {
			values = append(values, []string{fmt.Sprint(key.Interface()), fmt.Sprint(v.MapIndex(key).Interface())})
		}
		return tableRows{header: header, values: values}, nil
	}
	if v.Kind() == reflect.Struct {
		t := v.Type()
		header := make([]string, 0, v.NumField())
		values := make([]string, 0, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}
			header = append(header, headerName(field))
			values = append(values, fmt.Sprint(v.Field(i).Interface()))
		}
		return tableRows{header: header, values: [][]string{values}}, nil
	}
	return tableRows{}, fmt.Errorf("table output does not support %T yet", value)
}

func headerName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" || tag == "-" {
		return strings.ToLower(field.Name)
	}
	return strings.Split(tag, ",")[0]
}

func writePlainTable(w io.Writer, rows tableRows) {
	widths := columnWidths(rows)
	writePlainRow(w, rows.header, widths)
	for _, row := range rows.values {
		writePlainRow(w, row, widths)
	}
}

func writePlainRow(w io.Writer, row []string, widths []int) {
	for i, value := range row {
		if i > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprintf(w, "%-*s", widths[i], value)
	}
	fmt.Fprintln(w)
}

func writeMarkdownTable(w io.Writer, rows tableRows) {
	writeMarkdownRow(w, rows.header)
	separator := make([]string, len(rows.header))
	for i := range separator {
		separator[i] = "---"
	}
	writeMarkdownRow(w, separator)
	for _, row := range rows.values {
		writeMarkdownRow(w, row)
	}
}

func writeMarkdownRow(w io.Writer, row []string) {
	fmt.Fprintf(w, "| %s |\n", strings.Join(row, " | "))
}

func columnWidths(rows tableRows) []int {
	widths := make([]int, len(rows.header))
	for i, value := range rows.header {
		widths[i] = len(value)
	}
	for _, row := range rows.values {
		for i, value := range row {
			if len(value) > widths[i] {
				widths[i] = len(value)
			}
		}
	}
	return widths
}
