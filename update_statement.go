package godb

// updateStatement will contains all parts needed to build an UPDATE statement.
type updateStatement struct {
	db *DB

	updateTable string
	sets        []*setPart
	where       []*Condition
	suffixes    []string
}

// setPart contains elements for a single SET clause.
// The value could be nil for a raw clause (ie count=count+1)
type setPart struct {
	// The column name, or the full SET clause for a raw clause
	column string
	// The value, or nil if it's a raw clause
	value interface{}
}

// UpdateTable creates an updateStatement and specify table to update.
// It's the entry point to build an UPDATE query.
func (db *DB) UpdateTable(tableName string) *updateStatement {
	us := &updateStatement{db: db}
	us.updateTable = tableName
	return us
}

// Set adds a part of SET clause to the query.
func (us *updateStatement) Set(column string, value interface{}) *updateStatement {
	setClause := &setPart{
		column: column,
		value:  value,
	}
	us.sets = append(us.sets, setClause)
	return us
}

// SetRaw adds a raw SET clause to the query.
func (us *updateStatement) SetRaw(rawSQL string) *updateStatement {
	rawSetClause := &setPart{
		column: rawSQL,
		value:  nil,
	}
	us.sets = append(us.sets, rawSetClause)
	return us
}

// Where adds a condition using string and arguments.
func (us *updateStatement) Where(sql string, args ...interface{}) *updateStatement {
	return us.WhereQ(Q(sql, args...))
}

// WhereQ adds a simple or complex predicate generated with Q and
// confunctions.
func (us *updateStatement) WhereQ(condition *Condition) *updateStatement {
	us.where = append(us.where, condition)
	return us
}

// Suffix adds an expression to suffix the statement.
func (us *updateStatement) Suffix(suffix string) *updateStatement {
	us.suffixes = append(us.suffixes, suffix)
	return us
}

// approximateSetLength returns an approximation of final size of all set
// clauses.
func (us *updateStatement) approximateSetLength() int {
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
func (us *updateStatement) ToSQL() (string, []interface{}, error) {
	sqlWhereLength, argsWhereLength, err := sumOfConditionsLengths(us.where)
	if err != nil {
		return "", nil, err
	}

	sqlBuffer := newSQLBuffer(
		us.db.adapter,
		sqlWhereLength+us.approximateSetLength()+64,
		argsWhereLength,
	)

	sqlBuffer.write("UPDATE ")

	if err := sqlBuffer.write(us.updateTable); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeSets(us.sets); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeWhere(us.where); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeStrings(us.suffixes); err != nil {
		return "", nil, err
	}

	return sqlBuffer.sqlString(), sqlBuffer.sqlArguments(), nil
}

// Do executes the builded query, and return RowsAffected()
func (us *updateStatement) Do() (int64, error) {
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
func (us *updateStatement) DoWithReturning(record interface{}) error {
	recordDescription, err := buildRecordDescription(record)
	if err != nil {
		return err
	}

	// the function which will return the pointers according to the given columns
	f := func(record interface{}, columns []string) ([]interface{}, error) {
		pointers, err := recordDescription.structMapping.GetPointersForColumns(record, columns...)
		return pointers, err
	}

	return us.doWithReturning(recordDescription, f)
}

// DoWithReturning executes the statement and fills the fields according to
// the columns in RETURNING clause.
func (us *updateStatement) doWithReturning(recordDescription *recordDescription, pointersGetter pointersGetter) error {
	query, args, err := us.ToSQL()
	if err != nil {
		return err
	}

	return us.db.doWithReturning(query, args, recordDescription, pointersGetter)
}
