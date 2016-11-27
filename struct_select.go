package godb

// structSelect build a SELECT statement for the given object
type structSelect struct {
	Error             error
	selectStatement   *selectStatement
	recordDescription *recordDescription
}

// Select initialise a SQL Select Statement with the given pointer as
// target. The pointer could point to a single instance or a slice.
func (db *DB) Select(record interface{}) *structSelect {
	var err error

	ss := &structSelect{}
	ss.recordDescription, err = buildRecordDescription(record)
	if err != nil {
		ss.Error = err
		return ss
	}
	quotedTableName := db.adapter.Quote(ss.recordDescription.getTableName())
	ss.selectStatement = db.SelectFrom(quotedTableName)
	return ss
}

// Where add a condition using string and arguments.
func (ss *structSelect) Where(sql string, args ...interface{}) *structSelect {
	if ss.Error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.WhereQ(Q(sql, args...))
	return ss
}

// WhereQ add a simple or complex predicate generated with Q and
// confunctions.
func (ss *structSelect) WhereQ(condition *Condition) *structSelect {
	if ss.Error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.WhereQ(condition)
	return ss
}

// OrderBy add an expression for the ORDER BY clause.
func (ss *structSelect) OrderBy(orderBy string) *structSelect {
	if ss.Error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.OrderBy(orderBy)
	return ss
}

// Offset specify the value for the OFFSET clause.
func (ss *structSelect) Offset(offset int) *structSelect {
	if ss.Error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.Offset(offset)
	return ss
}

// Limit specify the value for the LIMIT clause.
func (ss *structSelect) Limit(limit int) *structSelect {
	if ss.Error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.Limit(limit)
	return ss
}

// Do execute the select statement, the record given to Select will contain
// the data.
func (ss *structSelect) Do() error {
	if ss.Error != nil {
		return ss.Error
	}

	// Columns names
	allColumns := ss.recordDescription.structMapping.GetAllColumnsNames()
	ss.selectStatement = ss.selectStatement.Columns(ss.selectStatement.db.quoteAll(allColumns)...)

	if ss.recordDescription.isSlice == false {
		// Only one row is requested
		ss.selectStatement.Limit(1)
	}

	f := func(record interface{}, columns []string) ([]interface{}, error) {
		pointers := ss.recordDescription.structMapping.GetAllFieldsPointers(record)
		return pointers, nil
	}

	return ss.selectStatement.do(ss.recordDescription, f)
}

// Count run the request with COUNT(*) and returns the count
func (ss *structSelect) Count() (int, error) {
	if ss.Error != nil {
		return 0, ss.Error
	}

	return ss.selectStatement.Count()
}
