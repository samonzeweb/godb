package godb

import "database/sql"

// RawSQL allows the execution of a custom SQL query.
// Note : the API to run a custom query could have been build without an
// intermediaite struct. But this produce a mode homogeneous API, and allows
// later evolutions without breaking the API.
type RawSQL struct {
	db        *DB
	sql       string
	arguments []interface{}
}

// RawSQL create a RawSQL structure.
func (db *DB) RawSQL(sql string, args ...interface{}) *RawSQL {
	return &RawSQL{
		db:        db,
		sql:       sql,
		arguments: args,
	}
}

// Do executes the raw query.
// The record argument has to be a pointer to a struct or a slice.
// If the argument is not a slice, a row is expected, and Do returns
// sql.ErrNoRows is none where found.
func (raw *RawSQL) Do(record interface{}) error {
	recordInfo, err := buildRecordDescription(record)
	if err != nil {
		return err
	}

	// the function which will return the pointers according to the given columns
	pointersGetter := func(record interface{}, columns []string) ([]interface{}, error) {
		var pointers []interface{}
		pointers, err := recordInfo.structMapping.GetPointersForColumns(record, columns...)
		return pointers, err
	}

	rowsCount, err := raw.db.doSelectOrWithReturning(raw.sql, raw.arguments, recordInfo, pointersGetter)
	if err != nil {
		return err
	}

	// When a single instance is requested but not found, sql.ErrNoRows is
	// returned like QueryRow in database/sql package.
	if !recordInfo.isSlice && rowsCount == 0 {
		err = sql.ErrNoRows
	}

	return err
}
