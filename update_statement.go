package godb

import "time"

// updateStatement will contains all parts needed to build an UPDATE statement.
type updateStatement struct {
	db *DB

	updateTable string
	sets        []*setPart
	where       []*Condition
	suffixes    []string
}

// setPart contain elements for a single SET clause
// the value could be nil for a raw clause (ie count=count+1)
type setPart struct {
	// The column name, or the full SET clause for a raw clause
	column string
	// The value, or nil if it's a raw clause
	value interface{}
}

// UpdateTable create an updateStatement and specify table to update.
// It's the entry point to build an UPDATE query.
func (db *DB) UpdateTable(tableName string) *updateStatement {
	us := &updateStatement{db: db}
	us.updateTable = tableName
	return us
}

// Set add a part of SET clause to the query.
func (us *updateStatement) Set(column string, value interface{}) *updateStatement {
	setClause := &setPart{
		column: column,
		value:  value,
	}
	us.sets = append(us.sets, setClause)
	return us
}

// SetRaw add a raw SET clause to the query.
func (us *updateStatement) SetRaw(rawSQL string) *updateStatement {
	rawSetClause := &setPart{
		column: rawSQL,
		value:  nil,
	}
	us.sets = append(us.sets, rawSetClause)
	return us
}

// Where add a condition using string and arguments.
func (us *updateStatement) Where(sql string, args ...interface{}) *updateStatement {
	return us.WhereQ(Q(sql, args...))
}

// WhereQ add a simple or complex predicate generated with Q and
// confunctions.
func (us *updateStatement) WhereQ(condition *Condition) *updateStatement {
	us.where = append(us.where, condition)
	return us
}

// Suffix add an expression to suffix the statement.
func (us *updateStatement) Suffix(suffix string) *updateStatement {
	us.suffixes = append(us.suffixes, suffix)
	return us
}

// approximateSetLength returns an approximation of final size of all set
// clauses.
func (us *updateStatement) approximateSetLength() int {
	// initialise with the count needed for "=" and ","
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
	sql, args, err := us.ToSQL()
	if err != nil {
		return 0, err
	}
	sql = us.db.replacePlaceholders(sql)
	us.db.logPrintln("UPDATE : ", sql, args)

	// Execute the UPDATE statement
	startTime := time.Now()
	queryable, err := us.db.getQueryable(sql)
	if err != nil {
		return 0, err
	}
	result, err := queryable.Exec(args...)
	condumedTime := timeElapsedSince(startTime)
	us.db.addConsumedTime(condumedTime)
	us.db.logDuration(condumedTime)
	if err != nil {
		us.db.logPrintln("ERROR : ", err)
		return 0, err
	}

	// TODO : check if RowsAffected() is implemented by the driver
	rowsAffected, err := result.RowsAffected()
	return rowsAffected, err
}
