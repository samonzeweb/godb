package godb

import (
	"log"
	"time"
)

func (db *DB) SetLogger(logger *log.Logger) {
	db.logger = logger
}

func (db *DB) LogPrintln(v ...interface{}) {
	if db.logger != nil {
		db.logger.Println(v...)
	}
}

func (db *DB) LogDuration(duration time.Duration) {
	if db.logger != nil {
		db.LogPrintln("Duration : ", duration)
	}
}
