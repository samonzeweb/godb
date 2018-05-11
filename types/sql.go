package types

import (
	"database/sql"
)

// NullStringFrom creates a valid NullString
func NullStringFrom(v string) sql.NullString {
	return sql.NullString{String: v, Valid: true}
}

// NullFloat64From creates a valid NullFloat64
func NullFloat64From(v float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: v, Valid: true}
}

// NullInt64From creates a valid NullInt64
func NullInt64From(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}

// NullBoolFrom creates a valid NullBool
func NullBoolFrom(v bool) sql.NullBool {
	return sql.NullBool{Bool: v, Valid: true}
}
