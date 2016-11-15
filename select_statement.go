package godb

//
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

//
type joinPart struct {
	joinType  string
	tableName string
	as        string
	on        *Condition
}

// SelectFrom initialise a select statement builder
func (db *DB) SelectFrom(tableName string) *selectStatement {
	return newSelectStatement(db, tableName)
}

// Create a new select statement
func newSelectStatement(db *DB, tableName string) *selectStatement {
	ss := &selectStatement{db: db}
	return ss.From(tableName)
}

// From add table to the SELECT statement. It can be called multiple times.
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

// Join add a generic join clause, wich will be inserted between FROM and Where
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
// confunctions.
func (ss *selectStatement) WhereQ(condition *Condition) *selectStatement {
	ss.where = append(ss.where, condition)
	return ss
}

// GroupBy add a generic join clause, wich will be inserted between FROM and Where
// clauses.
func (ss *selectStatement) GroupBy(groupBy string) *selectStatement {
	ss.groupBy = append(ss.groupBy, groupBy)
	return ss
}

// Having add a condition using string and arguments.
func (ss *selectStatement) Having(sql string, args ...interface{}) *selectStatement {
	return ss.HavingQ(Q(sql, args...))
}

// HavingQ add a simple or complex predicate generated with Q and
// conjunctions.
func (ss *selectStatement) HavingQ(condition *Condition) *selectStatement {
	ss.having = append(ss.having, condition)
	return ss
}

// OrderBy add an expression for the Order clause.
func (ss *selectStatement) OrderBy(orderBy string) *selectStatement {
	ss.orderBy = append(ss.orderBy, orderBy)
	return ss
}

// Offset specify the value for the Offset clause.
func (ss *selectStatement) Offset(offset int) *selectStatement {
	ss.offset = new(int)
	*ss.offset = offset
	return ss
}

// Limit specify the value for the Offset clause.
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
// The target argument has to be a pointer to a struct or a slice
func (ss *selectStatement) Do(target interface{}) error {
	targetInfo, err := extractType(target)
	if err != nil {
		return err
	}

	if targetInfo.IsSlice == false {
		// Only one row is requested
		ss.Limit(1)
	}

	return ss.do(targetInfo)
}

// do executes the statement and fill the struct or slice
func (ss *selectStatement) do(targetInfo *targetDescription) error {
	sql, args, err := ss.ToSQL()
	if err != nil {
		return err
	}

	rows, err := ss.db.sqlDB.Query(sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		err = targetInfo.fillTarget(
			// Fill one instance with one row
			func(target interface{}) error {
				fieldsPointers, innererr := targetInfo.StructMapping.GetPointersForColumns(target, columns...)
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
			return err
		}
	}

	return rows.Err()
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

	var count int
	err = ss.db.sqlDB.QueryRow(sql, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
