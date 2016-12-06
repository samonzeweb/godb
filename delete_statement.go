package godb

// deleteStatement is a DELETE sql statement builder.
// Initialize it with the DeleteFrom function.
type deleteStatement struct {
	db *DB

	fromTable string
	where     []*Condition
	suffixes  []string
}

// DeleteFrom initializes a DELETE statement builder.
func (db *DB) DeleteFrom(tableName string) *deleteStatement {
	ds := &deleteStatement{db: db}
	ds.fromTable = tableName
	return ds
}

// Where adds a condition using string and arguments.
func (ds *deleteStatement) Where(sql string, args ...interface{}) *deleteStatement {
	return ds.WhereQ(Q(sql, args...))
}

// WhereQ adds a simple or complex predicate generated with Q and
// confunctions.
func (ds *deleteStatement) WhereQ(condition *Condition) *deleteStatement {
	ds.where = append(ds.where, condition)
	return ds
}

// Suffix adds an expression to suffix the statement.
func (ds *deleteStatement) Suffix(suffix string) *deleteStatement {
	ds.suffixes = append(ds.suffixes, suffix)
	return ds
}

// ToSQL returns a string with the SQL statement (containing placeholders),
// the arguments slices, and an error.
func (ds *deleteStatement) ToSQL() (string, []interface{}, error) {
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

// Do executes the builded query, and return RowsAffected()
func (ds *deleteStatement) Do() (int64, error) {
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
func (ds *deleteStatement) DoWithReturning(record interface{}) error {
	recordDescription, err := buildRecordDescription(record)
	if err != nil {
		return err
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
func (ds *deleteStatement) doWithReturning(recordDescription *recordDescription, pointersGetter pointersGetter) error {
	query, args, err := ds.ToSQL()
	if err != nil {
		return err
	}

	return ds.db.doWithReturning(query, args, recordDescription, pointersGetter)
}
