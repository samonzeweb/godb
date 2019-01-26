package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StrSlice makes easy to handle JSON encoded string lists from/to db stored either in TEXT or BLOB.
//
// StrSlice is `[]string` type, adding `Value()` and `Scan()` methods for db access.
//
// Example:
//
//
// package main

// import (
// 	"encoding/json"
// 	"fmt"

// 	"github.com/samonzeweb/godb"
// 	"github.com/samonzeweb/godb/adapters/sqlite"
// 	"github.com/samonzeweb/godb/types"
// )

// // Country is db struct
// type Country struct {
// 	ID        int            `db:"id,key,auto"`
// 	Cities    types.StrSlice `db:"cities"`
// 	BigCities types.StrSlice `db:"big_cities"`
// }

// // TableName returns db tablename
// func (b *Country) TableName() string {
// 	return "countries"
// }

// func main() {
// 	db, err := godb.Open(sqlite.Adapter, "./countries.dat")
// 	if err != nil {
// 		panic(fmt.Sprintf("db connection err: %v", err))
// 	}
// 	if _, err = db.CurrentDB().Exec(`
// 		CREATE TABLE IF NOT EXISTS countries (
// 			id 			INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
// 			cities      TEXT NOT NULL,
// 			big_cities 	TEXT
// 		);`); err != nil {
// 		panic(err)
// 	}

// 	country := Country{
// 		Cities:    types.StrSlice([]string{"Amsterdam", "Antalya"}),
// 		BigCities: types.StrSlice([]string{"Istanbul", "Tokyo"}),
// 	}
// 	if err = db.Insert(&country).Do(); err != nil {
// 		panic(err)
// 	}

// 	// Select
// 	c := new(Country)
// 	if err = db.Select(c).Where("id = ?", country.ID).Do(); err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("Cities: %v, BigCities: %v\n", c.Cities, c.BigCities)

// 	res, err := json.MarshalIndent(c, "", "\t")
// 	fmt.Printf("Json : \n%s\n", res)
// 	// Scanning into variables
// 	country = Country{
// 		Cities:    types.StrSlice([]string{"Amsterdam", "Antalya"}),
// 		BigCities: nil,
// 	}
// 	if err = db.Insert(&country).Do(); err != nil {
// 		panic(err)
// 	}

// 	var cities types.StrSlice
// 	var bCities types.StrSlice
// 	if err = db.SelectFrom("countries").Columns("cities", "big_cities").Where("id = ?", country.ID).Scanx(&cities, &bCities); err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("Null Cities: %v, BigCities: %v\n", cities, bCities)
// }

// StrSlice makes easy to handle JSON encoded string lists stored at database's text fields(like VARCHAR,CHAR,TEXT) and blob fields
type StrSlice []string

// Value returns value.
// If value is invalid value, it returns an error.
func (ss StrSlice) Value() (driver.Value, error) {
	if ss == nil {
		return nil, nil
	}
	return json.Marshal(ss)
}

// Scan stores the value as StrSlice. Value can be string, []byte or or nil.
func (ss *StrSlice) Scan(src interface{}) error {
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
		return fmt.Errorf("incompatible type for StrSlice")
	}
	err := json.Unmarshal(source, ss)
	return err
}
