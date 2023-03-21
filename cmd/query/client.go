package query

import (
	"context"
	"github.com/c-bata/go-prompt"
	"github.com/selefra/selefra-provider-sdk/storage"
	"github.com/selefra/selefra-provider-sdk/storage_factory"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/pkg/utils"
	"os"
	"strings"
)

// ------------------------------------------------- --------------------------------------------------------------------

// SQLQueryClient TODO Optimize the experience of writing SQL statements
type SQLQueryClient struct {
	storageType storage_factory.StorageType
	Storage     storage.Storage

	Tables  []prompt.Suggest
	Columns []prompt.Suggest
}

func NewQueryClient(ctx context.Context, storageType storage_factory.StorageType, storage storage.Storage) (*SQLQueryClient, error) {
	client := &SQLQueryClient{
		storageType: storageType,
		Storage:     storage,
	}

	// TODO BUG: If you switch schema, the hints here will be outdated
	client.initTablesSuggest(ctx)
	client.initColumnsSuggest(ctx)

	return client, nil
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *SQLQueryClient) Run(ctx context.Context) {

	cli_ui.Infof("You can end the session by typing `exit` and press enter, now enter your query statement: \n")

	p := prompt.New(func(in string) {

		in = strings.TrimSpace(in)
		strArr := strings.Split(in, "\\")
		s := strArr[0]

		lowerSql := strings.ToLower(s)
		if lowerSql == "exit" || lowerSql == "exit;" || lowerSql == ".exit" {
			cli_ui.Infof("Bye.")
			os.Exit(0)
		}

		res, err := x.Storage.Query(ctx, s)
		if err != nil {
			cli_ui.Errorln(err)
		} else {
			tables, e := res.ReadRows(-1)
			if e != nil && e.HasError() {
				cli_ui.Errorln(err)
				return
			}
			header := tables.GetColumnNames()
			body := tables.GetMatrix()
			var tableBody [][]string
			for i := range body {
				var row []string
				for j := range body[i] {
					row = append(row, utils.Strava(body[i][j]))
				}
				tableBody = append(tableBody, row)
			}

			// \g or \G use row mode show query result
			if len(strArr) > 1 && (strArr[1] == "g" || strArr[1] == "G") {
				cli_ui.ShowRows(header, tableBody, []string{}, true)
			} else {
				cli_ui.ShowTable(header, tableBody, []string{}, true)
			}

		}

	}, x.completer,
		prompt.OptionTitle("Table"),
		prompt.OptionPrefix("> "),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(buffer *prompt.Buffer) {
				os.Exit(0)
			},
		}),
	)
	p.Run()
}

// ------------------------------------------------- --------------------------------------------------------------------

// if there are no spaces this is the first word
func (x *SQLQueryClient) isFirstWord(text string) bool {
	return strings.LastIndex(text, " ") == -1
}

func (x *SQLQueryClient) completer(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	s := x.formatSuggest(d.Text, text)
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func (x *SQLQueryClient) formatSuggest(text string, before string) []prompt.Suggest {
	var s []prompt.Suggest
	if x.isFirstWord(text) {
		if text != "" {
			s = []prompt.Suggest{
				{Text: "SELECT"},
				{Text: "WITH"},
			}
		}
	} else {
		texts := strings.Split(before, " ")
		if strings.ToLower(texts[len(texts)-2]) == "from" {
			s = x.Tables
		}
		if strings.ToLower(texts[len(texts)-2]) == "select" {
			s = x.Columns
		}
		if strings.ToLower(texts[len(texts)-2]) == "," {
			s = x.Columns
		}
	}
	return s
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *SQLQueryClient) initTablesSuggest(ctx context.Context) {
	res, diag := x.Storage.Query(ctx, x.getTablesSuggestSQL())
	var tables []prompt.Suggest
	if diag != nil {
		_ = cli_ui.PrintDiagnostic(diag.GetDiagnosticSlice())
	} else {
		rows, diag := res.ReadRows(-1)
		if diag != nil {
			_ = cli_ui.PrintDiagnostic(diag.GetDiagnosticSlice())
		}
		for i := range rows.GetMatrix() {
			tableName := rows.GetMatrix()[i][0].(string)
			tables = append(tables, prompt.Suggest{Text: tableName})
		}
	}
	x.Tables = tables
}

func (x *SQLQueryClient) getTablesSuggestSQL() string {
	switch x.storageType {
	case storage_factory.StorageTypePostgresql:
		return TABLESQL
	default:
		return ""
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *SQLQueryClient) initColumnsSuggest(ctx context.Context) {
	rs, err := x.Storage.Query(ctx, x.getColumnsSuggestSQL())
	var columns []prompt.Suggest
	if err != nil {
		_ = cli_ui.PrintDiagnostic(err.GetDiagnosticSlice())
	} else {
		rows, err := rs.ReadRows(-1)
		if err != nil {
			_ = cli_ui.PrintDiagnostic(err.GetDiagnosticSlice())
		}
		for i := range rows.GetMatrix() {
			schemaName := rows.GetMatrix()[i][0].(string)
			tableName := rows.GetMatrix()[i][1].(string)
			columnName := rows.GetMatrix()[i][2].(string)
			columns = append(columns, prompt.Suggest{Text: columnName, Description: schemaName + "." + tableName})
		}
	}
	x.Columns = columns
}

func (x *SQLQueryClient) getColumnsSuggestSQL() string {
	switch x.storageType {
	case storage_factory.StorageTypePostgresql:
		return COLUMNSQL
	default:
		return ""
	}
}

// ------------------------------------------------- --------------------------------------------------------------------
