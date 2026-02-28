package bondxls

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"

	"github.com/maciekmm/obligacje/internal/xlsconv"
)

const (
	indexURL  = "https://www.gov.pl/web/finanse/obligacje-detaliczne1"
	baseURL   = "https://www.gov.pl"
	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

var noRedirectClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func scrapeXLSURL(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, indexURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch index page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status fetching index page: %s", resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	href, found := findXLSAttachmentHref(doc)
	if !found {
		return "", fmt.Errorf("could not find XLS attachment link on page %s", indexURL)
	}

	// href is a relative path like /attachment/... — make it absolute.
	if strings.HasPrefix(href, "/") {
		return baseURL + href, nil
	}
	return href, nil
}

func findXLSAttachmentHref(n *html.Node) (string, bool) {
	if n.Type == html.ElementNode && n.Data == "a" {
		if hasClass(n, "file-download") && ariaLabelContains(n, "xls") {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					return attr.Val, true
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if href, ok := findXLSAttachmentHref(c); ok {
			return href, true
		}
	}
	return "", false
}

func hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			for _, c := range strings.Fields(attr.Val) {
				if c == class {
					return true
				}
			}
		}
	}
	return false
}

func ariaLabelContains(n *html.Node, sub string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "aria-label" {
			return strings.Contains(strings.ToLower(attr.Val), strings.ToLower(sub))
		}
	}
	return false
}

func downloadLatestBondXLS(ctx context.Context, output string) error {
	fileURL, err := scrapeXLSURL(ctx)
	if err != nil {
		return fmt.Errorf("failed to determine download URL: %w", err)
	}

	slog.Info("downloading bond file", "url", fileURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	out, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func DownloadLatestAndConvert(ctx context.Context, output string) error {
	// TODO: it might not be desired to create a temp dir in the output dir
	// the problem we're solving here is cross device rename though
	outputDir := filepath.Dir(output)
	slog.Info("downloading latest bond XLS", "outputDir", outputDir)

	tempFile, err := os.CreateTemp(outputDir, "data-*.xls")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFile.Close() // Close it since downloadLatestBondXLS will overwrite it
	xlsFile := tempFile.Name()
	defer os.Remove(xlsFile)

	if err := downloadLatestBondXLS(ctx, xlsFile); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	path, err := xlsconv.ToXLSX(xlsFile)
	if err != nil {
		return fmt.Errorf("failed to convert file: %w", err)
	}
	defer os.Remove(path) // Clean up the intermediate xlsx file as well

	slog.Info("converted XLS to XLSX", "path", path, "original", xlsFile)

	if err := os.Rename(path, output); err != nil {
		return fmt.Errorf("failed to move file to output: %w", err)
	}

	return nil
}
