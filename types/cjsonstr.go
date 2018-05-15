package types

import (
	"encoding/json"
	"fmt"
	"database/sql/driver"
	"bytes"
)
// CompactJSONStr is `json.RawMessage` (`[]byte`) type, adding `Value()` and `Scan()` methods for db access
// and `MarshalJSON`, `UnmarshalJSON` for json serializing.
// Also `Unmarshal` method can be used to unmarshals json to an interface type.
// This type make use of DB binary or string fields like a JSON container.
// DB field type can be text types like CHAR, STRING, CHARACTER VARYING, TEXT or binary types like JSONB, BYTEA, BLOB...
//
// JSON data is hold as "compacted" mode in database.
//
// For performance binary db field use is advised.
type CompactJSONStr json.RawMessage

// MarshalJSON returns JSON encoding
func (js CompactJSONStr) MarshalJSON() ([]byte, error) {
	if len(js) == 0 {
		return CompactJSONStr("{}"), nil
	}
	return js, nil
}

// UnmarshalJSON sets copy of data to instance
func (js *CompactJSONStr) UnmarshalJSON(data []byte) error {
	if js == nil {
		return fmt.Errorf("CompactJSONStr: UnmarshalJSON on nil pointer")
	}
	*js = append((*js)[0:0], data...)
	return nil
}

// Value returns value. Validates JSON.
// If value is invalid json, it returns an error.
func (c CompactJSONStr) Value() (driver.Value, error) {
	var m json.RawMessage
	var err = c.Unmarshal(&m)
	if err != nil {
		return []byte{}, err
	}
	var buf bytes.Buffer
	err = json.Compact(&buf, []byte(c))
	return buf.Bytes(), err
}

// Scan stores the value as CompactJSONStr. Value can be string, []byte or or nil.
func (js *CompactJSONStr) Scan(src interface{}) error {
	var source []byte
	switch t := src.(type) {
	case string:
		source = []byte(t)
	case []byte:
		if len(t) == 0 {
			source = CompactJSONStr("{}")
		} else {
			source = t
		}
	case nil:
		*js = CompactJSONStr("{}")
	default:
		return fmt.Errorf("Incompatible type for CompactJSONStr")
	}
	*js = CompactJSONStr(append((*js)[0:0], source...))
	return nil
}

// Unmarshal unmarshal's using json.Unmarshal.
func (js *CompactJSONStr) Unmarshal(v interface{}) error {
	if len(*js) == 0 {
		*js = CompactJSONStr("{}")
	}
	return json.Unmarshal([]byte(*js), v)
}

// String does pretty printing
func (js CompactJSONStr) String() string {
	return string(js)
}

// NullCompactJSONStr can be an CompactJSONStr or a null value. Can be used like `JSONStr`
type NullCompactJSONStr struct {
	CompactJSONStr
	Valid bool // Valid is true if CompactJSONStr is not NULL
}

// Scan implements the Scanner interface.
func (n *NullCompactJSONStr) Scan(value interface{}) error {
	if value == nil {
		n.CompactJSONStr, n.Valid = CompactJSONStr("{}"), false
		return nil
	}
	n.Valid = true
	return n.CompactJSONStr.Scan(value)
}

// Value implements the driver Valuer interface.
func (n NullCompactJSONStr) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.CompactJSONStr.Value()
}

// NullCompactJSONStrFrom creates a valid NullCompactJSONStr
func NullCompactJSONStrFrom(dst []byte) NullCompactJSONStr {
	return NullCompactJSONStr{CompactJSONStr: dst, Valid: true}
}
