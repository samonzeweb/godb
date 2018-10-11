// Package adapters contains database specific code, mainly in
// sub-packages.
package adapters

// Adapter interface is the minimal implementation for an adapter.
type Adapter interface {
	// DriverName must return the driver name to be used with sql.Open()
	DriverName() string
	// Quote must return an SQL identifier (table name, column name) quoted,
	// ie : "foo" for SQLite or Postgresql, `foo` for MySQL, [foo] for SQLServer.
	Quote(string) string
	// ParseError parses adapter errors and returns a understandable DBError
	ParseError(error) error
}

// PlaceholdersReplacer is an interface wrapping the optional
// ReplacePlaceholders method.
//
// ReplacePlaceholders changes all given placeholders in given sql query with
// the placeholder used by the database targeted by the adapter.
type PlaceholdersReplacer interface {
	ReplacePlaceholders(string, string) string
}

// ReturningBuilder is an interface wrapping the optional ReturningBuild
// and ReturningNewValues method.
//
// ReturningBuild gets a list of expressions and returns a clause to be added
// to the sql statement by the caller, allowing it to retrieve the values
// of those expressions.
//
// It is intended to either replace the use of LastInsertId() when the driver
// does not support it, or to fetch all fields initialized or updated by the
// database itself. PostgreSQL and SQL Server are concerned with the RETURNING
// and OUTPUT clauses.
//
// FormatForNewValues get a list of columns and format all of them to
// have expressions returning new values. The purpose is to always get new
// values when the database could either return the old or new values
// (before/after execution of the sql statement).
type ReturningBuilder interface {
	ReturningBuild([]string) string
	FormatForNewValues([]string) []string
	GetReturningPosition() ReturningPosition
}

// ReturningPosition specify the position of the returning clause if the sql
// statement.
//
// It's not abstract enough, some things are too coupled. Changing that will
// need a huge rewrite of godb, then now it does the trick.
type ReturningPosition int

const (
	// ReturningPostgreSQL for PostgreSQL
	ReturningPostgreSQL ReturningPosition = 1
	// ReturningSQLServer for SQL Server
	ReturningSQLServer = 2
)

// SQLPart is a struct containing a custom part of SQL query builded by an
// adapter.
type SQLPart struct {
	Sql       string
	Arguments []interface{}
}

// LimitBuilder is an interface wrapping the optional BuildLimit method.
//
// BuildLimit get an integer and returns a string containing a LIMIT sql clause
// or its equivalent for the adapter, and an array of sql arguments.
type LimitBuilder interface {
	BuildLimit(int) *SQLPart
}

// OffsetBuilder is an interface wrapping the optional BuildOffset method.
//
// BuildOffset get an integer and returns a string containing an OFFSET sql
// clause or its equivalent for the adapter, and an array of sql arguments.
type OffsetBuilder interface {
	BuildOffset(int) *SQLPart
}

// LimitOffsetOrderer is an interface wrapping the optional IsOffsetFirst
// method.
//
// The IsOffsetFirst returns true is the OFFSET clause has to precede the
// LIMIT clause. By default the LIMIT is before the OFFSET.
type LimitOffsetOrderer interface {
	IsOffsetFirst() bool
}
