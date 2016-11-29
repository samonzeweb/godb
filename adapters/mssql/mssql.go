package mssql

import _ "github.com/denisenkom/go-mssqldb"

type MSSQL struct{}

var Adapter = MSSQL{}

func (MSSQL) DriverName() string {
	return "mssql"
}

func (MSSQL) Quote(identifier string) string {
	return "[" + identifier + "]"
}
