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
// Create a buffer, add SQL parts with their arguments with Write method and
// its friends, and get the result with SQL() and Arguments().
//
// godb use it internally, but you can use it yourself to create raw queries.
type SQLBuffer struct {
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

// Write adds a sql string and its arguments into the buffer.
func (b *SQLBuffer) Write(sql string, args ...interface{}) *SQLBuffer {
	b.sql.WriteString(sql)
	b.arguments = append(b.arguments, args...)
	return b
}

// WriteIfNotEmpty writes the given string only if the sql buffer isn't empty.
func (b *SQLBuffer) WriteIfNotEmpty(sql string, args ...interface{}) *SQLBuffer {
	if b.sql.Len() > 0 {
		b.Write(sql, args...)
	}
	return b
}

// WriteBytes add the givent bytes to the internal SQL buffer, and append
// gibens arguments to the existing ones.
// It's useful when you have build something with a bytes.Buffer.
func (b *SQLBuffer) WriteBytes(sql []byte, args ...interface{}) *SQLBuffer {
	b.sql.Write(sql)
	b.arguments = append(b.arguments, args...)
	return b
}

// WriteStrings writes strings separated by the given separator.
func (b *SQLBuffer) WriteStrings(separator string, sqlParts ...string) *SQLBuffer {
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
	b.sql.Write(other.sql.Bytes())
	b.arguments = append(b.arguments, other.arguments...)

	return b
}

// WriteCondition writes single conditionnal expressions.
func (b *SQLBuffer) WriteCondition(condition *Condition) *SQLBuffer {
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
func (b *sqlBuffer) writeStringsWithSpaces(customs []string) {
	if len(customs) > 0 {
		b.WriteIfNotEmpty(" ")
	}
	b.WriteStrings(" ", customs...)
}

// writeColumns writes a list of columns into the buffer.
func (b *sqlBuffer) writeColumns(columns []string) error {
	if len(columns) == 0 {
		return fmt.Errorf("Missing columns in statement")
	}

	b.writeNameList(columns)
	return nil
}

// writeFrom writes FROM clause into the buffer.
func (b *sqlBuffer) writeFrom(fromTables ...string) error {
	if len(fromTables) == 0 {
		return fmt.Errorf("No from clause in statement")
	}

	b.Write(" FROM ")
	err := b.writeNameList(fromTables)
	return err
}

// writeJoins writes JOIN clause into the buffer.
func (b *sqlBuffer) writeJoins(joins []*joinPart) error {
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

	return nil
}

// writeWhere writes WHERE clause into the buffer.
func (b *sqlBuffer) writeWhere(conditions []*Condition) error {
	if len(conditions) != 0 {
		b.Write(" WHERE ")
		b.writeConditions(conditions)
	}

	return nil
}

// writeGroupByAndHaving writes ORDER BY and HAVING clauses into the buffer.
func (b *sqlBuffer) writeGroupByAndHaving(columns []string, conditions []*Condition) error {
	if len(columns) != 0 {
		b.WriteIfNotEmpty(" ").
			Write("GROUP BY ")
		b.writeNameList(columns)
	}

	if len(conditions) != 0 {
		if len(columns) == 0 {
			return fmt.Errorf("Having clause without Group By")
		}
		b.Write(" HAVING ")
		b.writeConditions(conditions)
	}

	return nil
}

// writeOrderBy writes ORDER BY clause into the buffer.
func (b *sqlBuffer) writeOrderBy(columns []string) error {
	if len(columns) != 0 {
		b.Write(" ORDER BY ")
		b.writeNameList(columns)
	}

	return nil
}

// writeOffset writes OFFSET clause into the buffer.
func (b *sqlBuffer) writeOffset(offset *int) error {
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

	return nil
}

// writeLimit writes LIMIT clauses into the buffer.
func (b *sqlBuffer) writeLimit(limit *int) error {
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

	return nil
}

// writeInto writes INTO clause into the buffer.
func (b *sqlBuffer) writeInto(intoTable string) error {
	if intoTable == "" {
		return fmt.Errorf("No INTO clause in INSERT statement")
	}

	b.Write("INTO ").
		Write(intoTable)
	return nil
}

// writeReturning writes RETURNING clause into the buffer if the given position
// is the one used by the adapter.
//
// If the columns list is empty, it always returns without error.
// If the columns list isn't empty, the adapter have to implements the
// ReturningBuilder interface.
func (b *sqlBuffer) writeReturningForPosition(columns []string, position adapters.ReturningPosition) error {

	if len(columns) == 0 {
		return nil
	}

	returningBuilder, ok := b.adapter.(adapters.ReturningBuilder)
	if !ok {
		return fmt.Errorf("The adapter does not manage RETUNING-like clause")
	}

	if returningBuilder.GetReturningPosition() == position {
		b.Write(" ").
			Write(returningBuilder.ReturningBuild(columns)).
			Write(" ")
	}

	return nil
}

// writeInsertValues writes all the values to insert to the database into
// the buffer.
func (b *sqlBuffer) writeInsertValues(args [][]interface{}, columnsCount int) error {
	if len(args) == 0 {
		return fmt.Errorf("Missing values in INSERT statement")
	}

	// build (?, ?, ?, ?)
	valuesPart := buildGroupOfPlaceholders(columnsCount).Bytes()

	// insert group of placeholders for each group of values
	groupCount := len(args)
	for i, currentGroup := range args {
		if len(currentGroup) != columnsCount {
			return fmt.Errorf("Values count does not match the columns count")
		}
		b.WriteBytes(valuesPart, currentGroup...)
		if i != (groupCount - 1) {
			b.Write(", ")
		}
	}

	return nil
}

// writeSets writes the SET clause of an UPDATE statement.
func (b *sqlBuffer) writeSets(sets []*setPart) error {
	if len(sets) == 0 {
		return fmt.Errorf("Missing SET clause in UPDATE statement")
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

	return nil
}

// writeNameList writes a list of names, expressions, ... separated by commas.
func (b *sqlBuffer) writeNameList(nameList []string) error {
	firstName := true
	for _, name := range nameList {
		name = strings.TrimSpace(name)
		if len(name) == 0 {
			return fmt.Errorf("Empty name")
		}
		if firstName == false {
			b.Write(", ")
		} else {
			firstName = false
		}
		b.Write(name)
	}

	return nil
}

// writeConditions writes conditionnal expressions for WHERE or HAVING clauses
// separated by AND conjunction.
func (b *sqlBuffer) writeConditions(conditions []*Condition) {
	c := And(conditions...)
	b.WriteCondition(c)
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
