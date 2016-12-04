package godb

import (
	"time"

	"gitlab.com/samonzeweb/godb/adapters"
)

// insertStatement is an INSERT statement builder.
type insertStatement struct {
	db *DB

	columns   []string
	intoTable string
	values    [][]interface{}
	suffixes  []string
}

// InsertInto initializes a INSERT statement builder
func (db *DB) InsertInto(tableName string) *insertStatement {
	ip := &insertStatement{db: db}
	ip.intoTable = tableName
	return ip
}

// Columns adds columns to insert.
func (is *insertStatement) Columns(columns ...string) *insertStatement {
	is.columns = append(is.columns, columns...)
	return is
}

// Values add values to insert.
func (is *insertStatement) Values(values ...interface{}) *insertStatement {
	is.values = append(is.values, values)
	return is
}

// Suffix adds an expression to suffix the statement.
func (is *insertStatement) Suffix(suffix string) *insertStatement {
	is.suffixes = append(is.suffixes, suffix)
	return is
}

// ToSQL returns a string with the SQL statement (containing placeholders),
// the arguments slices, and an error.
func (is *insertStatement) ToSQL() (string, []interface{}, error) {

	// TODO : estimate the buffer size.
	sqlBuffer := newSQLBuffer(is.db.adapter, 256, 16)

	sqlBuffer.write("INSERT ")

	if err := sqlBuffer.writeInto(is.intoTable); err != nil {
		return "", nil, err
	}

	sqlBuffer.write(" (")
	if err := sqlBuffer.writeColumns(is.columns); err != nil {
		return "", nil, err
	}
	sqlBuffer.write(") VALUES ")

	if err := sqlBuffer.writeInsertValues(is.values, len(is.columns)); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeStrings(is.suffixes); err != nil {
		return "", nil, err
	}

	return sqlBuffer.sqlString(), sqlBuffer.sqlArguments(), nil
}

// Do executes the builded INSERT statement and returns the LastInsertId() if
// the adapter does not implement InsertReturningSuffixer.
func (si *insertStatement) Do() (int64, error) {
	query, args, err := si.ToSQL()
	if err != nil {
		return 0, err
	}

	result, err := si.db.do(query, args)

	// Return the created 'Id' (if available)
	_, ok := si.db.adapter.(adapters.InsertReturningSuffixer)
	if ok {
		// adapters with InsertSuffixer does not use LastInsertId()
		return 0, nil
	}
	lastInsertId, err := result.LastInsertId()
	return lastInsertId, err
}

// DoWithReturning executes the statement and fills the fields according to
// the columns in RETURNING clause.
func (si *insertStatement) DoWithReturning(record interface{}) error {
	recordDescription, err := buildRecordDescription(record)
	if err != nil {
		return err
	}

	// the function which will return the pointers according to the given columns
	f := func(record interface{}, columns []string) ([]interface{}, error) {
		pointers, err := recordDescription.structMapping.GetPointersForColumns(record, columns...)
		return pointers, err
	}

	return si.doWithReturning(recordDescription, f)
}

// doWithReturning executes the statement and fills the auto fields.
// It is called when the adapter implements InsertReturningSuffixer.
func (si *insertStatement) doWithReturning(recordDescription *recordDescription, pointersGetter pointersGetter) error {
	sql, args, err := si.ToSQL()
	if err != nil {
		return err
	}
	sql = si.db.replacePlaceholders(sql)
	si.db.logPrintln("INSERT : ", sql, args)

	startTime := time.Now()
	queryable, err := si.db.getQueryable(sql)
	if err != nil {
		return err
	}
	rows, err := queryable.Query(args...)
	condumedTime := timeElapsedSince(startTime)
	si.db.addConsumedTime(condumedTime)
	si.db.logDuration(condumedTime)
	if err != nil {
		si.db.logPrintln("ERROR : ", err)
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		si.db.logPrintln("ERROR : ", err)
		return err
	}

	index := 0
	for rows.Next() {
		instancePtr := recordDescription.index(index)
		index++
		pointers, innererr := pointersGetter(instancePtr, columns)
		if innererr != nil {
			return innererr
		}
		innererr = rows.Scan(pointers...)
		if innererr != nil {
			si.db.logPrintln("ERROR : ", innererr)
			return innererr
		}
	}

	err = rows.Err()
	if err != nil {
		si.db.logPrintln("ERROR : ", err)
	}
	return err
}
