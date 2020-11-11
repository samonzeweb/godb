package ql

import (
	"strings"

	"github.com/samonzeweb/godb/dberror"

	ql "modernc.org/ql"
)

// QL is struct for adapter to be used for modernc.org/ql
type QL struct{}

// Adapter is QL adapter
var Adapter = QL{}

// DriverName returs drivername
func (QL) DriverName() string {
	return "ql"
}

// Quote quotes identifier
func (QL) Quote(identifier string) string {
	return "\"" + identifier + "\""
}

// ParseError parses underlying error and returns consistent error for common ones
func (QL) ParseError(err error) error {
	if err == nil {
		return nil
	}

	if ql.IsDuplicateUniqueIndexError(err) {
		return dberror.UniqueConstraint{Message: err.Error(), Field: strings.Split(strings.Split(err.Error(), "duplicate value(s): ")[1], ".")[1], Err: err}
	}
	// TODO: No constraint checks yet. Add when constraint checks is added to "modernc.org/ql"
	return err
}
