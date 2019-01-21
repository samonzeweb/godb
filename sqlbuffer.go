package godb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/samonzeweb/godb/adapters"
)

// SQLBuffer is a buffer for creating SQL queries with arguments of
// effectively, using bytes.Buffer.
//
// Create a buffer, add SQL parts with their arguments with the Write method and
// its friends, and get the result with SQL() and Arguments(). Check the presence
// presence of an error with Err().
//
// Unlike godb.Q, SQLBuffer does not check if arguments count and placeholders
// count matches, because strings similar to placeholders could be valid in
// other circumstances.
//
// godb use it internally, but you can use it yourself to create raw queries.
type SQLBuffer struct {
	err       error
	sql       *bytes.Buffer
	arguments []interface{}
}

// NewSQLBuffer create a new SQLBuffer, preallocating sqlLength bytes for the
// SQL parts, and argsLength for the arguments list.
// The lengths could be zero, but it's more efficient to give an approximate
// size.
func NewSQLBuffer(sqlLength int, argsLength int) *SQLBuffer {
	return &SQLBuffer{
		sql:       bytes.NewBuffer(make([]byte, 0, sqlLength)),
		arguments: make([]interface{}, 0, argsLength),
	}
}

// SQL returns the string for the SQL part.
// It will create the string from the buffer, avoid calling it multiple times.
func (b *SQLBuffer) SQL() string {
	return b.sql.String()
}

// Arguments returns the arguments given while building the SQL query.
func (b *SQLBuffer) Arguments() []interface{} {
	return b.arguments
}

// SQLLen returns the length of the SQL part (it's bytes count, not characters
// count, beware of unicode !).
// Use it instead of len(myBuffer.SQL()), it's faster and does not allocate
// memory.
func (b *SQLBuffer) SQLLen() int {
	return b.sql.Len()
}

// Err returns the error that may have occurred during the build.
func (b *SQLBuffer) Err() error {
	return b.err
}

// Write adds a sql string and its arguments into the buffer.
func (b *SQLBuffer) Write(sql string, args ...interface{}) *SQLBuffer {
	if b.err != nil {
		return b
	}

	b.sql.WriteString(sql)
	b.arguments = append(b.arguments, args...)
	return b
}

// WriteIfNotEmpty writes the given string only if the sql buffer isn't empty.
func (b *SQLBuffer) WriteIfNotEmpty(sql string, args ...interface{}) *SQLBuffer {
	if b.err != nil {
		return b
	}

	if b.sql.Len() > 0 {
		b.Write(sql, args...)
	}
	return b
}

// WriteBytes add the givent bytes to the internal SQL buffer, and append
// givens arguments to the existing ones.
// It's useful when you have build something with a bytes.Buffer.
func (b *SQLBuffer) WriteBytes(sql []byte, args ...interface{}) *SQLBuffer {
	if b.err != nil {
		return b
	}

	b.sql.Write(sql)
	b.arguments = append(b.arguments, args...)
	return b
}

// WriteStrings writes strings separated by the given separator.
func (b *SQLBuffer) WriteStrings(separator string, sqlParts ...string) *SQLBuffer {
	if b.err != nil {
		return b
	}

	first := true
	for _, sql := range sqlParts {
		if !first {
			b.sql.WriteString(separator)
		} else {
			first = false
		}
		b.sql.WriteString(sql)
	}
	return b
}

// Append add to the buffer the SQL string and arguments from other buffer.
// It does not add separator like space between the sql parts, if needed do it
// yourself.
func (b *SQLBuffer) Append(other *SQLBuffer) *SQLBuffer {
	if b.err != nil {
		return b
	}

	if other.err != nil {
		b.err = other.err
		return b
	}

	b.sql.Write(other.sql.Bytes())
	b.arguments = append(b.arguments, other.arguments...)
	return b
}

// WriteCondition writes single conditional expressions.
func (b *SQLBuffer) WriteCondition(condition *Condition) *SQLBuffer {
	if b.err != nil {
		return b
	}

	if condition.Err() != nil {
		b.err = condition.Err()
		return b
	}

	b.Write(condition.sql, condition.args...)
	return b
}

// ----------------------------------------------------------------------------

// sqlBuffer is an temporary type to build a SQL query. It differs from
// SQLBuffer as it's only used internally by godb (private), and in some
// cases it relies on the adapter.
type sqlBuffer struct {
	adapter adapters.Adapter
	*SQLBuffer
}

// newsqlBuffer creates a new buffer to build SQL query with corresponding arguments.
func newSQLBuffer(adapter adapters.Adapter, sqlLength int, argsLength int) *sqlBuffer {
	return &sqlBuffer{
		adapter:   adapter,
		SQLBuffer: NewSQLBuffer(sqlLength, argsLength),
	}
}

// writeStringsWithSpaces writes strings separated by spaces, and with a
// leading space.
func (b *sqlBuffer) writeStringsWithSpaces(customs []string) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if len(customs) > 0 {
		b.WriteIfNotEmpty(" ")
	}
	b.WriteStrings(" ", customs...)
	return b
}

// writeColumns writes a list of columns into the buffer.
func (b *sqlBuffer) writeColumns(columns []string) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if len(columns) == 0 {
		b.err = fmt.Errorf("Missing columns in statement")
		return b
	}

	b.writeNameList(columns)
	return b
}

// writeFrom writes FROM clause into the buffer.
func (b *sqlBuffer) writeFrom(fromTables ...string) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if len(fromTables) == 0 {
		b.err = fmt.Errorf("No from clause in statement")
		return b
	}

	b.Write(" FROM ")
	b.writeNameList(fromTables)
	return b
}

// writeJoins writes JOIN clause into the buffer.
func (b *sqlBuffer) writeJoins(joins []*joinPart) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	for _, join := range joins {
		b.WriteIfNotEmpty(" ").
			Write(join.joinType).
			Write(" ").
			Write(join.tableName)
		if join.as != "" {
			b.Write(" AS ").
				Write(join.as)
		}
		if join.on != nil {
			b.Write(" ON ").
				WriteCondition(join.on)
		}
	}

	return b
}

// writeWhere writes WHERE clause into the buffer.
func (b *sqlBuffer) writeWhere(conditions []*Condition) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if len(conditions) != 0 {
		b.Write(" WHERE ")
		b.writeConditions(conditions)
	}

	return b
}

// writeGroupByAndHaving writes ORDER BY and HAVING clauses into the buffer.
func (b *sqlBuffer) writeGroupByAndHaving(columns []string, conditions []*Condition) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if len(columns) != 0 {
		b.WriteIfNotEmpty(" ").
			Write("GROUP BY ")
		b.writeNameList(columns)
	}

	if len(conditions) != 0 {
		if len(columns) == 0 {
			b.err = fmt.Errorf("Having clause without Group By")
			return b
		}
		b.Write(" HAVING ")
		b.writeConditions(conditions)
	}

	return b
}

// writeOrderBy writes ORDER BY clause into the buffer.
func (b *sqlBuffer) writeOrderBy(columns []string) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if len(columns) != 0 {
		b.Write(" ORDER BY ")
		b.writeNameList(columns)
	}

	return b
}

// writeOffset writes OFFSET clause into the buffer.
func (b *sqlBuffer) writeOffset(offset *int) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if offset != nil {
		offsetBuilder, ok := b.adapter.(adapters.OffsetBuilder)
		if ok {
			sqlPart := offsetBuilder.BuildOffset(*offset)
			b.Write(" ").
				Write(sqlPart.Sql, sqlPart.Arguments...)
		} else {
			b.Write(" OFFSET ").
				Write(Placeholder, *offset)
		}
	}

	return b
}

// writeLimit writes LIMIT clauses into the buffer.
func (b *sqlBuffer) writeLimit(limit *int) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if limit != nil {
		limitBuilder, ok := b.adapter.(adapters.LimitBuilder)
		if ok {
			sqlPart := limitBuilder.BuildLimit(*limit)
			b.Write(" ").
				Write(sqlPart.Sql, sqlPart.Arguments...)
		} else {
			b.Write(" LIMIT ").
				Write(Placeholder, *limit)
		}
	}

	return b
}

// writeInto writes INTO clause into the buffer.
func (b *sqlBuffer) writeInto(intoTable string) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if intoTable == "" {
		b.err = fmt.Errorf("No INTO clause in INSERT statement")
		return b
	}

	b.Write("INTO ").
		Write(intoTable)
	return b
}

// writeReturning writes RETURNING clause into the buffer if the given position
// is the one used by the adapter.
//
// If the columns list is empty, it always returns without error.
// If the columns list isn't empty, the adapter have to implements the
// ReturningBuilder interface.
func (b *sqlBuffer) writeReturningForPosition(columns []string, position adapters.ReturningPosition) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if len(columns) == 0 {
		return b
	}

	returningBuilder, ok := b.adapter.(adapters.ReturningBuilder)
	if !ok {
		b.err = fmt.Errorf("The adapter does not manage RETUNING-like clause")
		return b
	}

	if returningBuilder.GetReturningPosition() == position {
		b.Write(" ").
			Write(returningBuilder.ReturningBuild(columns)).
			Write(" ")
	}

	return b
}

// writeInsertValues writes all the values to insert to the database into
// the buffer.
func (b *sqlBuffer) writeInsertValues(args [][]interface{}, columnsCount int) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if len(args) == 0 {
		b.err = fmt.Errorf("Missing values in INSERT statement")
		return b
	}

	// build (?, ?, ?, ?)
	valuesPart := buildGroupOfPlaceholders(columnsCount).Bytes()
	// insert group of placeholders for each group of values
	groupCount := len(args)
	for i, currentGroup := range args {
		if len(currentGroup) != columnsCount {
			b.err = fmt.Errorf("Values count does not match the columns count")
			return b
		}
		b.WriteBytes(valuesPart, currentGroup...)
		if i != (groupCount - 1) {
			b.Write(", ")
		}
	}

	return b
}

// writeSets writes the SET clause of an UPDATE statement.
func (b *sqlBuffer) writeSets(sets []*setPart) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	if len(sets) == 0 {
		b.err = fmt.Errorf("Missing SET clause in UPDATE statement")
		return b
	}

	b.Write(" SET ")

	for i, set := range sets {
		if i != 0 {
			b.Write(", ")
		}
		b.Write(set.column)
		if set.value != nil {
			// column and value are given (not raw sql)
			b.Write("=").
				Write(Placeholder, set.value)
		}
	}

	return b
}

// writeNameList writes a list of names, expressions, ... separated by commas.
func (b *sqlBuffer) writeNameList(nameList []string) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	firstName := true
	for _, name := range nameList {
		name = strings.TrimSpace(name)
		if len(name) == 0 {
			b.err = fmt.Errorf("Empty name")
			return b
		}
		if firstName == false {
			b.Write(", ")
		} else {
			firstName = false
		}
		b.Write(name)
	}

	return b
}

// writeConditions writes conditional expressions for WHERE or HAVING clauses
// separated by AND conjunction.
func (b *sqlBuffer) writeConditions(conditions []*Condition) *sqlBuffer {
	if b.Err() != nil {
		return b
	}

	b.WriteCondition(And(conditions...))
	return b
}

// buildGroupOfPlaceholders builds a group of placeholders for values
// like : (?, ?, ?, ?)
func buildGroupOfPlaceholders(placeholderCount int) *bytes.Buffer {
	placeholderGroup := bytes.NewBuffer(make([]byte, 0, placeholderCount*3))
	placeholderGroup.WriteString("(")
	placdeholderCommaSpace := Placeholder + ", "
	for i := 1; i <= placeholderCount; i++ {
		if i != placeholderCount {
			placeholderGroup.WriteString(placdeholderCommaSpace)
		} else {
			placeholderGroup.WriteString(Placeholder)
		}
	}
	placeholderGroup.WriteString(")")

	return placeholderGroup
}
