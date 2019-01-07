package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Map makes easy to handle JSON encoded map stored at database's text fields(like VARCHAR,CHAR,TEXT) and blob fields
type Map map[string]interface{}

// Value returns value.
// If value is invalid value, it returns an error.
func (ss Map) Value() (driver.Value, error) {
	if ss == nil {
		return nil, nil
	}
	return json.Marshal(ss)
}

// Scan stores the value as Map. Value can be string, []byte or or nil.
func (ss *Map) Scan(src interface{}) error {
	var source []byte
	switch t := src.(type) {
	case string:
		if len(t) == 0 {
			source = []byte("{}")
		} else {
			source = []byte(t)
		}
	case []byte:
		if len(t) == 0 {
			source = []byte("{}")
		} else {
			source = t
		}
	case nil:
		*ss = nil
		return nil
	default:
		return fmt.Errorf("Incompatible type for Map")
	}
	err := json.Unmarshal(source, ss)
	return err
}

// String does pretty printing
func (ss Map) String() string {
	out, _ := json.Marshal(ss)
	return string(out)
}
