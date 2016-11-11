// TODO: add package documentation
package godb

import (
	"database/sql"

	"gitlab.com/samonzeweb/godb/adapters"
)

// TODO
type DB struct {
	adapter adapters.DriverName
	sqlDB   *sql.DB
}

const Placeholder string = "?"

// Open create a new DB struct and initialise a sql.DB connection.
func Open(adapter adapters.DriverName, dataSourceName string) (*DB, error) {
	db := DB{adapter: adapter}
	var err error
	db.sqlDB, err = sql.Open(adapter.DriverName(), dataSourceName)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (db *DB) SelectFrom(tableName string) *selectStatement {
	return newSelectStatement(db, tableName)
}
