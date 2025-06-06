package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cast"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DescriptionType defines the type of description
type DescriptionType string

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Description defines the model for a description
type Description struct {
	s DescriptionType
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	DescriptionMarkdown    DescriptionType = "markdown"
	DescriptionText        DescriptionType = "text"
	DescriptionUnsupported DescriptionType = "unsupported"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewDescription creates a new description based on the file extension
func NewDescription(ext string) *Description {
	switch strings.ToLower(ext) {
	case "md":
		return &Description{s: DescriptionMarkdown}
	case "txt":
		return &Description{s: DescriptionText}
	}

	return &Description{s: DescriptionUnsupported}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetMarkdown sets the description type to Markdown
func (d *Description) SetMarkdown() {
	d.s = DescriptionMarkdown
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsMarkdown returns true is the description is of type Markdown
func (d Description) IsMarkdown() bool {
	return d.s == DescriptionMarkdown
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetText sets the description type to Text
func (d *Description) SetText() {
	d.s = DescriptionText
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsText returns true is the description is of type Text
func (d Description) IsText() bool {
	return d.s == DescriptionText
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsSupported returns true if the description is of type Markdown or Text
func (d Description) IsSupported() bool {
	return !d.IsUnsupported()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsUnsupported returns true if the description is of type Unsupported
func (d Description) IsUnsupported() bool {
	return d.s == DescriptionUnsupported || d.s == ""
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String implements the `Stringer` interface
func (d Description) String() string {
	return fmt.Sprint(d.s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type returns the type of the description
func (d Description) Type() DescriptionType {
	return d.s
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// MarshalJSON implements the `json.Marshaler` interface
func (d Description) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.s + `"`), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UnmarshalJSON implements the `json.Unmarshaler` interface
func (d *Description) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	return d.Scan(raw)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Value implements the `driver.Valuer` interface
func (d Description) Value() (driver.Value, error) {
	return d.String(), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan implements `sql.Scanner` interface
func (d *Description) Scan(value any) error {
	vv := cast.ToString(value)

	switch vv {
	case string(DescriptionMarkdown):
		d.s = DescriptionMarkdown
	case string(DescriptionText):
		d.s = DescriptionText
	default:
		d.s = DescriptionUnsupported
	}

	return nil
}
