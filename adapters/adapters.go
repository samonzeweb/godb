// The package adapters contains database specific code, mainly in
// sub-packages
package adapters

// Adapter interface is the minimal implementation for an adapter.
type Adapter interface {
	// DriverName must returns the driver name to be used with sql.Open()
	DriverName() string
	// Quote must returns an SQL identifier (table name, column name) quoted,
	// ie : "foo" for SQLite or Postgresql, `foo` for MySQL, [foo] for SQLServer.
	Quote(string) string
}

// PlaceholdersReplacer is an interface that wraps the optionnal
// ReplacePlaceholders method.
//
// PlaceholdersReplacer change all given placeholders in given sql query with
// the placeholder used by the databaser targeted by the adapter.
type PlaceholdersReplacer interface {
	ReplacePlaceholders(string, string) string
}

// InsertReturningSuffixer is an interface that wraps the optionnal
// InsertReturningSuffix method.
//
// InsertReturningSuffixer get a list of columns and return a suffix to be
// added to the sql statement by the caller, allowing it to retrieve the values
// of those columns. It is intended to replace the use of LastInsertId()
// when the driver does not support it, or if there are more 'automatic' fields
// initialized by the database.
type InsertReturningSuffixer interface {
	InsertReturningSuffix([]string) string
}

// SqlPart is a struct containing a custom part of SQL query builded by an
// adapter.
type SQLPart struct {
	Sql       string
	Arguments []interface{}
}

// LimitBuilder is an interface wrapping the optionnal BuildLimit method.
//
// BuildLimit get an integer and returns a string containing a LIMIT sql clause
// or its equivalent for the adapter, and an array of sql arguments.
type LimitBuilder interface {
	BuildLimit(int) *SQLPart
}

// OffsetBuilder is an interface wrapping the optionnal BuildOffset method.
//
// BuildOffset get an integer and returns a string containing an OFFSET sql
// clause or its equivalent for the adapter, and an array of sql arguments.
type OffsetBuilder interface {
	BuildOffset(int) *SQLPart
}

// LimitOffsetOrderer is an interface wrapping the optionnal IsOffsetFirst
// method.
//
// The IsOffsetFirst returns true is the OFFSET clause has to precede the
// LIMIT clause. By default the LIMIT is before the OFFSET.
type LimitOffsetOrderer interface {
	IsOffsetFirst() bool
}
