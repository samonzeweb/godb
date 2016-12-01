package mssql

import (
	_ "github.com/denisenkom/go-mssqldb"
	"gitlab.com/samonzeweb/godb/adapters"
)

type MSSQL struct{}

var Adapter = MSSQL{}

func (MSSQL) DriverName() string {
	return "mssql"
}

func (MSSQL) Quote(identifier string) string {
	return "[" + identifier + "]"
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
