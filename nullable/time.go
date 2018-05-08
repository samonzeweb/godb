package nullable


// NullTime is a type that can be null or a time
import (
	"time"
	"database/sql/driver"
	"fmt"
)

type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the time scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	if !nt.Valid {
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


// NullTimeFrom creates a valid NullTime
func NullTimeFrom(v time.Time) NullTime {
	return NullTime{Time: v, Valid: true}
}
