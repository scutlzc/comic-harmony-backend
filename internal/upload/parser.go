package upload

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ArchiveEntry struct {
	Name string
	Data []byte
}

type ArchiveParser interface {
	Parse(path string) ([]ArchiveEntry, error)
}

func GetParser(ext string) ArchiveParser {
	switch strings.ToLower(ext) {
	case ".cbz", ".zip":
		return &zipParser{}
	case ".cbr", ".rar":
		return &rarParser{}
	case ".cb7", ".7z":
		return &sevenZipParser{}
	default:
		return nil
	}
}

// --- ZIP ---

type zipParser struct{}

func (p *zipParser) Parse(path string) ([]ArchiveEntry, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	var entries []ArchiveEntry
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !isImageFile(f.Name) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open entry %s: %w", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("read entry %s: %w", f.Name, err)
		}
		entries = append(entries, ArchiveEntry{
			Name: filepath.Base(f.Name),
			Data: data,
		})
	}
	return entries, nil
}

// --- RAR ---

type rarParser struct{}

func (p *rarParser) Parse(path string) ([]ArchiveEntry, error) {
	tmpDir, err := os.MkdirTemp("", "rar-extract-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("unrar", "e", "-y", path, tmpDir+"/")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("unrar failed: %s: %w", string(out), err)
	}

	var entries []ArchiveEntry
	entries, err = readDirAsEntries(tmpDir)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// --- 7z ---

type sevenZipParser struct{}

func (p *sevenZipParser) Parse(path string) ([]ArchiveEntry, error) {
	tmpDir, err := os.MkdirTemp("", "7z-extract-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("7z", "e", path, fmt.Sprintf("-o%s", tmpDir), "-y")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("7z failed: %s: %w", string(out), err)
	}

	return readDirAsEntries(tmpDir)
}

// --- Helpers ---

var imageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	".gif": true, ".bmp": true, ".avif": true, ".tiff": true, ".tif": true,
}

func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return imageExts[ext]
}

func readDirAsEntries(dir string) ([]ArchiveEntry, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var entries []ArchiveEntry
	for _, f := range files {
		if f.IsDir() || !isImageFile(f.Name()) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			continue
		}
		entries = append(entries, ArchiveEntry{
			Name: f.Name(),
			Data: data,
		})
	}
	return entries, nil
}

// --- EPUB ---

func ParseEPUB(path string) ([]ArchiveEntry, string, error) {
	// EPUB is a ZIP file with .epub extension
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, "", fmt.Errorf("open epub: %w", err)
	}
	defer r.Close()

	var entries []ArchiveEntry
	var title string

	for _, f := range r.File {
		name := f.Name

		// Read OPF for metadata
		if strings.HasSuffix(name, ".opf") {
			rc, _ := f.Open()
			data, _ := io.ReadAll(rc)
			rc.Close()
			title = extractEPUBTitle(string(data))
		}

		// Extract images from OEBPS/images or similar paths
		if isImageFile(name) {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}
			entries = append(entries, ArchiveEntry{
				Name: filepath.Base(name),
				Data: data,
			})
		}
	}

	return entries, title, nil
}

// PDF - render to images using ffmpeg
func ParsePDF(path string, outputDir string) ([]ArchiveEntry, error) {
	pattern := filepath.Join(outputDir, "page-%04d.png")
	cmd := exec.Command("ffmpeg", "-i", path,
		"-vf", "scale=1200:-1",
		"-frames:v", "500",
		pattern,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdf extract failed: %s: %w", string(out), err)
	}
	return readDirAsEntries(outputDir)
}

// MOBI - convert to EPUB via Calibre, then parse EPUB
func ParseMOBI(path string, outputDir string) ([]ArchiveEntry, string, error) {
	epubPath := filepath.Join(outputDir, "converted.epub")
	cmd := exec.Command("ebook-convert", path, epubPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("mobi convert failed: %s: %w", string(out), err)
	}
	return ParseEPUB(epubPath)
}

func extractEPUBTitle(opfContent string) string {
	// Simple title extraction from OPF XML
	start := strings.Index(opfContent, "<dc:title>")
	if start == -1 {
		start = strings.Index(opfContent, "<title>")
		if start == -1 {
			return ""
		}
		start += 7
	} else {
		start += 10
	}
	end := strings.Index(opfContent[start:], "</")
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(opfContent[start : start+end])
}
