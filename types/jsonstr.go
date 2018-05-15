package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONStr is `json.RawMessage` (`[]byte`) type, adding `Value()` and `Scan()` methods for db access
// and `MarshalJSON`, `UnmarshalJSON` for json serializing.
// Also `Unmarshal` method can be used to unmarshals json to an interface type.
// This type make use of DB binary or string fields like a JSON container.
// DB field type can be text types like CHAR, STRING, CHARACTER VARYING, TEXT or binary types like JSONB, BYTEA, BLOB...
//
// For performance binary db field use is advised.
//
// Example Usage:
//
// For table:
//  	CREATE TABLE books (
//  	meta      TEXT
//  	metab	  BYTEA -- or BLOB or JSONB
//  	);
// Example:
//  	type Book struct {
//  		Id        int                 `db:"id,key,auto"`
//  		Meta      types.NullJSONStr   `db:"meta"`
//  		Metab     types.NullJSONStr   `db:"metab"`
//  	}
//  	func (b *Book) TableName() string {
//  		return "books"
//  	}

//  	bookTheHobbit := Book{
//  		Meta:      types.NullJSONStrFrom([]byte(`{"isdn":"123"}`)),
//  		Metab:     types.NullJSONStrFrom([]byte(`{"isdn":"123aaaaaaaa"}`)),
//  	}
//  	if err = db.Insert(&bookTheHobbit).Do(); err != nil {
//  		panic(fmt.Sprintf("can not insert book err: %v", err))
//  	}
//
//  	b := Book{}
//  	err = db.Select(&b).
//  		Where("id = ?", bookTheHobbit.Id).Do()
//  	if err == sql.ErrNoRows {
//  		fmt.Println("Book not found !")
//  	} else if err != nil {
//  		panic(fmt.Sprintf("can not get book err: %v", err))
//  	}
//  	fmt.Printf("------Meta: %v, Metabin: %v\n", b.Meta, b.Metab)
//  	meta := Meta{}
//  	err = b.Metab.Unmarshal(&meta)
//  	fmt.Printf("------Meta: %v, err: %v\n", meta, err)
//
//  	metaVal := ""
//  	metabVal := []byte("")
//  	err = db.SelectFrom("books"). //
//  		Columns("meta", "metab").
//  		Scanx(&metaVal, &metabVal)
//  	if err == sql.ErrNoRows {
//  		fmt.Println("Book not found !")
//  	} else if err != nil {
//  		panic(fmt.Sprintf("can not get book err: %v", err))
//  	}
//  	fmt.Printf("Meta: %v, Metabin: %v\n", metaVal, metabVal)
//

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
		return fmt.Errorf("Incompatible type for JSONStr")
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

// NullJSONStrFrom creates a valid NullJSONStr
func NullJSONStrFrom(dst []byte) NullJSONStr {
	return NullJSONStr{JSONStr: dst, Valid: true}
}
