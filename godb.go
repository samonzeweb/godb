// TODO: add package documentation
package godb

import (
	"database/sql"
	"log"
	"time"

	"gitlab.com/samonzeweb/godb/adapters"
)

// DB store a connection to the database, and others data like transaction,
// logger, ... Everything starts with a DB.
// DB is not thread safe.
type DB struct {
	adapter      adapters.DriverNamer
	sqlDB        *sql.DB
	sqlTx        *sql.Tx
	logger       *log.Logger
	consumedTime time.Duration
}

const Placeholder string = "?"

// Open create a new DB struct and initialise a sql.DB connection.
func Open(adapter adapters.DriverNamer, dataSourceName string) (*DB, error) {
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
		adapter:      db.adapter,
		sqlDB:        db.sqlDB,
		sqlTx:        nil,
		logger:       db.logger,
		consumedTime: 0,
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

// ConsumedTime returns the time consumed by SQL queries executions
// The duration is reseted when the DB is cloned.
func (db *DB) ConsumedTime() time.Duration {
	return db.consumedTime
}

// Reset the time consumed by SQL queries executions
func (db *DB) ResetConsumedTime() {
	db.consumedTime = 0
}

// Reset the time consumed by SQL queries executions
func (db *DB) addConsumedTime(duration time.Duration) {
	db.consumedTime += duration
}

// timeElapsedSince returns the time elapsed (duration) since a given
// start time.
func timeElapsedSince(startTime time.Time) time.Duration {
	return time.Now().Sub(startTime)
}

// quote returns all strings given quoted by the adapter if it implements
// the Quoter interface, or the given strings slice.
func (db *DB) quoteAll(identifiers []string) []string {
	if quoter, ok := db.adapter.(adapters.Quoter); ok {
		quotedIdentifiers := make([]string, 0, len(identifiers))
		for _, identifier := range identifiers {
			quotedIdentifiers = append(quotedIdentifiers, quoter.Quote(identifier))
		}
		return quotedIdentifiers
	}

	return identifiers
}
