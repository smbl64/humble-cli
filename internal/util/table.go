package util

import (
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// TableBuilder wraps tablewriter for consistent formatting
type TableBuilder struct {
	table   *tablewriter.Table
	headers []string
}

// NewTable creates a new table with the given headers
func NewTable(headers []string) *TableBuilder {
	// Right-align columns that contain "Size" or "#" in the header
	columnAlignments := make([]tw.Align, len(headers))
	for i, header := range headers {
		if strings.Contains(header, "Size") || header == "#" {
			columnAlignments[i] = tw.AlignRight
		} else {
			columnAlignments[i] = tw.AlignLeft
		}
	}

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeaderAutoFormat(tw.On),
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithRowAlignment(tw.AlignLeft),
		tablewriter.WithRowAlignmentConfig(tw.CellAlignment{
			PerColumn: columnAlignments,
		}),
	)

	table.Header(headers)

	return &TableBuilder{
		table:   table,
		headers: headers,
	}
}

// AddRow adds a row to the table
func (tb *TableBuilder) AddRow(row []string) {
	tb.table.Append(row)
}

// Render renders the table to stdout
func (tb *TableBuilder) Render() {
	tb.table.Render()
}
