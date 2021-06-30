package sqlite

import (
	"strings"

	"github.com/samonzeweb/godb/dberror"
)

type SQLite struct{}

var Adapter = SQLite{}

func (SQLite) DriverName() string {
	return "sqlite3"
}

func (SQLite) Quote(identifier string) string {
	return "\"" + identifier + "\""
}

func (SQLite) ParseError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "constraint failed") && strings.Contains(errMsg, "foreign key "):
		return dberror.UniqueConstraint{Message: errMsg, Field: strings.Split(strings.Split(errMsg, "failed: ")[1], ".")[1], Err: err}
	case strings.Contains(errMsg, "constraint failed"):
		return dberror.CheckConstraint{Message: errMsg, Field: strings.Split(strings.Split(errMsg, "failed: ")[1], ".")[1], Err: err}
	}

	return err
}
