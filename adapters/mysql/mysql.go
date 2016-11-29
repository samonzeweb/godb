package mysql

import _ "github.com/go-sql-driver/mysql"

type MySQL struct{}

var Adapter = MySQL{}

func (MySQL) DriverName() string {
	return "mysql"
}

func (MySQL) Quote(identifier string) string {
	return "`" + identifier + "`"
}
