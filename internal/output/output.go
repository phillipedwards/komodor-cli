package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatCSV   Format = "csv"
)

type Formatter interface {
	Print(data any) error
}

func New(format string, writers ...io.Writer) (Formatter, error) {
	w := io.Writer(os.Stdout)
	if len(writers) > 0 && writers[0] != nil {
		w = writers[0]
	}
	switch Format(strings.ToLower(format)) {
	case FormatTable, "":
		return &tableFormatter{w: w}, nil
	case FormatJSON:
		return &jsonFormatter{w: w}, nil
	case FormatYAML:
		return &yamlFormatter{w: w}, nil
	case FormatCSV:
		return &csvFormatter{w: w}, nil
	default:
		return nil, fmt.Errorf("unknown output format %q: use table, json, yaml, or csv", format)
	}
}

// TableData is implemented by types that know how to render themselves as a table.
type TableData interface {
	Headers() []string
	Rows() [][]string
}

type tableFormatter struct{ w io.Writer }

func (f *tableFormatter) Print(data any) error {
	if td, ok := data.(TableData); ok {
		tbl := tablewriter.NewTable(f.w,
			tablewriter.WithConfig(tablewriter.Config{
				Header: tw.CellConfig{
					Formatting: tw.CellFormatting{AutoFormat: tw.Off},
					Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
				},
				Row: tw.CellConfig{
					Alignment: tw.CellAlignment{Global: tw.AlignLeft},
				},
				Behavior: tw.Behavior{TrimSpace: tw.On},
			}),
			tablewriter.WithRendition(tw.Rendition{
				Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
				Settings: tw.Settings{
					Separators: tw.Separators{
						BetweenColumns: tw.On,
						BetweenRows:    tw.Off,
					},
				},
			}),
		)
		hdrs := td.Headers()
		hdrAny := make([]any, len(hdrs))
		for i, h := range hdrs {
			hdrAny[i] = h
		}
		tbl.Header(hdrAny...)
		for _, row := range td.Rows() {
			rowAny := make([]any, len(row))
			for i, v := range row {
				rowAny[i] = v
			}
			if err := tbl.Append(rowAny...); err != nil {
				return err
			}
		}
		return tbl.Render()
	}
	// Fallback: render as JSON for types without table support
	return (&jsonFormatter{w: f.w}).Print(data)
}

type jsonFormatter struct{ w io.Writer }

func (f *jsonFormatter) Print(data any) error {
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

type yamlFormatter struct{ w io.Writer }

func (f *yamlFormatter) Print(data any) error {
	return yaml.NewEncoder(f.w).Encode(data)
}

type csvFormatter struct{ w io.Writer }

func (f *csvFormatter) Print(data any) error {
	if td, ok := data.(TableData); ok {
		w := csv.NewWriter(f.w)
		if err := w.Write(td.Headers()); err != nil {
			return err
		}
		for _, row := range td.Rows() {
			if err := w.Write(row); err != nil {
				return err
			}
		}
		w.Flush()
		return w.Error()
	}
	return fmt.Errorf("csv output not supported for this command")
}

// Str safely dereferences a *string for table display.
func Str(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Bool formats a *bool for display.
func Bool(b *bool) string {
	if b == nil {
		return ""
	}
	if *b {
		return "true"
	}
	return "false"
}
