package mssql

import "database/sql/driver"
import "fmt"

// Rowversion represents a rowversion (or timestamp) column type
// See https://docs.microsoft.com/en-us/sql/t-sql/data-types/rowversion-transact-sql
type Rowversion struct {
	Version []uint8
}

// Scan implements sql.Scanner
func (r *Rowversion) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("rowversion does not accept NULL values")
	}

	source := value.([]byte)
	sourceLen := len(source)
	if len(r.Version) != sourceLen {
		r.Version = make([]uint8, sourceLen)
	}
	copy(r.Version, source)
	return nil
}

// Value implements driver.Valuer
func (r Rowversion) Value() (driver.Value, error) {
	return r.Version, nil
}

// Copy returns a copy of the given Rowversion
func (r Rowversion) Copy() Rowversion {
	newRowversion := Rowversion{}
	newRowversion.Version = append(newRowversion.Version, r.Version...)
	return newRowversion
}
