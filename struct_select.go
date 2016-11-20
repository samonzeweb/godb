package godb

type structSelect struct {
	Error             error
	selectStatement   *selectStatement
	targetDescription *targetDescription
}

// Select initialise a SQL Select Statement with the given pointer as
// targer. The pointer could point to a single instance or a slice.
func (db *DB) Select(target interface{}) *structSelect {
	var err error

	ss := &structSelect{}
	ss.targetDescription, err = extractType(target)
	if err != nil {
		ss.Error = err
		return ss
	}
	ss.selectStatement = db.SelectFrom(ss.targetDescription.getTableName())
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

// OrderBy add an expression for the Order clause.
func (ss *structSelect) OrderBy(orderBy string) *structSelect {
	if ss.Error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.OrderBy(orderBy)
	return ss
}

// Offset specify the value for the Offset clause.
func (ss *structSelect) Offset(offset int) *structSelect {
	if ss.Error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.Offset(offset)
	return ss
}

// Limit specify the value for the Offset clause.
func (ss *structSelect) Limit(limit int) *structSelect {
	if ss.Error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.Limit(limit)
	return ss
}

// Do execute the select statement
func (ss *structSelect) Do() error {
	if ss.Error != nil {
		return ss.Error
	}

	// Columns names
	allColumns := ss.targetDescription.structMapping.GetAllColumnsNames()
	ss.selectStatement = ss.selectStatement.Columns(ss.selectStatement.db.quoteAll(allColumns)...)

	if ss.targetDescription.isSlice == false {
		// Only one row is requested
		ss.selectStatement.Limit(1)
	}

	f := func(target interface{}, columns []string) ([]interface{}, error) {
		pointers := ss.targetDescription.structMapping.GetAllFieldsPointers(target)
		return pointers, nil
	}

	return ss.selectStatement.do(ss.targetDescription, f)
}

// Count run the request with COUNT(*) and returns the count
func (ss *structSelect) Count() (int, error) {
	if ss.Error != nil {
		return 0, ss.Error
	}

	return ss.selectStatement.Count()
}
