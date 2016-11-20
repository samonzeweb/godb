package godb

import "time"

// selectStatement is a SELECT sql statement builder
type selectStatement struct {
	db *DB

	distinct   bool
	columns    []string
	fromTables []string
	joins      []*joinPart
	where      []*Condition
	groupBy    []string
	having     []*Condition
	orderBy    []string
	limit      *int
	offset     *int
	suffixes   []string
}

// joinPart describe a sql JOIN clause
type joinPart struct {
	joinType  string
	tableName string
	as        string
	on        *Condition
}

// pointersGetter is a func type, returning a list of pointers (and error) for
// a given instance pointer and a columns names list.
type pointersGetter func(record interface{}, columns []string) ([]interface{}, error)

// SelectFrom initialise a SELECT statement builder
func (db *DB) SelectFrom(tableName string) *selectStatement {
	ss := &selectStatement{db: db}
	return ss.From(tableName)
}

// From add table to the select statement. It can be called multiple times.
func (ss *selectStatement) From(tableName string) *selectStatement {
	ss.fromTables = append(ss.fromTables, tableName)
	return ss
}

// Columns add columns to select.
func (ss *selectStatement) Columns(columns ...string) *selectStatement {
	ss.columns = append(ss.columns, columns...)
	return ss
}

// Distinct will add DISTINCT keyword the the generated statement.
func (ss *selectStatement) Distinct() *selectStatement {
	ss.distinct = true
	return ss
}

// LeftJoin add a LEFT JOIN clause, wich will be inserted between FROM and WHERE
// clauses.
func (ss *selectStatement) LeftJoin(tableName string, as string, on *Condition) *selectStatement {
	join := &joinPart{
		joinType:  "LEFT JOIN",
		tableName: tableName,
		as:        as,
		on:        on,
	}
	ss.joins = append(ss.joins, join)
	return ss
}

// Where add a condition using string and arguments.
func (ss *selectStatement) Where(sql string, args ...interface{}) *selectStatement {
	return ss.WhereQ(Q(sql, args...))
}

// WhereQ add a simple or complex predicate generated with Q and
// conjunctions.
func (ss *selectStatement) WhereQ(condition *Condition) *selectStatement {
	ss.where = append(ss.where, condition)
	return ss
}

// GroupBy add a GROUP BY clause.
func (ss *selectStatement) GroupBy(groupBy string) *selectStatement {
	ss.groupBy = append(ss.groupBy, groupBy)
	return ss
}

// Having add a HAVING clause with a condition build with a sql string and
// its arguments (like Where).
func (ss *selectStatement) Having(sql string, args ...interface{}) *selectStatement {
	return ss.HavingQ(Q(sql, args...))
}

// HavingQ add a simple or complex predicate generated with Q and
// conjunctions (like WhereQ)
func (ss *selectStatement) HavingQ(condition *Condition) *selectStatement {
	ss.having = append(ss.having, condition)
	return ss
}

// OrderBy add an expression for the ORDER BY clause.
func (ss *selectStatement) OrderBy(orderBy string) *selectStatement {
	ss.orderBy = append(ss.orderBy, orderBy)
	return ss
}

// Offset specify the value for the OFFSET clause.
func (ss *selectStatement) Offset(offset int) *selectStatement {
	ss.offset = new(int)
	*ss.offset = offset
	return ss
}

// Limit specify the value for the LIMIT clause.
func (ss *selectStatement) Limit(limit int) *selectStatement {
	ss.limit = new(int)
	*ss.limit = limit
	return ss
}

// Suffix add an expression to suffix the query.
func (ss *selectStatement) Suffix(suffix string) *selectStatement {
	ss.suffixes = append(ss.suffixes, suffix)
	return ss
}

// ToSQL returns a string with the SQL request (containing placeholders),
// the arguments slices, and an error.
func (ss *selectStatement) ToSQL() (string, []interface{}, error) {
	sqlWhereLength, argsWhereLength, err := sumOfConditionsLengths(ss.where)
	if err != nil {
		return "", nil, err
	}

	sqlHavingLength, argsHavingLength, err := sumOfConditionsLengths(ss.having)
	if err != nil {
		return "", nil, err
	}

	sqlBuffer := newSQLBuffer(
		sqlWhereLength+sqlHavingLength+64,
		argsWhereLength+argsHavingLength+4,
	)

	sqlBuffer.write("SELECT ")

	if ss.distinct == true {
		sqlBuffer.write("DISTINCT ")
	}

	if err := sqlBuffer.writeColumns(ss.columns); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeFrom(ss.fromTables...); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeJoins(ss.joins); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeWhere(ss.where); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeGroupByAndHaving(ss.groupBy, ss.having); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeOrderBy(ss.orderBy); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeOffset(ss.offset); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeLimit(ss.limit); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeStrings(ss.suffixes); err != nil {
		return "", nil, err
	}

	return sqlBuffer.sqlString(), sqlBuffer.sqlArguments(), nil
}

// Do execute the select statement
// The record argument has to be a pointer to a struct or a slice
func (ss *selectStatement) Do(record interface{}) error {
	recordInfo, err := buildRecordDescription(record)
	if err != nil {
		return err
	}

	if recordInfo.isSlice == false {
		// Only one row is requested
		ss.Limit(1)
	}

	// the function wich will return the pointers according to the given columns
	f := func(record interface{}, columns []string) ([]interface{}, error) {
		pointers, err := recordInfo.structMapping.GetPointersForColumns(record, columns...)
		return pointers, err
	}

	return ss.do(recordInfo, f)
}

// do executes the statement and fill the struct or slice
func (ss *selectStatement) do(recordInfo *recordDescription, pointersGetter pointersGetter) error {
	sql, args, err := ss.ToSQL()
	if err != nil {
		return err
	}
	ss.db.logPrintln("SELECT : ", sql, args)

	startTime := time.Now()
	rows, err := ss.db.getTxElseDb().Query(sql, args...)
	condumedTime := timeElapsedSince(startTime)
	ss.db.addConsumedTime(condumedTime)
	ss.db.logDuration(condumedTime)
	if err != nil {
		ss.db.logPrintln("ERROR : ", err)
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		ss.db.logPrintln("ERROR : ", err)
		return err
	}

	for rows.Next() {
		err = recordInfo.fillRecord(
			// Fill one instance with one row
			func(record interface{}) error {
				fieldsPointers, innererr := pointersGetter(record, columns)
				if innererr != nil {
					return innererr
				}
				innererr = rows.Scan(fieldsPointers...)
				if err != nil {
					return innererr
				}
				return nil
			})

		if err != nil {
			ss.db.logPrintln("ERROR : ", err)
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		ss.db.logPrintln("ERROR : ", err)
	}
	return err
}

// Count run the request with COUNT(*) (remove others columns)
// and returns the count
func (ss *selectStatement) Count() (int, error) {
	ss.columns = ss.columns[:0]
	ss.Columns("COUNT(*)")

	sql, args, err := ss.ToSQL()
	if err != nil {
		return 0, err
	}
	ss.db.logPrintln("SELECT : ", sql, args)

	var count int
	startTime := time.Now()
	err = ss.db.getTxElseDb().QueryRow(sql, args...).Scan(&count)
	condumedTime := timeElapsedSince(startTime)
	ss.db.addConsumedTime(condumedTime)
	ss.db.logDuration(condumedTime)
	if err != nil {
		ss.db.logPrintln("ERROR : ", err)
		return 0, err
	}

	return count, nil
}
