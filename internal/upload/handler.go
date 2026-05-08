package upload

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/muyue/comic-harmony-backend/internal/response"
)

type UploadHandler struct {
	svc      *UploadService
	tempDir  string
	maxSize  int64
}

func NewUploadHandler(svc *UploadService, tempDir string, maxSizeMB int64) *UploadHandler {
	if maxSizeMB <= 0 {
		maxSizeMB = 500
	}
	return &UploadHandler{
		svc:     svc,
		tempDir: tempDir,
		maxSize: maxSizeMB << 20, // MB → bytes
	}
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)

	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		response.BadRequest(w, "file too large or invalid multipart")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "missing file field")
		return
	}
	defer file.Close()

	// --- Security checks ---

	// 1. Content-Type validation
	contentType := header.Header.Get("Content-Type")
	if !IsContentTypeAllowed(contentType) {
		actual := contentType
		if actual == "" {
			actual = "unknown"
		}
		response.BadRequest(w, fmt.Sprintf("unsupported content type: %s", actual))
		return
	}

	// 2. Extension check
	if !IsSupported(header.Filename) {
		ext := filepath.Ext(header.Filename)
		response.BadRequest(w, fmt.Sprintf("unsupported format: %s (supported: %v)", ext, SupportedFormats))
		return
	}

	// 3. Sanitize filename (prevent path traversal)
	safeName := SanitizeFilename(header.Filename)
	if safeName != header.Filename {
		response.BadRequest(w, "invalid filename")
		return
	}

	// Save uploaded file to temp
	ext := filepath.Ext(header.Filename)
	tmpFile := filepath.Join(h.tempDir, fmt.Sprintf("upload-%d%s", time.Now().UnixNano(), ext))
	dst, err := os.Create(tmpFile)
	if err != nil {
		response.InternalError(w, "failed to save upload")
		return
	}
	defer os.Remove(tmpFile)

	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		response.InternalError(w, "failed to write upload")
		return
	}
	dst.Close()

	// Process upload
	result, err := h.svc.ProcessUpload(r.Context(), tmpFile, safeName)
	if err != nil {
		response.BadRequest(w, fmt.Sprintf("processing failed: %v", err))
		return
	}

	response.Created(w, result)
}
