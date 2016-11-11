// The package adapters contains database specific code, mainly in
// sub-packages
package adapters

type DriverName interface {
	DriverName() string
}
