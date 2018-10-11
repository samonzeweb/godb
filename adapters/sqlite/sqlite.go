package sqlite

import (
	"strings"

	"github.com/samonzeweb/godb/dberror"

	sqlite3 "github.com/mattn/go-sqlite3"
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

	e, _ := err.(sqlite3.Error)
	switch e.ExtendedCode {
	case sqlite3.ErrConstraintUnique:
		return dberror.UniqueConstraint{Message: e.Error(), Field: strings.Split(strings.Split(e.Error(), "failed: ")[1], ".")[1], Err: e}
	case sqlite3.ErrConstraintCheck:
		return dberror.CheckConstraint{Message: e.Error(), Field: strings.Split(strings.Split(e.Error(), "failed: ")[1], ".")[1], Err: e}
	default:
		return err
	}
}
