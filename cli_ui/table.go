package cli_ui

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"strings"
)

// ShowTable Shows which tables are currently available
func ShowTable(tableHeader []string, tableBody [][]string, tableFooter []string, setBorder bool) {
	data := tableBody
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(tableHeader)
	if len(tableFooter) > 0 {
		table.SetFooter(tableFooter) // Add Footer
	}
	table.SetBorder(setBorder) // Set Border to false
	table.AppendBulk(data)     // Add Bulk Data
	table.Render()
}

// ShowRows Display the table on the console
// TODO refactor function
func ShowRows(tableHeader []string, tableBodyMatrix [][]string, tableFooter []string, setBorder bool) {
	builder := strings.Builder{}
	tableF := "\t%" + strconv.Itoa(columnMaxWidth(tableHeader)) + "s"
	for rowIndex, row := range tableBodyMatrix {
		builder.WriteString(fmt.Sprintf("\n*********** Row %d **********\n\n", rowIndex))
		for columnIndex, column := range row {
			builder.WriteString(fmt.Sprintf(tableF+":\t%s\n", tableHeader[columnIndex], column))
		}
	}
	fmt.Println(builder.String())
}

// The width of the widest column of several columns is for column width alignment
func columnMaxWidth(columns []string) int {
	maxWidth := 0
	for _, column := range columns {
		if len(column) > maxWidth {
			maxWidth = len(column)
		}
	}
	return maxWidth
}
