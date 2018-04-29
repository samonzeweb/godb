package godb

import (
	"time"
)

type Logger interface {
	Println(v ...interface{})
}

// SetLogger sets the logger for the given DB.
// By default there is no logger.
func (db *DB) SetLogger(logger Logger) {
	db.logger = logger
}

// logPrintln is a wrapper for log.Logger.Println with the DB.logger
// as Logger.
func (db *DB) logPrintln(v ...interface{}) {
	if db.logger != nil {
		db.logger.Println(v...)
	}
}

// logDuration adds a log with a duration.
func (db *DB) logDuration(duration time.Duration) {
	if db.logger != nil {
		db.logPrintln("Duration : ", duration)
	}
}
