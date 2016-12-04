package godb

import (
	"database/sql"
	"time"
)

// pointersGetter is a func type, returning a list of pointers (and error) for
// a given instance pointer and a columns names list.
type pointersGetter func(record interface{}, columns []string) ([]interface{}, error)

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

// doWithReturning executes the statement and fills the auto fields.
// It is called when the adapter implements InsertReturningSuffixer.
func (db *DB) doWithReturning(query string, arguments []interface{}, recordDescription *recordDescription, pointersGetter pointersGetter) error {
	query = db.replacePlaceholders(query)
	db.logPrintln(query, arguments)

	startTime := time.Now()
	queryable, err := db.getQueryable(query)
	if err != nil {
		return err
	}
	rows, err := queryable.Query(arguments...)
	condumedTime := timeElapsedSince(startTime)
	db.addConsumedTime(condumedTime)
	db.logDuration(condumedTime)
	if err != nil {
		db.logPrintln("ERROR : ", err)
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		db.logPrintln("ERROR : ", err)
		return err
	}

	index := 0
	for rows.Next() {
		instancePtr := recordDescription.index(index)
		index++
		pointers, innererr := pointersGetter(instancePtr, columns)
		if innererr != nil {
			return innererr
		}
		innererr = rows.Scan(pointers...)
		if innererr != nil {
			db.logPrintln("ERROR : ", innererr)
			return innererr
		}
	}

	err = rows.Err()
	if err != nil {
		db.logPrintln("ERROR : ", err)
	}
	return err
}
