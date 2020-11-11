package ql2

import (
	"strings"

	"github.com/samonzeweb/godb/dberror"

	ql "modernc.org/ql"
)

// QL2 is struct for adapter to be used for modernc.org/ql
// QL2 uses version2 type of storage for QL
type QL2 struct{}

// Adapter is QL adapter
var Adapter = QL2{}

// DriverName returs drivername
func (QL2) DriverName() string {
	return "ql2"
}

// Quote quotes identifier
func (QL2) Quote(identifier string) string {
	return "\"" + identifier + "\""
}

// ParseError parses underlying error and returns consistent error for common ones
func (QL2) ParseError(err error) error {
	if err == nil {
		return nil
	}

	if ql.IsDuplicateUniqueIndexError(err) {
		return dberror.UniqueConstraint{Message: err.Error(), Field: strings.Split(strings.Split(err.Error(), "duplicate value(s): ")[1], ".")[1], Err: err}
	}
	// TODO: No constraint checks yet. Add when constraint checks is added to "modernc.org/ql"
	return err
}
