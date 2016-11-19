// The package adapters contains database specific code, mainly in
// sub-packages
package adapters

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
