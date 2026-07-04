package operations

import (
	"bytes"
	"testing"
)

type testMultipartFile struct {
	*bytes.Reader
}

func (file testMultipartFile) Close() error {
	return nil
}

func TestDetectMediaTypeRejectsSpoofedExtension(t *testing.T) {
	file := testMultipartFile{Reader: bytes.NewReader([]byte("plain text disguised as an image"))}

	_, _, err := detectMediaType(file, "avatar.jpg")
	if err == nil {
		t.Fatal("detectMediaType accepted text content with image extension")
	}
}

func TestDetectMediaTypeUsesContentSignature(t *testing.T) {
	pngHeader := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}
	contentType, extension, err := detectMediaType(testMultipartFile{Reader: bytes.NewReader(pngHeader)}, "upload.txt")
	if err != nil {
		t.Fatalf("detectMediaType png returned error: %v", err)
	}
	if contentType != "image/png" || extension != ".png" {
		t.Fatalf("detectMediaType png = %q/%q, want image/png/.png", contentType, extension)
	}
}

func TestDetectMediaTypeRecognizesWebPHeader(t *testing.T) {
	webpHeader := []byte{'R', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'E', 'B', 'P', 'V', 'P', '8', ' '}
	contentType, extension, err := detectMediaType(testMultipartFile{Reader: bytes.NewReader(webpHeader)}, "cover.bin")
	if err != nil {
		t.Fatalf("detectMediaType webp returned error: %v", err)
	}
	if contentType != "image/webp" || extension != ".webp" {
		t.Fatalf("detectMediaType webp = %q/%q, want image/webp/.webp", contentType, extension)
	}
}

func TestSafeOriginalNameUsesBaseName(t *testing.T) {
	if got := safeOriginalName("../cover:image.jpg"); got != "cover-image.jpg" {
		t.Fatalf("safeOriginalName returned %q, want cover-image.jpg", got)
	}
}
