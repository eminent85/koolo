package discordv2

import (
	"bytes"
	"strings"
)

// newByteReader wraps a byte slice in a *bytes.Reader for use as an io.Reader.
func newByteReader(data []byte) *bytes.Reader {
	return bytes.NewReader(data)
}

// contentTypeForFile returns a MIME content type based on file extension.
func contentTypeForFile(name string) string {
	if strings.HasSuffix(strings.ToLower(name), ".jpeg") || strings.HasSuffix(strings.ToLower(name), ".jpg") {
		return "image/jpeg"
	}
	if strings.HasSuffix(strings.ToLower(name), ".png") {
		return "image/png"
	}
	return "application/octet-stream"
}
