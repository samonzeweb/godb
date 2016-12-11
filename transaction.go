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
	db.logPrintln("SQL : Begin")

	if db.sqlTx != nil {
		return fmt.Errorf("Begin was called multiple times, sql transaction already exists")
	}

	startTime := time.Now()
	tx, err := db.sqlDB.Begin()
	condumedTime := timeElapsedSince(startTime)
	db.addConsumedTime(condumedTime)
	db.logDuration(condumedTime)
	if err != nil {
		return err
	}

	db.sqlTx = tx
	return nil
}

// Commit commits an existing transaction, fails if none exists.
func (db *DB) Commit() error {
	db.logPrintln("SQL : Commit")

	if db.sqlTx == nil {
		return fmt.Errorf("Commit was called without existing sql transaction")
	}

	db.resetPreparedStatementsCache()
	startTime := time.Now()
	err := db.sqlTx.Commit()
	condumedTime := timeElapsedSince(startTime)
	db.addConsumedTime(condumedTime)
	db.logDuration(condumedTime)
	db.sqlTx = nil
	return err
}

// Rollback rollbacks an existing transaction, fails if none exists.
func (db *DB) Rollback() error {
	db.logPrintln("SQL : Rollback")

	if db.sqlTx == nil {
		return fmt.Errorf("Rollback was called without existing sql transaction")
	}

	db.resetPreparedStatementsCache()
	startTime := time.Now()
	err := db.sqlTx.Rollback()
	condumedTime := timeElapsedSince(startTime)
	db.addConsumedTime(condumedTime)
	db.logDuration(condumedTime)
	db.sqlTx = nil
	return err
}

// CurrentTx returns the current Tx (or nil). Don't commit or rollback it
// directly !
func (db *DB) CurrentTx() *sql.Tx {
	return db.sqlTx
}

// getTxElseDb return either the current Tx, or the DB, throught
// the Queryable interface.
func (db *DB) getTxElseDb() preparableAndQueryable {
	if db.sqlTx != nil {
		return db.sqlTx
	}

	return db.sqlDB
}
