package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cast"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetType defines the type of asset
type AssetType string

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Asset struct {
	s AssetType
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	AssetVideo    AssetType = "video"
	AssetHTML     AssetType = "html"
	AssetPDF      AssetType = "pdf"
	AssetMarkdown AssetType = "markdown"
	AssetText     AssetType = "text"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAsset creates an Asset based upon an extension. For example "mp4" => AssetVideo. When an
// unknown extension is passed in, nil is returned
func NewAsset(ext string) *Asset {
	switch strings.ToLower(ext) {
	case "avi",
		"mkv",
		"flac",
		"mp4",
		"m4a",
		"mp3",
		"ogv",
		"ogm",
		"ogg",
		"oga",
		"opus",
		"webm",
		"wav":
		return &Asset{s: AssetVideo}
	case "htm", "html":
		return &Asset{s: AssetHTML}
	case "pdf":
		return &Asset{s: AssetPDF}
	case "md":
		return &Asset{s: AssetMarkdown}
	case "txt":
		return &Asset{s: AssetText}
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetVideo sets the asset type to video
func (a *Asset) SetVideo() {
	a.s = AssetVideo
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsVideo returns true is the asset is of type video
func (a Asset) IsVideo() bool {
	return a.s == AssetVideo
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetHTML sets the asset type to HTML
func (a *Asset) SetHTML() {
	a.s = AssetHTML
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsHTML returns true is the asset is of type HTML
func (a Asset) IsHTML() bool {
	return a.s == AssetHTML
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetPDF sets the asset type to PDF
func (a *Asset) SetPDF() {
	a.s = AssetPDF
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsPDF returns true is the asset is of type PDF
func (a Asset) IsPDF() bool {
	return a.s == AssetPDF
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetMarkdown sets the asset type to Markdown
func (a *Asset) SetMarkdown() {
	a.s = AssetMarkdown
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsMarkdown returns true is the asset is of type Markdown
func (a Asset) IsMarkdown() bool {
	return a.s == AssetMarkdown
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetText sets the asset type to Text
func (a *Asset) SetText() {
	a.s = AssetText
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsText returns true is the asset is of type Text
func (a Asset) IsText() bool {
	return a.s == AssetText
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String implements the `Stringer` interface
func (a Asset) String() string {
	return fmt.Sprint(a.s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type returns the type of the asset
func (a Asset) Type() AssetType {
	return a.s
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// MarshalJSON implements the `json.Marshaler` interface
func (a Asset) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.s + `"`), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UnmarshalJSON implements the `json.Unmarshaler` interface
func (a *Asset) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	return a.Scan(raw)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Value implements the `driver.Valuer` interface
func (a Asset) Value() (driver.Value, error) {
	return a.String(), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan implements `sql.Scanner` interface
func (a *Asset) Scan(value any) error {
	vv := cast.ToString(value)

	switch vv {
	case string(AssetVideo):
		a.s = AssetVideo
	case string(AssetHTML):
		a.s = AssetHTML
	case string(AssetPDF):
		a.s = AssetPDF
	case string(AssetMarkdown):
		a.s = AssetMarkdown
	case string(AssetText):
		a.s = AssetText
	default:
		return errors.New("invalid asset type")
	}

	return nil
}
