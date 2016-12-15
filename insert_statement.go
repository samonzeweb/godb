package godb

import "github.com/samonzeweb/godb/adapters"

// InsertStatement is an INSERT statement builder.
// Initialize it with the InsertInto method.
//
// Examples :
//
// 	id, err := db.InsertInto("bar").
// 		Columns("foo", "baz").
// 		Values(2, "something").
// 		Do()
type InsertStatement struct {
	db *DB

	columns   []string
	intoTable string
	values    [][]interface{}
	suffixes  []string
}

// InsertInto initializes a INSERT statement builder
func (db *DB) InsertInto(tableName string) *InsertStatement {
	ip := &InsertStatement{db: db}
	ip.intoTable = tableName
	return ip
}

// Columns adds columns to insert.
func (is *InsertStatement) Columns(columns ...string) *InsertStatement {
	is.columns = append(is.columns, columns...)
	return is
}

// Values add values to insert.
func (is *InsertStatement) Values(values ...interface{}) *InsertStatement {
	is.values = append(is.values, values)
	return is
}

// Suffix adds an expression to suffix the statement. Use it to add a
// RETURNING clause with PostgreSQL (or whatever you need).
func (is *InsertStatement) Suffix(suffix string) *InsertStatement {
	is.suffixes = append(is.suffixes, suffix)
	return is
}

// ToSQL returns a string with the SQL statement (containing placeholders),
// the arguments slices, and an error.
func (is *InsertStatement) ToSQL() (string, []interface{}, error) {

	// TODO : estimate the buffer size.
	sqlBuffer := newSQLBuffer(is.db.adapter, 256, 16)

	sqlBuffer.write("INSERT ")

	if err := sqlBuffer.writeInto(is.intoTable); err != nil {
		return "", nil, err
	}

	sqlBuffer.write(" (")
	if err := sqlBuffer.writeColumns(is.columns); err != nil {
		return "", nil, err
	}
	sqlBuffer.write(") VALUES ")

	if err := sqlBuffer.writeInsertValues(is.values, len(is.columns)); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeStrings(is.suffixes); err != nil {
		return "", nil, err
	}

	return sqlBuffer.sqlString(), sqlBuffer.sqlArguments(), nil
}

// Do executes the builded INSERT statement and returns the creadted 'id' if
// the adapter does not implement InsertReturningSuffixer.
func (is *InsertStatement) Do() (int64, error) {
	query, args, err := is.ToSQL()
	if err != nil {
		return 0, err
	}

	result, err := is.db.do(query, args)
	if err != nil {
		return 0, err
	}

	// Return the created 'Id' (if available)
	_, ok := is.db.adapter.(adapters.InsertReturningSuffixer)
	if ok {
		// adapters with InsertSuffixer does not use LastInsertId()
		return 0, nil
	}
	lastInsertID, err := result.LastInsertId()
	return lastInsertID, err
}

// DoWithReturning executes the statement and fills the fields according to
// the columns in RETURNING clause.
func (is *InsertStatement) DoWithReturning(record interface{}) error {
	recordDescription, err := buildRecordDescription(record)
	if err != nil {
		return err
	}

	// the function which will return the pointers according to the given columns
	f := func(record interface{}, columns []string) ([]interface{}, error) {
		pointers, err := recordDescription.structMapping.GetPointersForColumns(record, columns...)
		return pointers, err
	}

	return is.doWithReturning(recordDescription, f)
}

// DoWithReturning executes the statement and fills the fields according to
// the columns in RETURNING clause.
func (is *InsertStatement) doWithReturning(recordDescription *recordDescription, pointersGetter pointersGetter) error {
	query, args, err := is.ToSQL()
	if err != nil {
		return err
	}

	return is.db.doWithReturning(query, args, recordDescription, pointersGetter)
}
