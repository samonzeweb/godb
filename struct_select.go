package godb

// StructSelect builds a SELECT statement for the given object.
//
// Example (book is a struct instance, books a slice) :
//
// 	 err := db.Select(&book).Where("id = ?", 123).Do()
//
// 	 err = db.Select(&books).Where("id > 1").Where("id < 10").Do()
type StructSelect struct {
	error             error
	selectStatement   *SelectStatement
	recordDescription *recordDescription
}

// Select initializes a SELECT statement with the given pointer as
// target. The pointer could point to a single instance or a slice.
func (db *DB) Select(record interface{}) *StructSelect {
	var err error

	ss := &StructSelect{}
	ss.recordDescription, err = buildRecordDescription(record)
	if err != nil {
		ss.error = err
		return ss
	}
	quotedTableName := db.adapter.Quote(ss.recordDescription.getTableName())
	ss.selectStatement = db.SelectFrom(quotedTableName)
	return ss
}

// Where adds a condition using string and arguments.
func (ss *StructSelect) Where(sql string, args ...interface{}) *StructSelect {
	if ss.error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.WhereQ(Q(sql, args...))
	return ss
}

// WhereQ adds a simple or complex predicate generated with Q and
// confunctions.
func (ss *StructSelect) WhereQ(condition *Condition) *StructSelect {
	if ss.error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.WhereQ(condition)
	return ss
}

// OrderBy adds an expression for the ORDER BY clause.
func (ss *StructSelect) OrderBy(orderBy string) *StructSelect {
	if ss.error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.OrderBy(orderBy)
	return ss
}

// Offset specifies the value for the OFFSET clause.
func (ss *StructSelect) Offset(offset int) *StructSelect {
	if ss.error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.Offset(offset)
	return ss
}

// Limit specifies the value for the LIMIT clause.
func (ss *StructSelect) Limit(limit int) *StructSelect {
	if ss.error != nil {
		return ss
	}
	ss.selectStatement = ss.selectStatement.Limit(limit)
	return ss
}

// Do executes the select statement, the record given to Select will contain
// the data.
func (ss *StructSelect) Do() error {
	if ss.error != nil {
		return ss.error
	}

	// Columns names
	allColumns := ss.recordDescription.structMapping.GetAllColumnsNames()
	ss.selectStatement = ss.selectStatement.Columns(ss.selectStatement.db.quoteAll(allColumns)...)

	f := func(record interface{}, columns []string) ([]interface{}, error) {
		pointers := ss.recordDescription.structMapping.GetAllFieldsPointers(record)
		return pointers, nil
	}

	return ss.selectStatement.do(ss.recordDescription, f)
}

// Count run the request with COUNT(*) and returns the count
func (ss *StructSelect) Count() (int64, error) {
	if ss.error != nil {
		return 0, ss.error
	}

	return ss.selectStatement.Count()
}
