package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// BoolSlice makes easy to handle JSON encoded bool lists from/to db stored either in TEXT or BLOB.
//
// BoolSlice is `[]bool` type, adding `Value()` and `Scan()` methods for db access.
//

// BoolSlice makes easy to handle JSON encoded bool lists stored at database's text fields(like VARCHAR,CHAR,TEXT) and blob fields
type BoolSlice []bool

// Value returns value.
// If value is invalid value, it returns an error.
func (ss BoolSlice) Value() (driver.Value, error) {
	if ss == nil {
		return nil, nil
	}
	return json.Marshal(ss)
}

// Scan stores the value as BoolSlice. Value can be string, []byte or or nil.
func (ss *BoolSlice) Scan(src interface{}) error {
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
		return fmt.Errorf("Incompatible type for BoolSlice")
	}
	err := json.Unmarshal(source, ss)
	return err
}
