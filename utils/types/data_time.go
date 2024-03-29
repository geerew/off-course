package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/spf13/cast"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DefaultDateLayout specifies the default app date strings layout
const DefaultDateLayout = "2006-01-02 15:04:05.000Z"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NowDateTime returns new date time instance with the current local time
func NowDateTime() DateTime {
	return DateTime{t: time.Now()}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ParseDateTime creates a new date time from the provided value
func ParseDateTime(value any) (DateTime, error) {
	d := DateTime{}
	err := d.Scan(value)
	return d, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DateTime represents a [time.Time] instance in UTC that is wrapped and serialized using the app
// default date layout
type DateTime struct {
	t time.Time
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Time returns the internal [time.Time] instance
func (d DateTime) Time() time.Time {
	return d.t
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsZero checks whether the current DateTime instance has zero time value
func (d DateTime) IsZero() bool {
	return d.Time().IsZero()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String serializes the current DateTime instance into a formatted UTC date string
//
// The zero value is serialized to an empty string
func (d DateTime) String() string {
	if d.IsZero() {
		return ""
	}
	return d.Time().UTC().Format(DefaultDateLayout)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// MarshalJSON implements the `json.Marshaler` interface
func (d DateTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UnmarshalJSON implements the `json.Unmarshaler` interface
func (d *DateTime) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	return d.Scan(raw)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Value implements the `driver.Valuer` interface
func (d DateTime) Value() (driver.Value, error) {
	return d.String(), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan implements `sql.Scanner“ interface to scan the provided value into the current date time
// instance
func (d *DateTime) Scan(value any) error {
	switch v := value.(type) {
	case DateTime:
		d.t = v.Time()
	case time.Time:
		d.t = v
	case int:
		d.t = cast.ToTime(v)
	case string:
		if v == "" {
			d.t = time.Time{}
		} else {
			t, err := time.Parse(DefaultDateLayout, v)
			if err != nil {
				t = cast.ToTime(v)
			}
			d.t = t
		}
	default:
		str := cast.ToString(v)
		if str == "" {
			d.t = time.Time{}
		} else {
			d.t = cast.ToTime(str)
		}
	}

	return nil
}
