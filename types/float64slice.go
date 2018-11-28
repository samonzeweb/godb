package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Float64Slice makes easy to handle JSON encoded float64 lists from/to db stored either in TEXT or BLOB.
//
// Float64Slice is `[]float64` type, adding `Value()` and `Scan()` methods for db access.
//

// Float64Slice makes easy to handle JSON encoded float64 lists stored at database's text fields(like VARCHAR,CHAR,TEXT) and blob fields
type Float64Slice []float64

// Value returns value.
// If value is invalid value, it returns an error.
func (ss Float64Slice) Value() (driver.Value, error) {
	if ss == nil {
		return nil, nil
	}
	return json.Marshal(ss)
}

// Scan stores the value as Float64Slice. Value can be string, []byte or or nil.
func (ss *Float64Slice) Scan(src interface{}) error {
	var source []byte
	switch t := src.(type) {
	case string:
		if len(t) == 0 {
			source = []byte("[]")
		} else {
			source = []byte(t)
		}
	case []byte:
		if len(t) == 0 {
			source = []byte("[]")
		} else {
			source = t
		}
	case nil:
		*ss = nil
		return nil
	default:
		return fmt.Errorf("Incompatible type for Float64Slice")
	}
	err := json.Unmarshal(source, ss)
	return err
}
