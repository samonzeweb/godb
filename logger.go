package godb

import (
	"time"
	"fmt"
)

var logPrefix = "SQL:"
var prefixedLog []interface{}

type Logger interface {
	Println(v ...interface{})
}

// SetLogger sets the logger for the given DB.
// By default there is no logger.
func (db *DB) SetLogger(logger Logger) {
	db.logger = logger
	prefixedLog = append(prefixedLog, logPrefix)
}

// logPrintln is a wrapper for log.Logger.Println with the DB.logger
// as Logger.
func (db *DB) logPrintln(v ...interface{}) {
	if db.logger != nil {
		db.logger.Println(append(prefixedLog,v)...)
	}
}

// logExecution adds a log with a duration and SQL statement.
func (db *DB) logExecution(duration time.Duration, v ...interface{}) {
	if db.logger != nil {
		db.logger.Println(append(prefixedLog, v, fmt.Sprintf("(Duration: %v)", duration))...)
	}
}

// logExecution adds a log with a duration and SQL statement.
func (db *DB) logExecutionErr(err error, v ...interface{}) {
	if db.logger != nil {
		db.logger.Println(append(prefixedLog, v, fmt.Sprintf("(ERROR: %v)", err))...)
	}
}
