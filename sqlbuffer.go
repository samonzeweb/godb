package godb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/samonzeweb/godb/adapters"
)

// sqlBuffer is an temporary type to build a SQL query with its arguments.
// After the building operation use String() and Arguments() to get data to
// use with database/sql.
type sqlBuffer struct {
	adapter   adapters.Adapter
	sql       *bytes.Buffer
	arguments []interface{}
}

// newsqlBuffer creates a new buffer to build SQL query with corresponding arguments.
func newSQLBuffer(adapter adapters.Adapter, sqlLength int, argsLength int) *sqlBuffer {
	return &sqlBuffer{
		adapter:   adapter,
		sql:       bytes.NewBuffer(make([]byte, 0, sqlLength)),
		arguments: make([]interface{}, 0, argsLength),
	}
}

// String returns the SQL string.
func (b *sqlBuffer) sqlString() string {
	return b.sql.String()
}

// sqlArguments returns the arguments slice.
func (b *sqlBuffer) sqlArguments() []interface{} {
	return b.arguments
}

// write adds a custom string and arguments into the buffer.
func (b *sqlBuffer) write(sql string, args ...interface{}) error {
	b.sql.WriteString(sql)
	b.arguments = append(b.arguments, args...)

	return nil
}

// writeIfNotEmpty writes the given string if the sql buffer isn't empty.
func (b *sqlBuffer) writeIfNotEmpty(custom string) error {
	if b.sql.Len() > 0 {
		b.sql.WriteString(custom)
	}

	return nil
}

// writeStrings writes strings separated by spaces.
func (b *sqlBuffer) writeStrings(customs []string) error {
	for _, custom := range customs {
		b.writeIfNotEmpty(" ")
		b.sql.WriteString(custom)
	}
	return nil
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

	b.sql.WriteString(" FROM ")
	err := b.writeNameList(fromTables)
	return err
}

// writeJoins writes JOIN clause into the buffer.
func (b *sqlBuffer) writeJoins(joins []*joinPart) error {
	for _, join := range joins {
		b.writeIfNotEmpty(" ")
		b.sql.WriteString(join.joinType)
		b.sql.WriteString(" ")
		b.sql.WriteString(join.tableName)
		if join.as != "" {
			b.sql.WriteString(" AS ")
			b.sql.WriteString(join.as)
		}
		if join.on != nil {
			b.sql.WriteString(" ON ")
			b.writeCondition(join.on)
		}
	}

	return nil
}

// writeWhere writes WHERE clause into the buffer.
func (b *sqlBuffer) writeWhere(conditions []*Condition) error {
	if len(conditions) != 0 {
		b.sql.WriteString(" WHERE ")
		b.writeConditions(conditions)
	}

	return nil
}

// writeGroupByAndHaving writes ORDER BY and HAVING clauses into the buffer.
func (b *sqlBuffer) writeGroupByAndHaving(columns []string, conditions []*Condition) error {
	if len(columns) != 0 {
		b.writeIfNotEmpty(" ")
		b.sql.WriteString("GROUP BY ")
		b.writeNameList(columns)
	}

	if len(conditions) != 0 {
		if len(columns) == 0 {
			return fmt.Errorf("Having clause without Group By")
		}
		b.sql.WriteString(" HAVING ")
		b.writeConditions(conditions)
	}

	return nil
}

// writeOrderBy writes ORDER BY clause into the buffer.
func (b *sqlBuffer) writeOrderBy(columns []string) error {
	if len(columns) != 0 {
		b.sql.WriteString(" ORDER BY ")
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
			b.sql.WriteString(" ")
			b.sql.WriteString(sqlPart.Sql)
			b.arguments = append(b.arguments, sqlPart.Arguments...)
		} else {
			b.sql.WriteString(" OFFSET ")
			b.sql.WriteString(Placeholder)
			b.arguments = append(b.arguments, *offset)
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
			b.sql.WriteString(" ")
			b.sql.WriteString(sqlPart.Sql)
			b.arguments = append(b.arguments, sqlPart.Arguments...)
		} else {
			b.sql.WriteString(" LIMIT ")
			b.sql.WriteString(Placeholder)
			b.arguments = append(b.arguments, *limit)
		}
	}

	return nil
}

// writeInto writes INTO clause into the buffer.
func (b *sqlBuffer) writeInto(intoTable string) error {
	if intoTable == "" {
		return fmt.Errorf("No INTO clause in INSERT statement")
	}

	b.sql.WriteString("INTO ")
	b.sql.WriteString(intoTable)
	return nil
}

// writeInsertValues writes all the values to insert to the database into
// the buffer.
func (b *sqlBuffer) writeInsertValues(args [][]interface{}, columnsCount int) error {
	if len(args) == 0 {
		return fmt.Errorf("Missing values in INSERT statement")
	}

	// build (?, ?, ?, ?)
	valuesPart := buildGroupOfPlaceholders(columnsCount)

	// insert group of placeholders for each group of values
	groupCount := len(args)
	for i, currentGroup := range args {
		if len(currentGroup) != columnsCount {
			return fmt.Errorf("Values count does not match the columns count")
		}
		b.sql.Write(valuesPart.Bytes())
		if i != (groupCount - 1) {
			b.sql.WriteString(", ")
		}
		b.arguments = append(b.arguments, currentGroup...)
	}

	return nil
}

// writeSets writes the SET clause of an UPDATE statement.
func (b *sqlBuffer) writeSets(sets []*setPart) error {
	if len(sets) == 0 {
		return fmt.Errorf("Missing SET clause in UPDATE statement")
	}

	b.sql.WriteString(" SET ")

	for i, set := range sets {
		if i != 0 {
			b.sql.WriteString(", ")
		}
		b.sql.WriteString(set.column)
		if set.value != nil {
			// column and value are given (not raw sql)
			b.sql.WriteString("=")
			b.sql.WriteString(Placeholder)
			b.arguments = append(b.arguments, set.value)
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
			b.sql.WriteString(", ")
		} else {
			firstName = false
		}
		b.sql.WriteString(name)
	}

	return nil
}

// writeConditions writes single conditionnal expressions
func (b *sqlBuffer) writeCondition(condition *Condition) error {
	b.sql.WriteString(condition.sql)
	b.arguments = append(b.arguments, condition.args...)

	return nil
}

// writeConditions writes conditionnal expressions for WHERE or HAVING clauses
// separated by AND conjunction.
func (b *sqlBuffer) writeConditions(conditions []*Condition) error {
	c := And(conditions...)
	return b.writeCondition(c)
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
