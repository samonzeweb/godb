package godb

import "github.com/samonzeweb/godb/adapters"

// UpdateStatement will contains all parts needed to build an UPDATE statement.
// Initialize it with the UpdateTable method.
//
// Example :
// 	count, err := db.UpdateTable("bar").
// 		Set("foo", 1).
// 		Where("foo = ?", 2).
// 		Do()
type UpdateStatement struct {
	db *DB

	updateTable      string
	sets             []*setPart
	where            []*Condition
	returningColumns []string
	suffixes         []string
}

// setPart contains elements for a single SET clause.
// The value could be nil for a raw clause (ie count=count+1)
type setPart struct {
	// The column name, or the full SET clause for a raw clause
	column string
	// The value, or nil if it's a raw clause
	value interface{}
}

// UpdateTable creates an UpdateStatement and specify table to update.
// It's the entry point to build an UPDATE query.
func (db *DB) UpdateTable(tableName string) *UpdateStatement {
	us := &UpdateStatement{db: db}
	us.updateTable = tableName
	return us
}

// Set adds a part of SET clause to the query.
func (us *UpdateStatement) Set(column string, value interface{}) *UpdateStatement {
	setClause := &setPart{
		column: column,
		value:  value,
	}
	us.sets = append(us.sets, setClause)
	return us
}

// SetRaw adds a raw SET clause to the query.
func (us *UpdateStatement) SetRaw(rawSQL string) *UpdateStatement {
	rawSetClause := &setPart{
		column: rawSQL,
		value:  nil,
	}
	us.sets = append(us.sets, rawSetClause)
	return us
}

// Where adds a condition using string and arguments.
func (us *UpdateStatement) Where(sql string, args ...interface{}) *UpdateStatement {
	return us.WhereQ(Q(sql, args...))
}

// WhereQ adds a simple or complex predicate generated with Q and
// confunctions.
func (us *UpdateStatement) WhereQ(condition *Condition) *UpdateStatement {
	us.where = append(us.where, condition)
	return us
}

// Returning adds a RETURNING or OUTPUT clause to the statement. Use it with
// PostgreSQL and SQL Server.
func (us *UpdateStatement) Returning(columns ...string) *UpdateStatement {
	us.returningColumns = append(us.returningColumns, columns...)
	return us
}

// Suffix adds an expression to suffix the statement. Use it to add a
// RETURNING clause with PostgreSQL (or whatever you need).
func (us *UpdateStatement) Suffix(suffix string) *UpdateStatement {
	us.suffixes = append(us.suffixes, suffix)
	return us
}

// approximateSetLength returns an approximation of final size of all set
// clauses.
func (us *UpdateStatement) approximateSetLength() int {
	// initialized with the count needed for "=" and ","
	length := 2 * len(us.sets)
	for _, s := range us.sets {
		// column or raw sql
		length += len(s.column)
		if s.value != nil {
			stringValue, isString := s.value.(string)
			if isString {
				length += len(stringValue)
			} else {
				// arbitrary
				length += 2
			}
		}
	}

	return length
}

// ToSQL returns a string with the SQL statement (containing placeholders),
// the arguments slices, and an error.
func (us *UpdateStatement) ToSQL() (string, []interface{}, error) {
	sqlWhereLength, argsWhereLength, err := sumOfConditionsLengths(us.where)
	if err != nil {
		return "", nil, err
	}

	sqlBuffer := newSQLBuffer(
		us.db.adapter,
		sqlWhereLength+us.approximateSetLength()+64,
		argsWhereLength,
	)

	sqlBuffer.Write("UPDATE ")
	sqlBuffer.Write(us.updateTable)
	sqlBuffer.writeSets(us.sets).
		writeReturningForPosition(us.returningColumns, adapters.ReturningSQLServer).
		writeWhere(us.where).
		writeReturningForPosition(us.returningColumns, adapters.ReturningPostgreSQL).
		writeStringsWithSpaces(us.suffixes)

	return sqlBuffer.SQL(), sqlBuffer.Arguments(), sqlBuffer.Err()
}

// Do executes the builded query, and return RowsAffected()
func (us *UpdateStatement) Do() (int64, error) {
	query, args, err := us.ToSQL()
	if err != nil {
		return 0, err
	}

	result, err := us.db.do(query, args)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return rowsAffected, err
}

// DoWithReturning executes the statement and fills the fields according to
// the columns in RETURNING clause.
func (us *UpdateStatement) DoWithReturning(record interface{}) (int64, error) {
	recordDescription, err := buildRecordDescription(record)
	if err != nil {
		return 0, err
	}

	// the function which will return the pointers according to the given columns
	f := func(record interface{}, columns []string) ([]interface{}, error) {
		pointers, err := recordDescription.structMapping.GetPointersForColumns(record, columns...)
		return pointers, err
	}

	return us.doWithReturning(recordDescription, f)
}

// DoWithReturning executes the statement and fills the fields according to
// the columns in RETURNING clause. It returns the count of rows returned.
func (us *UpdateStatement) doWithReturning(recordDescription *recordDescription, pointersGetter pointersGetter) (int64, error) {
	query, args, err := us.ToSQL()
	if err != nil {
		return 0, err
	}

	return us.db.doSelectOrWithReturning(query, args, recordDescription, pointersGetter)
}
