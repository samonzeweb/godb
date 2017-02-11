package godb

import (
	"database/sql"
	"fmt"
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

	// Execute the statement
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

	// If the given slice is empty, the slice grows as the rows are read.
	// If the given slice isn't empty it's filled with rows, and both rows and
	// slice length have to be equals.
	// If it's a single instance, it's juste filled, and the result must have
	// only one row.
	if recordDescription.len() > 0 {
		err = db.fillWithValues(recordDescription, pointersGetter, columns, rows)
	} else {
		_, err = db.growAndFillWithValues(recordDescription, pointersGetter, columns, rows)
	}
	if err != nil {
		db.logPrintln("ERROR : ", err)
		return err
	}

	err = rows.Err()
	if err != nil {
		db.logPrintln("ERROR : ", err)
	}
	return err
}

// fillWithReturningValues fill the record with rows, the record size must has
// the same size has the rows count.
func (db *DB) fillWithValues(recordDescription *recordDescription, pointersGetter pointersGetter, columns []string, rows *sql.Rows) error {
	rowsCount := 0
	recordLength := recordDescription.len()
	for rows.Next() {
		rowsCount++
		if rowsCount > recordLength {
			return fmt.Errorf("There are more rows returned than the target size : %v", recordLength)
		}
		instancePtr := recordDescription.index(rowsCount - 1)

		pointers, err := pointersGetter(instancePtr, columns)
		if err != nil {
			return err
		}
		err = rows.Scan(pointers...)
		if err != nil {
			return err
		}
	}

	if rowsCount < recordLength {
		return fmt.Errorf("There are less rows returned than the target size, rows count : %v", rowsCount)
	}

	return nil
}

// growAndFillWithReturningValues fill the record with rows, and make it growing.
func (db *DB) growAndFillWithValues(recordDescription *recordDescription, pointersGetter pointersGetter, columns []string, rows *sql.Rows) (int, error) {
	rowsCount := 0
	for rows.Next() {
		rowsCount++
		if rowsCount > 1 && !recordDescription.isSlice {
			return 0, fmt.Errorf("There are multiple rows for a single instance")
		}
		err := recordDescription.fillRecord(
			// Fill one instance with one row
			func(record interface{}) error {
				pointers, err := pointersGetter(record, columns)
				if err != nil {
					return err
				}
				err = rows.Scan(pointers...)
				if err != nil {
					return err
				}
				return nil
			})

		if err != nil {
			return 0, err
		}
	}

	return rowsCount, nil
}
