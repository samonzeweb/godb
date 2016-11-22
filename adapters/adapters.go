// The package adapters contains database specific code, mainly in
// sub-packages
package adapters

// Adapter interface is the minimal implementation for an adapter.
type Adapter interface {
	DriverNamer
	Quoter
}

// DriverNamer is an interface that wraps the DriverName method.
// Its implemntation is required for all adapters.
//
// DriverName must returns the driver name to be used with sql.Open()
type DriverNamer interface {
	DriverName() string
}

// Quoter is an interface that wraps the Quote method.
//
// Quote must returns an SQL identifier (table name, column name) quoted,
// ie : "foo" for SQLite or Postgresql, `foo` for MySQL, [foo] for SQLServer.
type Quoter interface {
	Quote(string) string
}

// PlaceholdersReplacer is an interface that wraps the ReplacePlaceholders method.
//
// PlaceholdersReplacer change all given placeholders in given sql query with
// the placeholder used by the databaser targeted by the adapter.
type PlaceholdersReplacer interface {
	ReplacePlaceholders(string, string) string
}
