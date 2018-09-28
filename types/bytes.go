package types

import (
	"database/sql/driver"
)

// NullBytes can be an []byte or a null value.
type NullBytes struct {
	Bytes []byte
	Valid bool // Valid is true if Bytes is not NULL
}

// Scan implements the Scanner interface.
func (n *NullBytes) Scan(value interface{}) error {
	n.Bytes, n.Valid = value.([]byte)
	return nil
}

// Value implements the driver Valuer interface.
func (n NullBytes) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Bytes, nil
}

// ToNullBytes creates a valid NullBytes
func ToNullBytes(v []byte) NullBytes {
	if v == nil {
		return NullBytes{Bytes: v, Valid: false}
	}
	return NullBytes{Bytes: v, Valid: true}
}
