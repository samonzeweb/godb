package godb

import "database/sql"

// Iterator is an interface to iterate over the result of a sql query
// and scan each row one at a time instead of getting all into one slice.
// The principe is similar to the standard sql.Rows type.
type Iterator interface {
	Next() bool
	Scan(interface{}) error
	Scanx(...interface{}) error
	Close() error
	Err() error
}

// iteratorInternals is the Iterator implementation (hidden)
type iteratorInternals struct {
	rows       *sql.Rows
	recordInfo *recordDescription
	columns    []string
}

// Next prepares the next result row for reading with the Scan method.
// It returns false is case of error or if there are not more data to fetch.
// If the end of the resultset is reached, it automaticaly free ressources like
// the Close method.
func (i *iteratorInternals) Next() bool {
	return i.rows.Next()
}

// Scan fill the given struct with the current row.
func (i *iteratorInternals) Scan(record interface{}) error {
	var err error

	// First scan
	if i.recordInfo == nil {
		// Reflection part
		i.recordInfo, err = buildRecordDescription(record)
		if err != nil {
			return err
		}
	}

	pointers, err := i.recordInfo.structMapping.GetPointersForColumns(record, i.columns...)
	if err != nil {
		return err
	}

	return i.rows.Scan(pointers...)
}

//Scanx scans record values to destination columns
func (i *iteratorInternals) Scanx(dest ...interface{}) error {
	return i.rows.Scan(dest...)
}

// Close frees ressources created by the request execution.
func (i *iteratorInternals) Close() error {
	return i.rows.Close()
}

// Err returns the error that was encountered during iteration, or nil.
// Always check Err after an iteration, like with the standard sql.Err method.
func (i *iteratorInternals) Err() error {
	return i.rows.Err()
}
