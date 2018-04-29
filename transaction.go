package godb

import (
	"database/sql"
	"fmt"
	"time"
)

// preparableAndQueryable represents either a Tx or DB.
type preparableAndQueryable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
}

// Begin starts a new transaction, fails if there is already one.
func (db *DB) Begin() error {

	if db.sqlTx != nil {
		return fmt.Errorf("Begin was called multiple times, sql transaction already exists")
	}

	startTime := time.Now()
	tx, err := db.sqlDB.Begin()
	consumedTime := timeElapsedSince(startTime)
	db.addConsumedTime(consumedTime)
	db.logExecution(consumedTime, "BEGIN")
	if err != nil {
		db.logExecutionErr(err, "BEGIN")
		return err
	}

	db.sqlTx = tx
	return nil
}

// Commit commits an existing transaction, fails if none exists.
func (db *DB) Commit() error {

	if db.sqlTx == nil {
		return fmt.Errorf("Commit was called without existing sql transaction")
	}

	db.stmtCacheTx.clearWithoutClosingStmt()
	startTime := time.Now()
	err := db.sqlTx.Commit()
	consumedTime := timeElapsedSince(startTime)
	db.addConsumedTime(consumedTime)
	db.logExecution(consumedTime, "COMMIT")
	db.sqlTx = nil
	if err!=nil {
		db.logExecutionErr(err, "COMMIT")
	}
	return err
}

// Rollback rollbacks an existing transaction, fails if none exists.
func (db *DB) Rollback() error {
	if db.sqlTx == nil {
		return fmt.Errorf("Rollback was called without existing sql transaction")
	}

	db.stmtCacheTx.clearWithoutClosingStmt()
	startTime := time.Now()
	err := db.sqlTx.Rollback()
	consumedTime := timeElapsedSince(startTime)
	db.addConsumedTime(consumedTime)
	db.logExecution(consumedTime, "ROLLBACK")
	if err!=nil {
		db.logExecutionErr(err, "ROLLBACK")
	}
	db.sqlTx = nil
	return err
}

// CurrentTx returns the current Tx (or nil). Don't commit or rollback it
// directly !
func (db *DB) CurrentTx() *sql.Tx {
	return db.sqlTx
}
