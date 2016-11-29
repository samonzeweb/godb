package godb

import "database/sql"

// Queryable represents either a Tx, a DB, or a Stmt
type Queryable interface {
	Exec(args ...interface{}) (sql.Result, error)
	Query(args ...interface{}) (*sql.Rows, error)
	QueryRow(args ...interface{}) *sql.Row
}

// The queryable type implements Queryable for sql.DB and sql.Tx
type queryable struct {
	db       PreparableAndQueryable
	sqlQuery string
}

func (q *queryable) Exec(args ...interface{}) (sql.Result, error) {
	return q.db.Exec(q.sqlQuery, args...)
}

func (q *queryable) Query(args ...interface{}) (*sql.Rows, error) {
	return q.db.Query(q.sqlQuery, args...)
}

func (q *queryable) QueryRow(args ...interface{}) *sql.Row {
	return q.db.QueryRow(q.sqlQuery, args...)
}

// EnableStmtCache enable prepared statement cache
// (it is enabled by default)
func (db *DB) EnableStmtCache() {
	db.isPrepStmtEnabled = true
}

// DisableStmtCache disable prepared statement cache
func (db *DB) DisableStmtCache() {
	db.isPrepStmtEnabled = false
}

// IsStmtCacheEnabled return true if the cache of prepared
// statements is enabled.
func (db *DB) IsStmtCacheEnabled() bool {
	return db.isPrepStmtEnabled
}

// getQueryable manage prepared statement, and its cache.
func (db *DB) getQueryable(sql string) (Queryable, error) {
	// Prepared statements are managed only in a Tx, and when the
	// cache is enabled.
	if db.CurrentTx() == nil || db.isPrepStmtEnabled == false {
		wrapper := queryable{
			db:       db.getTxElseDb(),
			sqlQuery: sql,
		}
		return &wrapper, nil
	}

	// Already prepared ?
	prepStmt, ok := db.preparedStmts[sql]
	if ok {
		db.logPrintln("Use cached prepared statement")
		return prepStmt, nil
	}

	// New prepared statement
	db.logPrintln("Prepare statement and cache it")
	prepStmt, err := db.CurrentTx().Prepare(sql)
	if err != nil {
		return nil, err
	}
	db.preparedStmts[sql] = prepStmt
	return prepStmt, nil
}

// clearPreparedStatement clear the prepared statements cache
func (db *DB) resetPreparedStatementsCache() {
	db.preparedStmts = make(map[string]*sql.Stmt)
}
