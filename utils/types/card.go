package types

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CardExtension represents a valid card image file extension
type CardExtension string

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	CardExtensionJPG  CardExtension = "jpg"
	CardExtensionJPEG CardExtension = "jpeg"
	CardExtensionPNG  CardExtension = "png"
	CardExtensionWebP CardExtension = "webp"
	CardExtensionTIFF CardExtension = "tiff"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// validCardExtensions contains all valid card extensions
var validCardExtensions = map[CardExtension]bool{
	CardExtensionJPG:  true,
	CardExtensionJPEG: true,
	CardExtensionPNG:  true,
	CardExtensionWebP: true,
	CardExtensionTIFF: true,
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsValid returns true if the extension is a valid card extension
func (c CardExtension) IsValid() bool {
	return validCardExtensions[c]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String implements the `Stringer` interface
func (c CardExtension) String() string {
	return string(c)
}
