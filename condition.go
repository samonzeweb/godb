package godb

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

// Condition is a struct allowing complex condition building, composing
// SQL predicates, and managing associated arguments.
type Condition struct {
	error error
	sql   string
	args  []interface{}
}

// Q builds a simple condition, managing slices in a particular way : it
// replace the single placeholder with multiples ones according to the number
// of arguments.
func Q(sql string, args ...interface{}) *Condition {
	c := Condition{}

	if strings.Count(sql, Placeholder) != len(args) {
		c.error = fmt.Errorf("Wrong number of arguments in condition %s", sql)
		return &c
	}

	buffer := bytes.NewBuffer(make([]byte, 0, len(sql)))
	var placeholderPos int
	remainingSQL := sql[:]
	// Search slice args to manage case like "WHERE id IN (?)"
	for _, arg := range args {
		if arg == nil {
			c.error = fmt.Errorf("Using nil as argument in condition %s", sql)
			return &c
		}

		placeholderPos = strings.Index(remainingSQL, Placeholder)
		buffer.WriteString(remainingSQL[:placeholderPos])
		remainingSQL = remainingSQL[placeholderPos+1:]
		t := reflect.TypeOf(arg)
		// t could be nil if arguments are not given (nil) to prepare a sql statement
		if t != nil && t.Kind() == reflect.Slice {
			// Slices. They can't be empty.
			v := reflect.ValueOf(arg)
			length := reflect.ValueOf(arg).Len()
			if length == 0 {
				c.error = fmt.Errorf("Empty slice used as argument in condition %s", sql)
				return &c
			}
			for i := 0; i < length; i++ {
				c.args = append(c.args, v.Index(i).Interface())
			}
			buffer.WriteString(Placeholder + strings.Repeat(","+Placeholder, length-1))
		} else {
			// Not a slice
			buffer.WriteString(Placeholder)
			c.args = append(c.args, arg)
		}
	}
	buffer.WriteString(remainingSQL)

	c.sql = buffer.String()
	return &c
}

// And combines two or more conditions inserting 'AND' between each
// given conditions.
func And(conditions ...*Condition) *Condition {
	if len(conditions) == 1 {
		return conditions[0]
	}

	sqlLength, argsLength, err := sumOfConditionsLengths(conditions)
	if err != nil {
		return &Condition{error: err}
	}

	// because len(" AND ") == 5
	sqlLength += 5 * (len(conditions) - 1)
	buffer := bytes.NewBuffer(make([]byte, 0, sqlLength))
	joinedArgs := make([]interface{}, 0, argsLength)

	return &Condition{
		sql:  joinSQL(buffer, " AND ", conditions).String(),
		args: joinArgs(joinedArgs, conditions),
	}
}

// Or combines two or more conditions inserting 'OR' between each
// given conditions, and surrounding all with parentheses.
func Or(conditions ...*Condition) *Condition {
	if len(conditions) == 1 {
		return conditions[0]
	}

	sqlLength, argsLength, err := sumOfConditionsLengths(conditions)
	if err != nil {
		return &Condition{error: err}
	}

	// len(" OR ") == 4 , plus parentheses
	sqlLength += 2 + 4*(len(conditions)-1)
	buffer := bytes.NewBuffer(make([]byte, 0, sqlLength))
	buffer.WriteString("(")
	joinSQL(buffer, " OR ", conditions)
	buffer.WriteString(")")

	joinedArgs := make([]interface{}, 0, argsLength)

	return &Condition{
		sql:  buffer.String(),
		args: joinArgs(joinedArgs, conditions),
	}
}

// Not negates a given condition surrounding it with 'NOT (' and ')'.
func Not(condition *Condition) *Condition {
	if condition.error != nil {
		return condition
	}

	// len("NOT (") == 5 , and closing parenthesis
	sqlLength := 6 + len(condition.sql)
	buffer := bytes.NewBuffer(make([]byte, 0, sqlLength))
	buffer.WriteString("NOT (")
	buffer.WriteString(condition.sql)
	buffer.WriteString(")")

	return &Condition{
		sql:  buffer.String(),
		args: condition.args,
	}
}

// sumOfConditionsLengths returns the sum of all sql length, the sum of all
// arguments count, or the first error found.
func sumOfConditionsLengths(conditions []*Condition) (int, int, error) {
	sqlLength := 0
	argsLength := 0
	for _, c := range conditions {
		if c.error != nil {
			return 0, 0, c.error
		}
		sqlLength += len(c.sql)
		argsLength += len(c.args)
	}

	return sqlLength, argsLength, nil
}

// joinSQL all sql in given conditions, inserting a conjunction between each
// of them. Like strings.Join, but with a resulting buffer.
func joinSQL(buffer *bytes.Buffer, conjunction string, conditions []*Condition) *bytes.Buffer {
	firstCondition := true
	for _, c := range conditions {
		if firstCondition == false {
			buffer.WriteString(conjunction)
		} else {
			firstCondition = false
		}
		buffer.WriteString(c.sql)
	}

	return buffer
}

// joinArgs concatenates all arguments of given conditions
func joinArgs(args []interface{}, conditions []*Condition) []interface{} {
	for _, c := range conditions {
		args = append(args, c.args...)
	}

	return args
}
