package upload

import (
	"path/filepath"
	"strings"
)

// SanitizeFilename removes path separators and limits dangerous characters.
func SanitizeFilename(name string) string {
	// Remove path traversal
	name = filepath.Clean(name)
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")

	// Only keep the base name
	name = filepath.Base(name)

	// Remove null bytes
	name = strings.ReplaceAll(name, "\x00", "")

	if name == "" || name == "." {
		return "upload.bin"
	}
	return name
}

// AllowedMimeTypes for comic uploads
var AllowedMimeTypes = map[string]string{
	"application/zip":                  ".cbz",
	"application/x-rar-compressed":     ".cbr",
	"application/x-7z-compressed":      ".cb7",
	"application/x-cbz":               ".cbz",
	"application/x-cbr":               ".cbr",
	"application/epub+zip":             ".epub",
	"application/pdf":                  ".pdf",
	"application/vnd.amazon.mobi8-ebook": ".mobi",
	"application/x-mobipocket-ebook":   ".mobi",
	"image/jpeg":                       ".jpg",
	"image/png":                        ".png",
}

// IsContentTypeAllowed checks if the Content-Type is a supported format.
func IsContentTypeAllowed(contentType string) bool {
	_, ok := AllowedMimeTypes[contentType]
	return ok
}
