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

	columns          []string
	intoTable        string
	values           [][]interface{}
	returningColumns []string
	suffixes         []string
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

// Returning adds a RETURNING or OUTPUT clause to the statement. Use it with
// PostgreSQL and SQL Server.
func (is *InsertStatement) Returning(columns ...string) *InsertStatement {
	is.returningColumns = append(is.returningColumns, columns...)
	return is
}

// Suffix adds an expression to suffix the statement.
func (is *InsertStatement) Suffix(suffix string) *InsertStatement {
	is.suffixes = append(is.suffixes, suffix)
	return is
}

// ToSQL returns a string with the SQL statement (containing placeholders),
// the arguments slices, and an error.
func (is *InsertStatement) ToSQL() (string, []interface{}, error) {
	// TODO : estimate the buffer size.
	sqlBuffer := newSQLBuffer(is.db.adapter, 256, 16)

	sqlBuffer.Write("INSERT ")
	sqlBuffer.writeInto(is.intoTable)
	sqlBuffer.Write(" (")
	sqlBuffer.writeColumns(is.columns)
	sqlBuffer.Write(") ")
	sqlBuffer.writeReturningForPosition(is.returningColumns, adapters.ReturningSQLServer)
	sqlBuffer.Write("VALUES ")
	sqlBuffer.writeInsertValues(is.values, len(is.columns))
	sqlBuffer.writeReturningForPosition(is.returningColumns, adapters.ReturningPostgreSQL)
	sqlBuffer.writeStringsWithSpaces(is.suffixes)

	return sqlBuffer.SQL(), sqlBuffer.Arguments(), sqlBuffer.Err()
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
	_, ok := is.db.adapter.(adapters.ReturningBuilder)
	if ok {
		// adapters with ReturningSuffixer does not use LastInsertId()
		return 0, nil
	}
	lastInsertID, err := result.LastInsertId()
	return lastInsertID, err
}

// DoWithReturning executes the statement and fills the fields according to
// the columns in RETURNING clause.
func (is *InsertStatement) DoWithReturning(record interface{}) (int64, error) {
	recordDescription, err := buildRecordDescription(record)
	if err != nil {
		return 0, err
	}

	// the function which will return the pointers according to the given columns
	f := func(record interface{}, columns []string) ([]interface{}, error) {
		pointers, err := recordDescription.structMapping.GetPointersForColumns(record, columns...)
		return pointers, err
	}

	return is.doWithReturning(recordDescription, f)
}

// DoWithReturning executes the statement and fills the fields according to
// the columns in RETURNING clause. It returns the count of rows returned.
func (is *InsertStatement) doWithReturning(recordDescription *recordDescription, pointersGetter pointersGetter) (int64, error) {
	query, args, err := is.ToSQL()
	if err != nil {
		return 0, err
	}

	return is.db.doSelectOrWithReturning(query, args, recordDescription, pointersGetter)
}
