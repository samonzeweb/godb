package godb

import (
	"log"
	"time"
)

// Set the logger for the given DB.
// By default there is no logger.
func (db *DB) SetLogger(logger *log.Logger) {
	db.logger = logger
}

// logPrintln is just a wrapper for log.Logger.Println with the DB.logger
// as Logger
func (db *DB) logPrintln(v ...interface{}) {
	if db.logger != nil {
		db.logger.Println(v...)
	}
}

// logDuration simply add a log with a duration
func (db *DB) logDuration(duration time.Duration) {
	if db.logger != nil {
		db.logPrintln("Duration : ", duration)
	}
}
