package upload

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/muyue/comic-harmony-backend/internal/model"
	"github.com/muyue/comic-harmony-backend/internal/repository"
)

type UploadService struct {
	comicRepo   repository.ComicRepository
	chapterRepo repository.ChapterRepository
	converter   ImageConverter
	storage     *LocalStorage
	tempDir     string
}

func NewUploadService(
	comicRepo repository.ComicRepository,
	chapterRepo repository.ChapterRepository,
	converter ImageConverter,
	storage *LocalStorage,
	tempDir string,
) *UploadService {
	os.MkdirAll(tempDir, 0755)
	return &UploadService{
		comicRepo:   comicRepo,
		chapterRepo: chapterRepo,
		converter:   converter,
		storage:     storage,
		tempDir:     tempDir,
	}
}

type UploadResult struct {
	Comic   *model.Comic   `json:"comic"`
	Chapter *model.Chapter `json:"chapter"`
	Pages   int            `json:"pages"`
}

func (s *UploadService) ProcessUpload(ctx context.Context, filePath, originalName string) (*UploadResult, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	log.Printf("[upload] processing %s (ext=%s)", originalName, ext)

	var entries []ArchiveEntry
	var title string
	var err error

	switch ext {
	case ".cbz", ".zip":
		var parser ArchiveParser
		parser = GetParser(ext)
		if parser == nil {
			return nil, fmt.Errorf("unsupported format: %s", ext)
		}
		entries, err = parser.Parse(filePath)
		if err != nil {
			return nil, fmt.Errorf("parse archive: %w", err)
		}
		title = strings.TrimSuffix(originalName, ext)

	case ".cbr", ".rar", ".cb7":
		parser := GetParser(ext)
		if parser == nil {
			return nil, fmt.Errorf("unsupported format: %s", ext)
		}
		entries, err = parser.Parse(filePath)
		if err != nil {
			return nil, fmt.Errorf("parse archive: %w", err)
		}
		title = strings.TrimSuffix(originalName, ext)

	case ".epub":
		var epubTitle string
		entries, epubTitle, err = ParseEPUB(filePath)
		if err != nil {
			return nil, fmt.Errorf("parse epub: %w", err)
		}
		title = epubTitle
		if title == "" {
			title = strings.TrimSuffix(originalName, ext)
		}

	case ".pdf":
		pdfDir, err := os.MkdirTemp(s.tempDir, "pdf-*")
		if err != nil {
			return nil, fmt.Errorf("create pdf temp: %w", err)
		}
		defer os.RemoveAll(pdfDir)
		entries, err = ParsePDF(filePath, pdfDir)
		if err != nil {
			return nil, fmt.Errorf("parse pdf: %w", err)
		}
		title = strings.TrimSuffix(originalName, ext)

	case ".mobi":
		mobiDir, err := os.MkdirTemp(s.tempDir, "mobi-*")
		if err != nil {
			return nil, fmt.Errorf("create mobi temp: %w", err)
		}
		defer os.RemoveAll(mobiDir)
		var mobiTitle string
		entries, mobiTitle, err = ParseMOBI(filePath, mobiDir)
		if err != nil {
			return nil, fmt.Errorf("parse mobi: %w", err)
		}
		title = mobiTitle
		if title == "" {
			title = strings.TrimSuffix(originalName, ext)
		}

	default:
		if isImageFile(originalName) {
			var data []byte
			data, err = os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			entries = []ArchiveEntry{{Name: originalName, Data: data}}
			title = strings.TrimSuffix(originalName, ext)
		} else {
			return nil, fmt.Errorf("unsupported format: %s", ext)
		}
	}

	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no images found in %s", originalName)
	}

	// Sort entries by name
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	// Create comic record
	comicReq := model.CreateComicRequest{
		Title:      title,
		Author:     "unknown",
		CategoryID: 6, // 未分类
	}
	comic, err := s.comicRepo.Create(ctx, comicReq)
	if err != nil {
		return nil, fmt.Errorf("create comic: %w", err)
	}

	// Convert and save pages
	var pageURLs []string
	for i, entry := range entries {
		webpData, convErr := s.converter.ToWebP(entry.Data, 80)
		if convErr != nil {
			log.Printf("[upload] convert page %d failed: %v (using original)", i, convErr)
			url, saveErr := s.storage.Save(
				fmt.Sprintf("%d/%04d.webp", comic.ID, i+1),
				entry.Data,
			)
			if saveErr != nil {
				return nil, fmt.Errorf("save page %d: %w", i, saveErr)
			}
			pageURLs = append(pageURLs, url)
			continue
		}

		url, saveErr := s.storage.Save(
			fmt.Sprintf("%d/%04d.webp", comic.ID, i+1),
			webpData,
		)
		if saveErr != nil {
			return nil, fmt.Errorf("save page %d: %w", i, saveErr)
		}
		pageURLs = append(pageURLs, url)
	}

	// Create chapter
	chapter := &model.Chapter{
		ComicID:   comic.ID,
		Title:     "第1话",
		SortOrder: 1,
		PageCount: len(pageURLs),
	}
	if err := s.chapterRepo.Create(ctx, chapter); err != nil {
		return nil, fmt.Errorf("create chapter: %w", err)
	}

	// Set cover URL to first page
	if len(pageURLs) > 0 {
		comic.CoverURL = pageURLs[0]
	}

	log.Printf("[upload] done: comic=%d chapter=%d pages=%d", comic.ID, chapter.ID, len(pageURLs))

	return &UploadResult{
		Comic:   comic,
		Chapter: chapter,
		Pages:   len(pageURLs),
	}, nil
}
