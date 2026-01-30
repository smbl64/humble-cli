package util

import (
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// TableBuilder wraps tablewriter for consistent formatting
type TableBuilder struct {
	table   *tablewriter.Table
	headers []string
}

// NewTable creates a new table with the given headers
func NewTable(headers []string) *TableBuilder {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("+")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")
	table.SetHeaderLine(true)
	table.SetBorder(false)
	table.SetTablePadding(" ")
	table.SetNoWhiteSpace(false)

	// Right-align columns that contain "Size" or "#" in the header
	columnAlignments := make([]int, len(headers))
	for i, header := range headers {
		if strings.Contains(header, "Size") || header == "#" {
			columnAlignments[i] = tablewriter.ALIGN_RIGHT
		} else {
			columnAlignments[i] = tablewriter.ALIGN_LEFT
		}
	}
	table.SetColumnAlignment(columnAlignments)

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
