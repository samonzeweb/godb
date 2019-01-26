package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONStr makes easy to handle JSON data at database's text fields(like VARCHAR,CHAR,TEXT) and blob fields(like BLOB,
// BYTEA, JSONB). NullJSONStr is Nullable version of JSONStr.
//
// JSONStr is `json.RawMessage` (`[]byte`) type, adding `Value()` and `Scan()` methods for db access
// and `MarshalJSON`, `UnmarshalJSON` for json serializing.
//
// This type make use of DB binary or string fields like a JSON container.
// DB field type can be text types like CHAR, STRING, CHARACTER VARYING, TEXT or binary types like JSONB, BYTEA, BLOB...
//
// For performance binary db field use is advised.
//
//
// Example:
//
//  package main
//
//  import (
//  	"fmt"
//  	"github.com/samonzeweb/godb"
//  	"github.com/samonzeweb/godb/adapters/sqlite"
//  	"github.com/samonzeweb/godb/types"
//  )
//
//  type Meta struct {
//  	Code        string `json:"code"`
//  	Population  int64  `json:"population"`
//  	PhonePrefix string `json:"phone_prefix"`
//  }
//
//  type Country struct {
//  	Id    int                 `db:"id,key,auto"`
//  	Meta  types.NullJSONStr   `db:"meta"`
//  	Metab types.NullJSONStr   `db:"metab"`
//  }
//
//  func (b *Country) TableName() string {
//  	return "countries"
//  }
//
//  func main() {
//  	db, err := godb.Open(sqlite.Adapter, "./countries.dat")
//  	if err != nil {
//  		panic(fmt.Sprintf("db connection err: %v", err))
//  	}
//  	if _, err = db.CurrentDB().Exec(`
//  		CREATE TABLE IF NOT EXISTS countries (
//  			id integer not null primary key autoincrement,
//  			meta      TEXT,
//  			metab	  BLOB
//  		);`); err != nil {
//  		panic(err)
//  	}
//
//  	// Insert nullable JSONStr
//  	country := Country{
//  		Meta:      types.ToNullJSONStr([]byte(`{"code": "US", "phone_prefix": "1"}`)),
//  		Metab:     types.ToNullJSONStr([]byte(`{"code": "TR", "phone_prefix": "90"}`)),
//  	}
//  	if err = db.Insert(&country).Do(); err != nil {
//  		panic(err)
//  	}
//
//  	// Select
//  	c := new(Country)
//  	if err = db.Select(c).Where("id = ?", country.Id).Do(); err != nil {
//  		panic(err)
//  	}
//  	fmt.Printf("Meta: %v, Metab: %v\n", c.Meta, c.Metab)
//
//  	// Loading JSON into struct
//  	meta := Meta{}
//  	if err = c.Metab.Unmarshal(&meta); err != nil {
//  		panic(err)
//  	}
//  	fmt.Printf("Meta Struct: %v\n", meta)
//  }
//

// JSONStr makes easy to handle JSON data at database's text fields(like VARCHAR,CHAR,TEXT) and blob fields
type JSONStr json.RawMessage

// MarshalJSON returns JSON encoding
func (js JSONStr) MarshalJSON() ([]byte, error) {
	if len(js) == 0 {
		return JSONStr("{}"), nil
	}
	return js, nil
}

// UnmarshalJSON sets copy of data to instance
func (js *JSONStr) UnmarshalJSON(data []byte) error {
	if js == nil {
		return fmt.Errorf("JSONStr: UnmarshalJSON on nil pointer")
	}
	*js = append((*js)[0:0], data...)
	return nil
}

// Value returns value. Validates JSON.
// If value is invalid json, it returns an error.
func (js JSONStr) Value() (driver.Value, error) {
	var m json.RawMessage
	var err = js.Unmarshal(&m)
	if err != nil {
		return []byte{}, err
	}
	return []byte(js), nil
}

// Scan stores the value as JSONStr. Value can be string, []byte or or nil.
func (js *JSONStr) Scan(src interface{}) error {
	var source []byte
	switch t := src.(type) {
	case string:
		source = []byte(t)
	case []byte:
		if len(t) == 0 {
			source = JSONStr("{}")
		} else {
			source = t
		}
	case nil:
		*js = JSONStr("{}")
	default:
		return fmt.Errorf("incompatible type for JSONStr")
	}
	*js = JSONStr(append((*js)[0:0], source...))
	return nil
}

// Unmarshal unmarshal's using json.Unmarshal.
func (js *JSONStr) Unmarshal(v interface{}) error {
	if len(*js) == 0 {
		*js = JSONStr("{}")
	}
	return json.Unmarshal([]byte(*js), v)
}

// String does pretty printing
func (js JSONStr) String() string {
	return string(js)
}

// NullJSONStr can be an JSONStr or a null value. Can be used like `sql.NullSting`
type NullJSONStr struct {
	JSONStr
	Valid bool // Valid is true if JSONStr is not NULL
}

// Scan implements the Scanner interface.
func (n *NullJSONStr) Scan(value interface{}) error {
	if value == nil {
		n.JSONStr, n.Valid = JSONStr("{}"), false
		return nil
	}
	n.Valid = true
	return n.JSONStr.Scan(value)
}

// Value implements the driver Valuer interface.
func (n NullJSONStr) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.JSONStr.Value()
}

// ToNullJSONStr creates a valid NullJSONStr
func ToNullJSONStr(dst []byte) NullJSONStr {
	return NullJSONStr{JSONStr: dst, Valid: true}
}
