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

// InsertInto initialise a INSERT INTO statement builder
func (db *DB) InsertInto(tableName string) *insertStatement {
	ip := &insertStatement{db: db}
	ip.intoTable = tableName
	return ip
}

// Columns add columns to insert.
func (is *insertStatement) Columns(columns ...string) *insertStatement {
	is.columns = append(is.columns, columns...)
	return is
}

// Values to insert
func (is *insertStatement) Values(values ...interface{}) *insertStatement {
	is.values = append(is.values, values)
	return is
}

// Suffix add an expression to suffix the statement.
func (is *insertStatement) Suffix(suffix string) *insertStatement {
	is.suffixes = append(is.suffixes, suffix)
	return is
}

// ToSQL returns a string with the SQL statement (containing placeholders),
// the arguments slices, and an error.
func (is *insertStatement) ToSQL() (string, []interface{}, error) {

	// TODO : estimate the buffer size !!!
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

// Do executes the builded INSERT statement and return the LastInsertId() if
// the adapter does not implements InsertReturningSuffixer.
func (si *insertStatement) Do() (int64, error) {
	sql, args, err := si.ToSQL()
	if err != nil {
		return 0, err
	}
	sql = si.db.replacePlaceholders(sql)
	si.db.logPrintln("INSERT : ", sql, args)

	// Execute the INSERT statement
	startTime := time.Now()
	queryable, err := si.db.getQueryable(sql)
	if err != nil {
		return 0, err
	}
	result, err := queryable.Exec(args...)
	condumedTime := timeElapsedSince(startTime)
	si.db.addConsumedTime(condumedTime)
	si.db.logDuration(condumedTime)
	if err != nil {
		si.db.logPrintln("ERROR : ", err)
		return 0, err
	}

	// Return the created 'Id' (if available)
	_, ok := si.db.adapter.(adapters.InsertReturningSuffixer)
	if ok {
		// adapters with InsertSuffixer does not use LastInsertId()
		return 0, nil
	}
	lastInsertId, err := result.LastInsertId()
	return lastInsertId, err
}

func (si *insertStatement) doWithReturning(record interface{}) error {
	recordDescription, err := buildRecordDescription(record)
	if err != nil {
		return err
	}

	sql, args, err := si.ToSQL()
	if err != nil {
		return err
	}
	sql = si.db.replacePlaceholders(sql)
	si.db.logPrintln("INSERT : ", sql, args)

	// Execute the INSERT statement
	startTime := time.Now()
	pointers, err := recordDescription.structMapping.GetAutoFieldsPointers(record)
	if err != nil {
		return err
	}
	queryable, err := si.db.getQueryable(sql)
	if err != nil {
		return err
	}
	err = queryable.QueryRow(args...).Scan(pointers...)
	condumedTime := timeElapsedSince(startTime)
	si.db.addConsumedTime(condumedTime)
	si.db.logDuration(condumedTime)
	if err != nil {
		si.db.logPrintln("ERROR : ", err)
		return err
	}

	return nil
}
