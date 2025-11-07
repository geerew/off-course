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

// DateTime represents a [time.Time] instance in UTC that is serialized
// using the app default date layout
type DateTime time.Time

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NowDateTime returns new DateTime instance with the current local time
func NowDateTime() DateTime {
	return DateTime(time.Now())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ParseDateTime creates a new DateTime from the provided value
func ParseDateTime(value any) (DateTime, error) {
	var d DateTime
	err := d.Scan(value)
	return d, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Time returns the internal [time.Time] instance
func (d DateTime) Time() time.Time {
	return time.Time(d)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsZero checks whether the current DateTime instance has zero time value
func (d DateTime) IsZero() bool {
	return d.Time().IsZero()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Equal checks if two DateTime instances represent the same point in time
func (d DateTime) Equal(other DateTime) bool {
	return time.Time(d).UTC().Truncate(time.Millisecond).Equal(time.Time(other).UTC().Truncate(time.Millisecond))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String serializes the current DateTime instance into a formatted
// UTC date string
//
// A zero value is serialized to an empty string
func (d DateTime) String() string {
	t := d.Time()
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(DefaultDateLayout)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// MarshalJSON implements the [json.Marshaler] interface
func (d DateTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UnmarshalJSON implements the [json.Unmarshaler] interface
func (d *DateTime) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	return d.Scan(raw)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Value implements the [driver.Valuer] interface
func (d DateTime) Value() (driver.Value, error) {
	return d.String(), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan implements [sql.Scanner] interface to scan the provided value
// into the current DateTime instance
func (d *DateTime) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		*d = DateTime(v)
	case DateTime:
		*d = v
	case string:
		if v == "" {
			*d = DateTime(time.Time{})
		} else {
			t, err := time.Parse(DefaultDateLayout, v)
			if err != nil {
				// check for other common date layouts
				t = cast.ToTime(v)
			}
			*d = DateTime(t)
		}
	case int, int64, int32, uint, uint64, uint32:
		*d = DateTime(cast.ToTime(v))
	default:
		str := cast.ToString(v)
		if str == "" {
			*d = DateTime(time.Time{})
		} else {
			*d = DateTime(cast.ToTime(str))
		}
	}

	return nil
}
