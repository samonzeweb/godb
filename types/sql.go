package types

import (
	"database/sql"
	"encoding/json"
)

// NullString is a type that can be null or a string, wrapped for JSON encoding/decoding
type NullString struct {
	sql.NullString
}

// NullFloat64 is a type that can be null or a float64, wrapped for JSON encoding/decoding
type NullFloat64 struct {
	sql.NullFloat64
}

// NullInt64 is a type that can be null or an int, wrapped for JSON encoding/decoding
type NullInt64 struct {
	sql.NullInt64
}

// NullBool is a type that can be null or a bool, wrapped for JSON encoding/decoding
type NullBool struct {
	sql.NullBool
}

// NullStringFrom creates a valid NullString
func NullStringFrom(v string) NullString {
	return NullString{sql.NullString{String: v, Valid: true}}
}

// MarshalJSON serializes a NullString to JSON
func (n NullString) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.String)
		return j, e
	}
	return []byte("null"), nil
}

// UnmarshalJSON parses NullString from JSON
func (n *NullString) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// NullFloat64From creates a valid NullFloat64
func NullFloat64From(v float64) NullFloat64 {
	return NullFloat64{sql.NullFloat64{Float64: v, Valid: true}}
}

// MarshalJSON serializes a NullFloat64 to JSON
func (n NullFloat64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.Float64)
		return j, e
	}
	return []byte("null"), nil
}

// UnmarshalJSON parses NullFloat64 from JSON
func (n *NullFloat64) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// NullInt64From creates a valid NullInt64
func NullInt64From(v int64) NullInt64 {
	return NullInt64{sql.NullInt64{Int64: v, Valid: true}}
}

// MarshalJSON NullInt64 to JSON
func (n NullInt64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.Int64)
		return j, e
	}
	return []byte("null"), nil
}

// UnmarshalJSON parses NullInt64 from JSON
func (n *NullInt64) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// NullBoolFrom creates a valid NullBool
func NullBoolFrom(v bool) NullBool {
	return NullBool{sql.NullBool{Bool: v, Valid: true}}
}

// MarshalJSON serializes a NullBool to JSON
func (n NullBool) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.Bool)
		return j, e
	}
	return []byte("null"), nil
}

// UnmarshalJSON parses NullBool from JSON
func (n *NullBool) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}
