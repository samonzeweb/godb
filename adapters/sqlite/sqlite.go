package sqlite

import _ "github.com/mattn/go-sqlite3"

type SQLite struct{}

func (SQLite) DriverName() string {
	return "sqlite"
}
