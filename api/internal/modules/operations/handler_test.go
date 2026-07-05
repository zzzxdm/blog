package operations

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blog/api/internal/modules/auth"

	"github.com/gin-gonic/gin"
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

func TestAdminOperationsRoutes(t *testing.T) {
	authStore := auth.NewMemoryStore()
	_, token, err := authStore.Authenticate("admin@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}

	router := gin.New()
	router.Use(auth.Middleware(authStore))
	RegisterRoutes(router, NewMemoryRepository(), "")

	redirectsReq := httptest.NewRequest(http.MethodGet, "/admin/redirects", nil)
	redirectsReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	redirectsRec := httptest.NewRecorder()
	router.ServeHTTP(redirectsRec, redirectsReq)
	if redirectsRec.Code != http.StatusOK || !strings.Contains(redirectsRec.Body.String(), `"total":1`) {
		t.Fatalf("expected redirects response, got status=%d body=%q", redirectsRec.Code, redirectsRec.Body.String())
	}

	replaceReq := httptest.NewRequest(http.MethodPut, "/admin/redirects", bytes.NewBufferString(`{"items":[{"from":"/legacy","to":"/archive","code":302}]}`))
	replaceReq.Header.Set("Content-Type", "application/json")
	replaceReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	replaceRec := httptest.NewRecorder()
	router.ServeHTTP(replaceRec, replaceReq)
	if replaceRec.Code != http.StatusOK || !strings.Contains(replaceRec.Body.String(), `"/legacy"`) {
		t.Fatalf("expected redirects replace response, got status=%d body=%q", replaceRec.Code, replaceRec.Body.String())
	}

	statsReq := httptest.NewRequest(http.MethodGet, "/admin/statistics?range=7d", nil)
	statsReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	statsRec := httptest.NewRecorder()
	router.ServeHTTP(statsRec, statsReq)
	if statsRec.Code != http.StatusOK || !strings.Contains(statsRec.Body.String(), `"range":"7d"`) {
		t.Fatalf("expected statistics response, got status=%d body=%q", statsRec.Code, statsRec.Body.String())
	}

	exportReq := httptest.NewRequest(http.MethodPost, "/admin/export", bytes.NewBufferString(`{"scope":"users"}`))
	exportReq.Header.Set("Content-Type", "application/json")
	exportReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	exportRec := httptest.NewRecorder()
	router.ServeHTTP(exportRec, exportReq)
	if exportRec.Code != http.StatusAccepted {
		t.Fatalf("expected export job status 202, got %d body=%q", exportRec.Code, exportRec.Body.String())
	}

	var exportJob AdminJob
	if err := json.Unmarshal(exportRec.Body.Bytes(), &exportJob); err != nil {
		t.Fatalf("decode export job: %v", err)
	}
	if exportJob.Status != "completed" || exportJob.Progress != 100 {
		t.Fatalf("unexpected export job: %+v", exportJob)
	}

	jobReq := httptest.NewRequest(http.MethodGet, "/admin/jobs/"+exportJob.ID, nil)
	jobReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	jobRec := httptest.NewRecorder()
	router.ServeHTTP(jobRec, jobReq)
	if jobRec.Code != http.StatusOK || !strings.Contains(jobRec.Body.String(), exportJob.ID) {
		t.Fatalf("expected job lookup response, got status=%d body=%q", jobRec.Code, jobRec.Body.String())
	}

	importReq := httptest.NewRequest(http.MethodPost, "/admin/import", bytes.NewBufferString(`{"scope":"posts","fileName":"posts.json"}`))
	importReq.Header.Set("Content-Type", "application/json")
	importReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	importRec := httptest.NewRecorder()
	router.ServeHTTP(importRec, importReq)
	if importRec.Code != http.StatusAccepted || !strings.Contains(importRec.Body.String(), `"status":"queued"`) {
		t.Fatalf("expected import job response, got status=%d body=%q", importRec.Code, importRec.Body.String())
	}
}
