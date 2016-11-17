// TODO: add package documentation
package godb

import (
	"database/sql"
	"log"

	"gitlab.com/samonzeweb/godb/adapters"
)

// DB store a connection to the database, and others data like transaction,
// logger, ... Everything starts with a DB.
// DB is not thread safe.
type DB struct {
	adapter adapters.DriverName
	sqlDB   *sql.DB
	sqlTx   *sql.Tx
	logger  *log.Logger
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

// Clone create a copy of an existing DB, without the current transaction.
// Use it to create new DB object before starting a goroutine.
func (db *DB) Clone() *DB {
	return &DB{
		adapter: db.adapter,
		sqlDB:   db.sqlDB,
		sqlTx:   nil,
		logger:  db.logger,
	}
}

// Close close an existing DB created by Open.
// Dont't close a cloned DB used by others goroutines !
// Don't use a DB anymore after a call to Close.
func (db *DB) Close() error {
	db.LogPrintln("CLOSE DB")
	if db.sqlTx != nil {
		db.LogPrintln("Warning, there is a current transaction")
	}
	return db.sqlDB.Close()
}
