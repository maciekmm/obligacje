package bondxls

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/maciekmm/obligacje/internal/xlsconv"
)

const (
	fileURL = "https://api.dane.gov.pl/resources/765987,sprzedaz-obligacji-detalicznych/file"
)

func downloadLatestBondXLS(ctx context.Context, output string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
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

	xlsFile := filepath.Join(outputDir, "data.xls")

	if err := downloadLatestBondXLS(ctx, xlsFile); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	path, err := xlsconv.ToXLSX(xlsFile)
	if err != nil {
		return fmt.Errorf("failed to convert file: %w", err)
	}

	slog.Info("converted XLS to XLSX", "path", path, "original", xlsFile)

	if err := os.Rename(path, output); err != nil {
		return fmt.Errorf("failed to move file to output: %w", err)
	}

	if err := os.Remove(xlsFile); err != nil {
		return fmt.Errorf("failed to remove XLS file: %w", err)
	}

	return nil
}
