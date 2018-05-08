package nullable

import (
	"database/sql"
)

// NullString is a type that can be null or a string
type NullString struct {
	sql.NullString
}

// NullFloat64 is a type that can be null or a float64
type NullFloat64 struct {
	sql.NullFloat64
}

// NullInt64 is a type that can be null or an int
type NullInt64 struct {
	sql.NullInt64
}

// NullBool is a type that can be null or a bool
type NullBool struct {
	sql.NullBool
}

// NullStringFrom creates a valid NullString
func NullStringFrom(v string) NullString {
	return NullString{sql.NullString{String: v, Valid: true}}
}

// NullFloat64From creates a valid NullFloat64
func NullFloat64From(v float64) NullFloat64 {
	return NullFloat64{sql.NullFloat64{Float64: v, Valid: true}}
}

// NullInt64From creates a valid NullInt64
func NullInt64From(v int64) NullInt64 {
	return NullInt64{sql.NullInt64{Int64: v, Valid: true}}
}

// NullBoolFrom creates a valid NullBool
func NullBoolFrom(v bool) NullBool {
	return NullBool{sql.NullBool{Bool: v, Valid: true}}
}