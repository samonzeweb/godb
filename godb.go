package godb

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/samonzeweb/godb/adapters"
)

// DB stores a connection to the database, the current transaction, logger, ...
// Everything starts with a DB.
// DB is not thread safe (see Clone).
type DB struct {
	adapter      adapters.Adapter
	sqlDB        *sql.DB
	sqlTx        *sql.Tx
	logger       *log.Logger
	consumedTime time.Duration

	// Prepared Statement cache for DB and Tx
	stmtCacheDB *StmtCache
	stmtCacheTx *StmtCache
}

// Placeholder is the placeholder string, use it to build queries.
// Adapters could change it before queries are executed.
const Placeholder string = "?"

// ErrOpLock is an error returned when Optimistic Locking failure occurs
var ErrOpLock = errors.New("optimistic locking failure")

// Open creates a new DB struct and initialise a sql.DB connection.
func Open(adapter adapters.Adapter, dataSourceName string) (*DB, error) {
	db := DB{
		adapter:     adapter,
		stmtCacheDB: newStmtCache(),
		stmtCacheTx: newStmtCache(),
	}

	// Prepared statements cache is disabled by default except for Tx
	db.stmtCacheDB.Disable()

	var err error
	db.sqlDB, err = sql.Open(adapter.DriverName(), dataSourceName)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

// Clone creates a copy of an existing DB, without the current transaction.
// The clone has consumedTime set to zero, and new prepared statements caches with
// the same characteristics.
// Use it to create new DB object before starting a goroutine.
func (db *DB) Clone() *DB {
	clone := &DB{
		adapter:      db.adapter,
		sqlDB:        db.sqlDB,
		sqlTx:        nil,
		logger:       db.logger,
		consumedTime: 0,
		stmtCacheDB:  newStmtCache(),
		stmtCacheTx:  newStmtCache(),
	}

	clone.stmtCacheDB.SetSize(db.stmtCacheDB.GetSize())
	if !db.stmtCacheDB.IsEnabled() {
		clone.stmtCacheDB.Disable()
	}

	clone.stmtCacheTx.SetSize(db.stmtCacheTx.GetSize())
	if !db.stmtCacheTx.IsEnabled() {
		clone.stmtCacheTx.Disable()
	}

	return clone
}

// Wrap creates a godb.DB by using provided and initialized sql.DB Helpful for
// using custom configured sql.DB instance for godb. Can be used before
// starting a goroutine.
func (db *DB) Wrap(dbInst *sql.DB) *DB {
	clone := &DB{
		adapter:      db.adapter,
		sqlDB:        dbInst,
		sqlTx:        nil,
		logger:       db.logger,
		consumedTime: 0,
		stmtCacheDB:  newStmtCache(),
		stmtCacheTx:  newStmtCache(),
	}

	clone.stmtCacheDB.SetSize(db.stmtCacheDB.GetSize())
	if !db.stmtCacheDB.IsEnabled() {
		clone.stmtCacheDB.Disable()
	}

	clone.stmtCacheTx.SetSize(db.stmtCacheTx.GetSize())
	if !db.stmtCacheTx.IsEnabled() {
		clone.stmtCacheTx.Disable()
	}

	return clone
}

// Close closes an existing DB created by Open.
// Don't close a cloned DB still used by others goroutines as the sql.DB
// is shared !
// Don't use a DB anymore after a call to Close.
func (db *DB) Close() error {
	db.logPrintln("CLOSE DB")
	if db.sqlTx != nil {
		db.logPrintln("Warning, there is a current transaction")
	}
	return db.sqlDB.Close()
}

// Adapter returns the current adapter.
func (db *DB) Adapter() adapters.Adapter {
	return db.adapter
}

// CurrentDB returns the current *sql.DB.
// Use it wisely.
func (db *DB) CurrentDB() *sql.DB {
	return db.sqlDB
}

// ConsumedTime returns the time consumed by SQL queries executions
// The duration is reseted when the DB is cloned.
func (db *DB) ConsumedTime() time.Duration {
	return db.consumedTime
}

// ResetConsumedTime resets the time consumed by SQL queries executions
func (db *DB) ResetConsumedTime() {
	db.consumedTime = 0
}

// addConsumedTime adds duration to the consumed time
func (db *DB) addConsumedTime(duration time.Duration) {
	db.consumedTime += duration
}

// timeElapsedSince returns the time elapsed (duration) since a given
// start time.
func timeElapsedSince(startTime time.Time) time.Duration {
	return time.Now().Sub(startTime)
}

// quoteAll returns all strings given quoted by the adapter.
func (db *DB) quoteAll(identifiers []string) []string {
	quotedIdentifiers := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		quotedIdentifiers = append(quotedIdentifiers, db.adapter.Quote(identifier))
	}
	return quotedIdentifiers
}

// replacePlaceholders uses the adapter to change placeholders according to
// the database used.
func (db *DB) replacePlaceholders(sql string) string {
	placeholderReplacer, ok := (db.adapter).(adapters.PlaceholdersReplacer)
	if !ok {
		return sql
	}

	return placeholderReplacer.ReplacePlaceholders(Placeholder, sql)
}
