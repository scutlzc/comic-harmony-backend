package komga

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	authHeader string
	httpClient *http.Client
}

func NewClient(baseURL, username, password string) *Client {
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return &Client{
		baseURL:    baseURL,
		authHeader: "Basic " + auth,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) doRequest(ctx context.Context, path string) (*http.Response, error) {
	u, _ := url.JoinPath(c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Accept", "application/json")
	return c.httpClient.Do(req)
}

// --- Komga API DTOs ---

type KomgaLibrary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type KomgaSeries struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Metadata    struct {
		Title       string `json:"title"`
		Status      string `json:"status"`
		Summary     string `json:"summary"`
		Authors     []struct {
			Name string `json:"name"`
			Role string `json:"role"`
		} `json:"authors"`
	} `json:"metadata"`
	BooksCount int    `json:"booksCount"`
	Created    string `json:"created"`
}

type KomgaBook struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SeriesID  string `json:"seriesId"`
	Number    float64 `json:"number"`
	PageCount int     `json:"pageCount"`
	Created   string  `json:"created"`
}

type KomgaPage struct {
	Number int    `json:"number"`
	URL    string `json:"-"`
}

// GetLibraries fetches all libraries from Komga.
func (c *Client) GetLibraries(ctx context.Context) ([]KomgaLibrary, error) {
	resp, err := c.doRequest(ctx, "/api/v1/libraries")
	if err != nil {
		return nil, fmt.Errorf("request libraries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("komga status %d", resp.StatusCode)
	}

	var libraries []KomgaLibrary
	if err := json.NewDecoder(resp.Body).Decode(&libraries); err != nil {
		return nil, fmt.Errorf("decode libraries: %w", err)
	}
	return libraries, nil
}

// GetSeries fetches series from a library.
func (c *Client) GetSeries(ctx context.Context, libraryID string, page, size int) ([]KomgaSeries, int, error) {
	path := fmt.Sprintf("/api/v1/series?library_id=%s&page=%d&size=%d", libraryID, page, size)
	resp, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Content []KomgaSeries `json:"content"`
		Total   int           `json:"totalElements"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}
	return result.Content, result.Total, nil
}

// GetBooks fetches books/chapters for a series.
func (c *Client) GetBooks(ctx context.Context, seriesID string) ([]KomgaBook, error) {
	path := fmt.Sprintf("/api/v1/series/%s/books?sort=number", seriesID)
	resp, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Content []KomgaBook `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Content, nil
}

// GetPageURL generates the URL for a specific page in a book.
func (c *Client) GetPageURL(bookID string, pageNumber int) string {
	return fmt.Sprintf("%s/api/v1/books/%s/pages/%d", c.baseURL, bookID, pageNumber)
}

// SearchSeries searches series by title.
func (c *Client) SearchSeries(ctx context.Context, query string, page, size int) ([]KomgaSeries, int, error) {
	path := fmt.Sprintf("/api/v1/series?search=%s&page=%d&size=%d", url.QueryEscape(query), page, size)
	resp, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Content []KomgaSeries `json:"content"`
		Total   int           `json:"totalElements"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}
	return result.Content, result.Total, nil
}

// ExtractAuthor extracts the first author name from series metadata.
func ExtractAuthor(series *KomgaSeries) string {
	for _, a := range series.Metadata.Authors {
		if a.Role == "writer" || a.Role == "artist" {
			return a.Name
		}
	}
	return ""
}
