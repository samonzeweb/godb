package types

import (
	"database/sql"
	"time"
)

// NullStringFrom creates a valid NullString
// Deprecated: will be renamed in future version, use `ToNullString`
func NullStringFrom(v string) sql.NullString {
	return sql.NullString{String: v, Valid: true}
}

// NullFloat64From creates a valid NullFloat64
// Deprecated: will be renamed in future version, use `ToNullFloat64`
func NullFloat64From(v float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: v, Valid: true}
}

// NullInt64From creates a valid NullInt64
// Deprecated: will be renamed in future version, use `ToNullInt64`
func NullInt64From(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}

// NullBoolFrom creates a valid NullBool
// Deprecated: will be renamed in future version, use `ToNullBool`
func NullBoolFrom(v bool) sql.NullBool {
	return sql.NullBool{Bool: v, Valid: true}
}

// NullTimeFrom creates a valid NullTime
// Deprecated: will be renamed in future version, use `ToNullTime`
func NullTimeFrom(v time.Time) NullTime {
	return ToNullTime(v)
}

// NullJSONStrFrom creates a valid NullJSONStr
// Deprecated: will be renamed in future version, use `ToNullJSONStr`
func NullJSONStrFrom(dst []byte) NullJSONStr {
	return ToNullJSONStr(dst)
}

// NullCompactJSONStrFrom creates a valid NullCompactJSONStr
// Deprecated: will be renamed in future version, use `ToNullCompactJSONStr`
func NullCompactJSONStrFrom(dst []byte) NullCompactJSONStr {
	return ToNullCompactJSONStr(dst)
}

// NullBytesFrom creates a valid NullBytes
// Deprecated: will be renamed in future version, use `ToNullBytes`
func NullBytesFrom(v []byte) NullBytes {
	return ToNullBytes(v)
}
