package upload

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ImageConverter interface {
	ToWebP(input []byte, quality int) ([]byte, error)
	Resize(input []byte, maxWidth int) ([]byte, error)
}

type ffmpegConverter struct {
	tempDir string
}

func NewImageConverter(tempDir string) ImageConverter {
	return &ffmpegConverter{tempDir: tempDir}
}

func (c *ffmpegConverter) ToWebP(input []byte, quality int) ([]byte, error) {
	if quality <= 0 || quality > 100 {
		quality = 80
	}

	inPath := filepath.Join(c.tempDir, "input-tmp")
	outPath := filepath.Join(c.tempDir, "output.webp")

	if err := os.WriteFile(inPath, input, 0644); err != nil {
		return nil, fmt.Errorf("write temp input: %w", err)
	}
	defer os.Remove(inPath)
	defer os.Remove(outPath)

	cmd := exec.Command("ffmpeg", "-y", "-i", inPath,
		"-vf", "scale=1200:-1:flags=lanczos",
		"-q:v", fmt.Sprintf("%d", quality),
		"-compression_level", "6",
		outPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg webp failed: %s: %w", string(out), err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("read output: %w", err)
	}
	return data, nil
}

func (c *ffmpegConverter) Resize(input []byte, maxWidth int) ([]byte, error) {
	inPath := filepath.Join(c.tempDir, "resize-input")
	outPath := filepath.Join(c.tempDir, "resized.webp")

	if err := os.WriteFile(inPath, input, 0644); err != nil {
		return nil, fmt.Errorf("write temp input: %w", err)
	}
	defer os.Remove(inPath)
	defer os.Remove(outPath)

	cmd := exec.Command("ffmpeg", "-y", "-i", inPath,
		"-vf", fmt.Sprintf("scale=%d:-1:flags=lanczos", maxWidth),
		"-q:v", "70",
		"-compression_level", "6",
		outPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg resize failed: %s: %w", string(out), err)
	}

	return os.ReadFile(outPath)
}

// DetectFormat detects image format from magic bytes
func DetectFormat(data []byte) string {
	if len(data) < 4 {
		return "unknown"
	}
	switch {
	case data[0] == 0xFF && data[1] == 0xD8:
		return "jpeg"
	case data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G':
		return "png"
	case data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F':
		return "webp"
	case data[0] == 'G' && data[1] == 'I' && data[2] == 'F':
		return "gif"
	case data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x1C:
		return "avif"
	default:
		return "unknown"
	}
}

var SupportedFormats = []string{".cbz", ".cbr", ".cb7", ".zip", ".rar", ".epub", ".pdf", ".mobi", ".jpg", ".jpeg", ".png", ".webp"}

func IsSupported(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, s := range SupportedFormats {
		if ext == s {
			return true
		}
	}
	return false
}
