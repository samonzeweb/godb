package godb

import (
	"database/sql"
)

// queryable represents either a Tx, a DB, or a Stmt.
type queryable interface {
	Exec(args ...interface{}) (sql.Result, error)
	Query(args ...interface{}) (*sql.Rows, error)
	QueryRow(args ...interface{}) *sql.Row
}

// The queryWrapper type implements Queryable for sql.DB and sql.Tx
type queryWrapper struct {
	db       preparableAndQueryable
	sqlQuery string
}

// Exec wraps the Exec method for sql.DB or sql.Tx.
func (q *queryWrapper) Exec(args ...interface{}) (sql.Result, error) {
	return q.db.Exec(q.sqlQuery, args...)
}

// Query wraps the Query method for sql.DB or sql.Tx.
func (q *queryWrapper) Query(args ...interface{}) (*sql.Rows, error) {
	return q.db.Query(q.sqlQuery, args...)
}

// QueryRow wraps the QueryRow method for sql.DB or sql.Tx.
func (q *queryWrapper) QueryRow(args ...interface{}) *sql.Row {
	return q.db.QueryRow(q.sqlQuery, args...)
}

// getQueryable manages prepared statement, and its cache.
func (db *DB) getQueryable(query string) (queryable, error) {
	// One cache for sql.DB, and one for sql.Tx
	var cache *StmtCache

	if db.CurrentTx() == nil {
		cache = db.stmtCacheDB
	} else {
		cache = db.stmtCacheTx
	}

	// If the cache is disabled, just return a wrapper to look like a
	// prepared statement.
	if !cache.IsEnabled() {
		wrapper := queryWrapper{
			db:       db.getTxElseDb(),
			sqlQuery: query,
		}
		return &wrapper, nil
	}

	// Already prepared ?
	stmt := cache.get(query)
	if stmt != nil {
		db.logPrintln("Use cached prepared statement")
		return stmt, nil
	}

	// New prepared statement
	db.logPrintln("Prepare statement and cache it")
	stmt, err := db.getTxElseDb().Prepare(query)
	if err != nil {
		return nil, err
	}
	cache.add(query, stmt)
	return stmt, nil
}

// StmtCacheDB returns the prepared statement cache for queries outside a
// transaction (run with sql.DB).
func (db *DB) StmtCacheDB() *StmtCache {
	return db.stmtCacheDB
}

// StmtCacheTx returns the prepared statement cache for queries inside a
// transaction (run with sql.Tx).
func (db *DB) StmtCacheTx() *StmtCache {
	return db.stmtCacheTx
}
