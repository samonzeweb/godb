package mysql

import (
	"github.com/go-sql-driver/mysql"
	"github.com/samonzeweb/godb/dberror"
)

type MySQL struct{}

var Adapter = MySQL{}

func (MySQL) DriverName() string {
	return "mysql"
}

func (MySQL) Quote(identifier string) string {
	return "`" + identifier + "`"
}

func (MySQL) ParseError(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*mysql.MySQLError); ok {
		switch e.Number {
		case 1062:
			return dberror.UniqueConstraint{Message: e.Error(), Field: dberror.ExtractStr(e.Message, "key '", "'"), Err: e}
		case 1452:
			return dberror.CheckConstraint{Message: e.Error(), Field: dberror.ExtractStr(e.Message, "CONSTRAINT `", "`"), Err: e}
		}
	}

	return err
}
