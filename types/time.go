package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// NullTime is a type that can be null or a time
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the time scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	if !nt.Valid && value != nil {
		return fmt.Errorf("invalid type %T for NullTime: %v", value, value)
	}
	return nil
}

// Value implements the time driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// ToNullTime creates a valid NullTime
func ToNullTime(v time.Time) NullTime {
	return NullTime{Time: v, Valid: true}
}

// MarshalJSON serializes a NullTime to JSON.
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if nt.Valid {
		return nt.Time.MarshalJSON()
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes a NullTime from JSON.
func (nt *NullTime) UnmarshalJSON(b []byte) (err error) {
	// scan for null
	if bytes.Equal(b, []byte("null")) {
		return nt.Scan(nil)
	}
	var t time.Time
	var err2 error
	if err := json.Unmarshal(b, &t); err != nil {
		// Try for JS new Date().toJSON()
		if t, err2 = time.Parse("2006-01-02T15:04:05.000Z", string(b)); err2 != nil {
			return err
		}
	}
	return nt.Scan(t)
}
