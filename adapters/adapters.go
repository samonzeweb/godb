// The package adapters contains database specific code, mainly in
// sub-packages
package adapters

// Adapter interface is the minimal implementation for an adapter.
type Adapter interface {
	DriverNamer
	Quoter
}

// DriverNamer is an interface that wraps the mandatory DriverName method.
// Its implemntation is required for all adapters.
//
// DriverName must returns the driver name to be used with sql.Open()
type DriverNamer interface {
	DriverName() string
}

// Quoter is an interface that wraps the mandatory Quote method.
//
// Quote must returns an SQL identifier (table name, column name) quoted,
// ie : "foo" for SQLite or Postgresql, `foo` for MySQL, [foo] for SQLServer.
type Quoter interface {
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

// InsertSuffixer is an interface that wraps the optionnal InsertSuffix method.
//
// InsertSuffix get a list of columns and return a suffix to be added to the
// sql statement by the caller, allowing it to retrieve the values of those
// columns. It is intended to replace the use of LastInsertId() when the
// driver does not support it.
type InsertSuffixer interface {
	InsertSuffix([]string) string
}
