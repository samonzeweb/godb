package mysql

import (
	"github.com/samonzeweb/godb/dberror"
	"strings"
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
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "Error 1062:"):
		return dberror.UniqueConstraint{Message: errMsg, Field: dberror.ExtractStr(errMsg, "key '", "'"), Err: err}
	case strings.Contains(errMsg, "Error 1452:"):
		return dberror.CheckConstraint{Message: errMsg, Field: dberror.ExtractStr(errMsg, "CONSTRAINT `", "`"), Err: err}
	}

	return err
}
