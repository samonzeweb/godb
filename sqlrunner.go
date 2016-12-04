package godb

import (
	"database/sql"
	"time"
)

// do executes the given query (with its arguments) after replacing the
// placeholders if neeeded, and returns sql.Result.
func (db *DB) do(query string, arguments []interface{}) (sql.Result, error) {
	query = db.replacePlaceholders(query)
	db.logPrintln(query, arguments)

	// Execute the UPDATE statement
	startTime := time.Now()
	queryable, err := db.getQueryable(query)
	if err != nil {
		return nil, err
	}
	result, err := queryable.Exec(arguments...)
	condumedTime := timeElapsedSince(startTime)
	db.addConsumedTime(condumedTime)
	db.logDuration(condumedTime)
	if err != nil {
		db.logPrintln("ERROR : ", err)
		return nil, err
	}

	return result, err
}
