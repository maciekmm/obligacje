package bondxls

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"io"
)

const (
	fileURL = "https://api.dane.gov.pl/resources/765987,sprzedaz-obligacji-detalicznych/file"
)

func DownloadLatestBondXLS(ctx context.Context, output string) error {
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
