package mssql

import (
	"bytes"

	"github.com/samonzeweb/godb/adapters"
	"github.com/samonzeweb/godb/dbreflect"

	_ "github.com/denisenkom/go-mssqldb"
)

// init registers types of mssql package corresponding to fields values
func init() {
	dbreflect.RegisterScannableStruct(Rowversion{})
}

type MSSQL struct{}

var Adapter = MSSQL{}

func (MSSQL) DriverName() string {
	return "mssql"
}

func (MSSQL) Quote(identifier string) string {
	return "[" + identifier + "]"
}

func (m MSSQL) ReturningBuild(columns []string) string {
	suffixBuffer := bytes.NewBuffer(make([]byte, 0, 16*len(columns)+1))
	suffixBuffer.WriteString("OUTPUT ")
	for i, column := range columns {
		if i > 0 {
			suffixBuffer.WriteString(", ")
		}
		suffixBuffer.WriteString(column)
	}
	return suffixBuffer.String()
}

func (m MSSQL) FormatForNewValues(columns []string) []string {
	formatedColumns := make([]string, 0, len(columns))
	for _, column := range columns {
		formatedColumns = append(formatedColumns, "INSERTED."+m.Quote(column))
	}
	return formatedColumns
}

func (m MSSQL) GetReturningPosition() adapters.ReturningPosition {
	return adapters.ReturningSQLServer
}

func (MSSQL) BuildLimit(limit int) *adapters.SQLPart {
	sqlPart := adapters.SQLPart{}
	sqlPart.Sql = "FETCH NEXT ? ROWS ONLY"
	sqlPart.Arguments = make([]interface{}, 0, 1)
	sqlPart.Arguments = append(sqlPart.Arguments, limit)
	return &sqlPart
}

func (MSSQL) BuildOffset(offset int) *adapters.SQLPart {
	sqlPart := adapters.SQLPart{}
	sqlPart.Sql = "OFFSET ? ROWS"
	sqlPart.Arguments = make([]interface{}, 0, 1)
	sqlPart.Arguments = append(sqlPart.Arguments, offset)
	return &sqlPart
}

func (MSSQL) IsOffsetFirst() bool {
	return true
}
