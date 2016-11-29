package mysql

import _ "github.com/go-sql-driver/mysql"

type PostgreSQL struct{}

var Adapter = PostgreSQL{}

func (PostgreSQL) DriverName() string {
	return "mysql"
}

func (PostgreSQL) Quote(identifier string) string {
	return "`" + identifier + "`"
}
