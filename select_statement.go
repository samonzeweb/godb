package godb

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/samonzeweb/godb/adapters"
)

// SelectStatement is a SELECT sql statement builder.
// Initialize it with the SelectFrom method.
//
// Examples :
// 	err := db.SelecFrom("bar").
// 		Columns("foo", "baz").
// 		Where("foo > 2").
// 		Do(&target)
type SelectStatement struct {
	db    *DB
	error error

	distinct             bool
	columns              []string
	areColumnsFromStruct bool
	columnAliases        map[string]string
	fromTables           []string
	joins                []*joinPart
	where                []*Condition
	groupBy              []string
	having               []*Condition
	orderBy              []string
	limit                *int
	offset               *int
	suffixes             []string
}

// joinPart describes a sql JOIN clause.
type joinPart struct {
	joinType  string
	tableName string
	as        string
	on        *Condition
}

// SelectFrom initializes a SELECT statement builder.
func (db *DB) SelectFrom(tableNames ...string) *SelectStatement {
	ss := &SelectStatement{db: db, columnAliases: map[string]string{}}
	return ss.From(tableNames...)
}

// From adds table to the select statement. It can be called multiple times.
func (ss *SelectStatement) From(tableNames ...string) *SelectStatement {
	ss.fromTables = append(ss.fromTables, tableNames...)
	return ss
}

// Columns adds columns to select. Multple calls of columns are allowed.
func (ss *SelectStatement) Columns(columns ...string) *SelectStatement {
	if ss.areColumnsFromStruct {
		ss.error = fmt.Errorf("You can't mix Columns and ColumnsFromStruct to build a select query")
		return ss
	}

	ss.columns = append(ss.columns, columns...)
	return ss
}

// ColumnsFromStruct adds columns to select, extrating them from the
// given struct (or slice of struct). Always use a pointer as argument.
// You can't mix the use of ColumnsFromStruct and Columns methods.
func (ss *SelectStatement) ColumnsFromStruct(record interface{}) *SelectStatement {
	if len(ss.columns) > 0 {
		ss.error = fmt.Errorf("You can't mix Columns and ColumnsFromStruct to build a select query")
		return ss
	}
	ss.areColumnsFromStruct = true

	recordInfo, err := buildRecordDescription(record)
	if err != nil {
		ss.error = err
	} else {
		columns := ss.db.quoteAll(recordInfo.structMapping.GetAllColumnsNames())
		ss.columns = append(ss.columns, columns...)
	}

	return ss
}

// ColumnAlias allows to define alias for a column. Useful if selectable
// columns are built with ColumnsFromStruct and when using joins.
func (ss *SelectStatement) ColumnAlias(column, alias string) *SelectStatement {
	quoted := strings.Split(column, ".")
	for i := range quoted {
		quoted[i] = ss.db.adapter.Quote(quoted[i])
	}
	ss.columnAliases[ss.db.adapter.Quote(alias)] = strings.Join(quoted, ".")
	return ss
}

// Distinct adds DISTINCT keyword the the generated statement.
func (ss *SelectStatement) Distinct() *SelectStatement {
	ss.distinct = true
	return ss
}

// InnerJoin adds as INNER JOIN clause, which will be inserted between FROM and WHERE
// clauses.
func (ss *SelectStatement) InnerJoin(tableName string, as string, on *Condition) *SelectStatement {
	return ss.addJoin("INNER JOIN", tableName, as, on)
}

// LeftJoin adds a LEFT JOIN clause, which will be inserted between FROM and WHERE
// clauses.
func (ss *SelectStatement) LeftJoin(tableName string, as string, on *Condition) *SelectStatement {
	return ss.addJoin("LEFT JOIN", tableName, as, on)
}

// addJoin adds a join clause.
func (ss *SelectStatement) addJoin(joinType string, tableName string, as string, on *Condition) *SelectStatement {
	join := &joinPart{
		joinType:  joinType,
		tableName: tableName,
		as:        as,
		on:        on,
	}
	ss.joins = append(ss.joins, join)
	return ss
}

// Where adds a condition using string and arguments.
func (ss *SelectStatement) Where(sql string, args ...interface{}) *SelectStatement {
	return ss.WhereQ(Q(sql, args...))
}

// WhereQ adds a simple or complex predicate generated with Q and
// conjunctions.
func (ss *SelectStatement) WhereQ(condition *Condition) *SelectStatement {
	ss.where = append(ss.where, condition)
	return ss
}

// GroupBy adds a GROUP BY clause. You can call GroupBy multiple times.
func (ss *SelectStatement) GroupBy(groupBy string) *SelectStatement {
	ss.groupBy = append(ss.groupBy, groupBy)
	return ss
}

// Having adds a HAVING clause with a condition build with a sql string and
// its arguments (like Where).
func (ss *SelectStatement) Having(sql string, args ...interface{}) *SelectStatement {
	return ss.HavingQ(Q(sql, args...))
}

// HavingQ adds a simple or complex predicate generated with Q and
// conjunctions (like WhereQ).
func (ss *SelectStatement) HavingQ(condition *Condition) *SelectStatement {
	ss.having = append(ss.having, condition)
	return ss
}

// OrderBy adds an expression for the ORDER BY clause.
// You can call GroupBy multiple times.
func (ss *SelectStatement) OrderBy(orderBy string) *SelectStatement {
	ss.orderBy = append(ss.orderBy, orderBy)
	return ss
}

// Offset specifies the value for the OFFSET clause.
func (ss *SelectStatement) Offset(offset int) *SelectStatement {
	ss.offset = new(int)
	*ss.offset = offset
	return ss
}

// Limit specifies the value for the LIMIT clause.
func (ss *SelectStatement) Limit(limit int) *SelectStatement {
	ss.limit = new(int)
	*ss.limit = limit
	return ss
}

// Suffix adds an expression to suffix the query.
func (ss *SelectStatement) Suffix(suffix string) *SelectStatement {
	ss.suffixes = append(ss.suffixes, suffix)
	return ss
}

// ToSQL returns a string with the SQL request (containing placeholders),
// the arguments slices, and an error.
func (ss *SelectStatement) ToSQL() (string, []interface{}, error) {
	if ss.error != nil {
		return "", nil, ss.error
	}

	sqlWhereLength, argsWhereLength, err := sumOfConditionsLengths(ss.where)
	if err != nil {
		return "", nil, err
	}

	sqlHavingLength, argsHavingLength, err := sumOfConditionsLengths(ss.having)
	if err != nil {
		return "", nil, err
	}

	sqlBuffer := newSQLBuffer(
		ss.db.adapter,
		sqlWhereLength+sqlHavingLength+64,
		argsWhereLength+argsHavingLength+4,
	)

	sqlBuffer.Write("SELECT ")

	if ss.distinct == true {
		sqlBuffer.Write("DISTINCT ")
	}

	sqlBuffer.writeColumns(ss.columns).
		writeFrom(ss.fromTables...).
		writeJoins(ss.joins).
		writeWhere(ss.where).
		writeGroupByAndHaving(ss.groupBy, ss.having).
		writeOrderBy(ss.orderBy)

	offsetFirst := false
	if limitOffsetOrderer, ok := ss.db.adapter.(adapters.LimitOffsetOrderer); ok {
		offsetFirst = limitOffsetOrderer.IsOffsetFirst()
	}
	if offsetFirst {
		// Offset is before limit
		sqlBuffer.writeOffset(ss.offset).
			writeLimit(ss.limit)
	} else {
		// Limit is before offset (default case)
		sqlBuffer.writeLimit(ss.limit).
			writeOffset(ss.offset)
	}

	sqlBuffer.writeStringsWithSpaces(ss.suffixes)

	return sqlBuffer.SQL(), sqlBuffer.Arguments(), sqlBuffer.Err()
}

// Do executes the select statement.
// The record argument has to be a pointer to a struct or a slice.
// If no columns is defined for current select statement, all columns are
// added from record parameter's struct.
// If the argument is not a slice, a row is expected, and Do returns
// sql.ErrNoRows is none where found.
func (ss *SelectStatement) Do(record interface{}) error {
	if ss.error != nil {
		return ss.error
	}

	recordInfo, err := buildRecordDescription(record)
	if err != nil {
		return err
	}
	// If no columns defined for selection, get all columns (SELECT * FROM)
	if len(ss.columns) == 0 {
		ss.areColumnsFromStruct = true
		columns := ss.db.quoteAll(recordInfo.structMapping.GetAllColumnsNames())
		ss.columns = append(ss.columns, columns...)
	}

	// Replace columns with aliases
	for i := range ss.columns {
		if c, ok := ss.columnAliases[ss.columns[i]]; ok {
			ss.columns[i] = fmt.Sprintf("%s as %s", c, ss.columns[i])
		}
	}

	// the function which will return the pointers according to the given columns
	f := func(record interface{}, columns []string) ([]interface{}, error) {
		var pointers []interface{}
		var err error
		if ss.areColumnsFromStruct {
			pointers = recordInfo.structMapping.GetAllFieldsPointers(record)
		} else {
			pointers, err = recordInfo.structMapping.GetPointersForColumns(record, columns...)
		}
		return pointers, err
	}

	return ss.do(recordInfo, f)
}

// do executes the statement and fill the struct or slice given through the
// recordDescription.
func (ss *SelectStatement) do(recordInfo *recordDescription, pointersGetter pointersGetter) error {
	if recordInfo.isSlice == false {
		// Only one row is requested
		ss.Limit(1)
		// Some DB require an offset if a limit is specified (MS SQL Server)
		if ss.offset == nil {
			ss.Offset(0)
		}
		// Some DB require an order by if offset and limit are used
		// (still MS SQL Server)
		if len(ss.orderBy) == 0 {
			keysColumns := recordInfo.structMapping.GetKeyColumnsNames()
			for _, keyColumn := range keysColumns {
				ss.OrderBy(keyColumn)
			}
		}
	}

	sqlQuery, args, err := ss.ToSQL()
	if err != nil {
		return err
	}

	rowsCount, err := ss.db.doSelectOrWithReturning(sqlQuery, args, recordInfo, pointersGetter)
	if err != nil {
		return err
	}

	// When a single instance is requested but not found, sql.ErrNoRows is
	// returned like QueryRow in database/sql package.
	if recordInfo.isSlice == false && rowsCount == 0 {
		err = sql.ErrNoRows
	}

	return err
}

// Scanx runs the request and scans results to dest params
func (ss *SelectStatement) Scanx(dest ...interface{}) error {
	stmt, args, err := ss.ToSQL()
	if err != nil {
		return err
	}
	stmt = ss.db.replacePlaceholders(stmt)

	startTime := time.Now()
	queryable, err := ss.db.getQueryable(stmt)
	if err != nil {
		ss.db.logExecutionErr(err, stmt, args)
		return err
	}
	err = queryable.QueryRow(args...).Scan(dest...)
	consumedTime := timeElapsedSince(startTime)
	ss.db.addConsumedTime(consumedTime)
	ss.db.logExecution(consumedTime, stmt, args)
	if err != nil {
		ss.db.logExecutionErr(err, stmt, args)
		return err
	}

	return nil
}

// Count runs the request with COUNT(*) (remove others columns)
// and returns the count.
func (ss *SelectStatement) Count() (int64, error) {
	ss.columns = ss.columns[:0]
	ss.Columns("COUNT(*)")

	var count int64
	err := ss.Scanx(&count)
	return count, err
}

// DoWithIterator executes the select query and returns an Iterator allowing
// the caller to fetch rows one at a time.
// Warning : it does not use an existing transation to avoid some pitfalls with
// drivers, nor the prepared statement.
func (ss *SelectStatement) DoWithIterator() (Iterator, error) {
	sqlQuery, args, err := ss.ToSQL()
	if err != nil {
		return nil, err
	}

	return ss.db.doWithIterator(sqlQuery, args)
}
