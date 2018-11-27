package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Int64Slice makes easy to handle JSON encoded string lists from/to db stored either in TEXT or BLOB.
//
// Int64Slice is `[]String` type, adding `Value()` and `Scan()` methods for db access.
//

// Int64Slice makes easy to handle JSON data at database's text fields(like VARCHAR,CHAR,TEXT) and blob fields
type Int64Slice []int64

// Value returns value.
// If value is invalid value, it returns an error.
func (ss Int64Slice) Value() (driver.Value, error) {
	if ss == nil {
		return nil, nil
	}
	return json.Marshal(ss)
}

// Scan stores the value as Int64Slice. Value can be string, []byte or or nil.
func (ss *Int64Slice) Scan(src interface{}) error {
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
		return fmt.Errorf("Incompatible type for Int64Slice")
	}
	err := json.Unmarshal(source, ss)
	return err
}
