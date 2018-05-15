package godb

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/samonzeweb/godb/adapters"
	"github.com/samonzeweb/godb/tablenamer"
)


// DB stores a connection to the database, the current transaction, logger, ...
// Everything starts with a DB.
// DB is not thread safe (see Clone).
type DB struct {
	adapter      adapters.Adapter
	sqlDB        *sql.DB
	sqlTx        *sql.Tx	
	logger       Logger
	consumedTime time.Duration
	// Called to format db table name if TableName() func is not defined for model struct
	defaultTableNamer tablenamer.TableNamerFn
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
	dbInst, err := sql.Open(adapter.DriverName(), dataSourceName)
	if err != nil {
		return nil, err
	}
	return initialize(adapter, dbInst), nil
}

// Wrap creates a godb.DB by using provided and initialized sql.DB Helpful for
// using custom configured sql.DB instance for godb. Can be used before
// starting a goroutine.
func Wrap(adapter adapters.Adapter, dbInst *sql.DB) *DB {
	return initialize(adapter, dbInst)
}

// initialize a new godb.DB struct
func initialize(adapter adapters.Adapter, dbInst *sql.DB) *DB {
	db := DB{
		adapter:     adapter,
		sqlDB:       dbInst,
		defaultTableNamer: tablenamer.TableNamerSame(),
		stmtCacheDB: newStmtCache(),
		stmtCacheTx: newStmtCache(),
	}

	// Prepared statements cache is disabled by default except for Tx
	db.stmtCacheDB.Disable()

	return &db
}

// Clone creates a copy of an existing DB, without the current transaction.
// The clone has consumedTime set to zero, and new prepared statements caches with
// the same characteristics.
// Use it to create new DB object before starting a goroutine.
// Use Clear when a clone is not longer useful to free ressources.
func (db *DB) Clone() *DB {
	clone := &DB{
		adapter:      db.adapter,
		sqlDB:        db.sqlDB,
		sqlTx:        nil,
		logger:       db.logger,
		consumedTime: 0,
		defaultTableNamer: db.defaultTableNamer,
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

// Clear closes current transaction (rollback) and frees statements caches.
// It does not close de underlying database connection.
// Use Clear when a clone of godb is not longer useful, or when
// you don't use anymore godb but want to keep the underlying database
// connection open.
func (db *DB) Clear() error {
	if db.sqlTx != nil {
		db.logPrintln("Warning, there is a current transaction")
		if err := db.sqlTx.Rollback(); err != nil {
			return err
		}
	}
	return db.stmtCacheDB.Clear()
}

// Close closes an existing DB created by Open.
// Don't close a cloned DB still used by others goroutines as the sql.DB
// is shared !
// Don't use a DB anymore after a call to Close.
func (db *DB) Close() error {
	db.logPrintln("CLOSE DB")
	db.Clear()
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
		var quotedIdentifier string
		parts := strings.Split(identifier, ".")
		if len(parts) == 1 {
			// just column
			quotedIdentifier = db.adapter.Quote(identifier)
		} else {
			// table.column
			quotedIdentifier = db.adapter.Quote(parts[0]) + "." + db.adapter.Quote(parts[1])
		}
		quotedIdentifiers = append(quotedIdentifiers, quotedIdentifier)
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

func (db *DB) SetDefaultTableNamer(tnamer tablenamer.TableNamerFn) {
	db.defaultTableNamer = tnamer
}
