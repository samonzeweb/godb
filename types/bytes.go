package types

// NullBytes can be an []byte or a null value.
import (
	"database/sql/driver"
)

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

// NullBytesFrom creates a valid NullBytesFrom
func NullBytesFrom(v []byte) NullBytes {
	if v == nil {
		return 	NullBytes{Bytes: v, Valid: false}
	}
	return NullBytes{Bytes: v, Valid: true}
}