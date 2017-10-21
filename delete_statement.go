package godb

// DeleteStatement is a DELETE sql statement builder.
// Initialize it with the DeleteFrom method.
//
// Example :
// 	count, err := db.DeleteFrom("bar").Where("foo = 1").Do()
type DeleteStatement struct {
	db *DB

	fromTable string
	where     []*Condition
	suffixes  []string
}

// DeleteFrom initializes a DELETE statement builder.
func (db *DB) DeleteFrom(tableName string) *DeleteStatement {
	ds := &DeleteStatement{db: db}
	ds.fromTable = tableName
	return ds
}

// Where adds a condition using string and arguments.
func (ds *DeleteStatement) Where(sql string, args ...interface{}) *DeleteStatement {
	return ds.WhereQ(Q(sql, args...))
}

// WhereQ adds a simple or complex predicate generated with Q and
// confunctions.
func (ds *DeleteStatement) WhereQ(condition *Condition) *DeleteStatement {
	ds.where = append(ds.where, condition)
	return ds
}

// Suffix adds an expression to suffix the statement. Use it to add a
// RETURNING clause with PostgreSQL (or whatever you need).
func (ds *DeleteStatement) Suffix(suffix string) *DeleteStatement {
	ds.suffixes = append(ds.suffixes, suffix)
	return ds
}

// ToSQL returns a string with the SQL statement (containing placeholders),
// the arguments slices, and an error.
func (ds *DeleteStatement) ToSQL() (string, []interface{}, error) {
	sqlWhereLength, argsWhereLength, err := sumOfConditionsLengths(ds.where)
	if err != nil {
		return "", nil, err
	}

	sqlBuffer := newSQLBuffer(
		ds.db.adapter,
		sqlWhereLength+64,
		argsWhereLength,
	)

	sqlBuffer.write("DELETE")

	if err := sqlBuffer.writeFrom(ds.fromTable); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeWhere(ds.where); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeStrings(ds.suffixes); err != nil {
		return "", nil, err
	}

	return sqlBuffer.sqlString(), sqlBuffer.sqlArguments(), nil
}

// Do executes the builded query, and return thr rows affected count.
func (ds *DeleteStatement) Do() (int64, error) {
	query, args, err := ds.ToSQL()
	if err != nil {
		return 0, err
	}

	result, err := ds.db.do(query, args)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return rowsAffected, err
}

// DoWithReturning executes the statement and fills the fields according to
// the columns in RETURNING clause.
func (ds *DeleteStatement) DoWithReturning(record interface{}) (int64, error) {
	recordDescription, err := buildRecordDescription(record)
	if err != nil {
		return 0, err
	}

	// the function which will return the pointers according to the given columns
	f := func(record interface{}, columns []string) ([]interface{}, error) {
		pointers, err := recordDescription.structMapping.GetPointersForColumns(record, columns...)
		return pointers, err
	}

	return ds.doWithReturning(recordDescription, f)
}

// DoWithReturning executes the statement and fills the fields according to
// the columns in RETURNING clause.
func (ds *DeleteStatement) doWithReturning(recordDescription *recordDescription, pointersGetter pointersGetter) (int64, error) {
	query, args, err := ds.ToSQL()
	if err != nil {
		return 0, err
	}

	return ds.db.doWithReturning(query, args, recordDescription, pointersGetter)
}
