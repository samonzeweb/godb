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

// doSelectOrWithReturning executes the statement and fills the auto fields.
// It returns the count of rows returned.
// It is called when the adapter implements ReturningSuffixer.
func (db *DB) doSelectOrWithReturning(query string, arguments []interface{}, recordDescription *recordDescription, pointersGetter pointersGetter) (int64, error) {
	rows, columns, err := db.executeQuery(query, arguments, false, false)
	defer rows.Close()
	if err != nil {
		return 0, err
	}

	// If the given slice is empty, the slice grows as the rows are read.
	// If the given slice isn't empty it's filled with rows, and both rows and
	// slice length have to be equals.
	// If it's a single instance, it's juste filled, and the result must have
	// only one row.
	var rowsCount int
	if recordDescription.len() > 0 {
		rowsCount, err = db.fillWithValues(recordDescription, pointersGetter, columns, rows)
	} else {
		rowsCount, err = db.growAndFillWithValues(recordDescription, pointersGetter, columns, rows)
	}
	if err != nil {
		db.logPrintln("ERROR : ", err)
		return 0, err
	}

	err = rows.Err()
	if err != nil {
		db.logPrintln("ERROR : ", err)
	}
	return int64(rowsCount), err
}

// executeQuery executes the given query with its arguments and returns the
// resulting *sql.Rows, the list of columns names, and an error.
func (db *DB) executeQuery(query string, arguments []interface{}, noTx, noStmtCache bool) (*sql.Rows, []string, error) {
	query = db.replacePlaceholders(query)
	db.logPrintln(query, arguments)

	startTime := time.Now()
	queryable, err := db.getQueryableWithOptions(query, noTx, noStmtCache)
	if err != nil {
		return nil, nil, err
	}
	rows, err := queryable.Query(arguments...)
	condumedTime := timeElapsedSince(startTime)
	db.addConsumedTime(condumedTime)
	db.logDuration(condumedTime)
	if err != nil {
		db.logPrintln("ERROR : ", err)
		return nil, nil, err
	}

	columns, err := rows.Columns()
	if err != nil {
		db.logPrintln("ERROR : ", err)
		rows.Close()
		return nil, nil, err
	}

	return rows, columns, nil
}

// fillWithReturningValues fill the record with rows, the record size must have
// at least the same size has the rows count.
// There could be less rows than awaited, it must be checked by the caller. It's
// not managed here because is could be specific case like optimistic locking failure.
func (db *DB) fillWithValues(recordDescription *recordDescription, pointersGetter pointersGetter, columns []string, rows *sql.Rows) (int, error) {
	rowsCount := 0
	recordLength := recordDescription.len()
	for rows.Next() {
		rowsCount++
		if rowsCount > recordLength {
			return 0, fmt.Errorf("There are more rows returned than the target size : %v", recordLength)
		}
		instancePtr := recordDescription.index(rowsCount - 1)

		pointers, err := pointersGetter(instancePtr, columns)
		if err != nil {
			return 0, err
		}
		err = rows.Scan(pointers...)
		if err != nil {
			return 0, err
		}
	}

	return rowsCount, nil
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

// doWithIterator executes the given query (with its arguments) and returns
// an Iterator.
func (db *DB) doWithIterator(query string, arguments []interface{}) (Iterator, error) {
	rows, columns, err := db.executeQuery(query, arguments, true, true)
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		return nil, err
	}

	iterator := iteratorInternals{
		rows:    rows,
		columns: columns,
	}

	return &iterator, nil
}
