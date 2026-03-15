package discordv2

import "testing"

func TestContentTypeForFile(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Screenshot.jpeg", "image/jpeg"},
		{"Screenshot.JPEG", "image/jpeg"},
		{"photo.jpg", "image/jpeg"},
		{"photo.JPG", "image/jpeg"},
		{"image.png", "image/png"},
		{"image.PNG", "image/png"},
		{"data.bin", "application/octet-stream"},
		{"noext", "application/octet-stream"},
		{"", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contentTypeForFile(tt.name); got != tt.want {
				t.Errorf("contentTypeForFile(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
