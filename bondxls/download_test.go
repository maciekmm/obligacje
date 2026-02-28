package bondxls

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maciekmm/obligacje/internal/xlsconv"
	"github.com/maciekmm/obligacje/tz"
)

func TestDownloadLatestBondXLS_FindsFile(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI as gov website is flaky, we have proper retry logic in source.go")
	}
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "bonds.xls")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := downloadLatestBondXLS(ctx, outputFile)
	if err != nil {
		t.Fatalf("DownloadLatestBondXLS() error = %v", err)
	}

	info, err := os.Stat(outputFile)

	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("Downloaded file is empty")
	}

	t.Logf("Successfully downloaded file: %s (size: %d bytes)", outputFile, info.Size())
}

func TestDownloadLatestBondXLS_ConvertsToXLSX(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI as gov website is flaky, we have proper retry logic in source.go")
	}
	tmpDir := t.TempDir()
	xlsFile := filepath.Join(tmpDir, "bonds.xls")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := downloadLatestBondXLS(ctx, xlsFile)
	if err != nil {
		t.Fatalf("DownloadLatestBondXLS() error = %v", err)
	}

	xlsxFile, err := xlsconv.ToXLSX(xlsFile)
	if err != nil {
		t.Fatalf("ToXLSX() error = %v", err)
	}

	if _, err := os.Stat(xlsxFile); os.IsNotExist(err) {
		t.Fatalf("Converted XLSX file does not exist: %s", xlsxFile)
	}

	info, err := os.Stat(xlsxFile)
	if err != nil {
		t.Fatalf("Failed to stat XLSX file: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("Converted XLSX file is empty")
	}

	t.Logf("Successfully converted to XLSX: %s (size: %d bytes)", xlsxFile, info.Size())
}

func TestDownloadLatestAndConvert(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI as gov website is flaky, we have proper retry logic in source.go")
	}
	tmpDir := t.TempDir()
	xlsxFile := filepath.Join(tmpDir, "bonds.xlsx")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := DownloadLatestAndConvert(ctx, xlsxFile)
	if err != nil {
		t.Fatalf("DownloadLatestAndConvert() error = %v", err)
	}

	if _, err := os.Stat(xlsxFile); os.IsNotExist(err) {
		t.Fatalf("Converted XLSX file does not exist: %s", xlsxFile)
	}

	info, err := os.Stat(xlsxFile)
	if err != nil {
		t.Fatalf("Failed to stat XLSX file: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("Converted XLSX file is empty")
	}

	t.Logf("Successfully downloaded and converted to XLSX: %s (size: %d bytes)", xlsxFile, info.Size())
}

func TestDownloadLatestBondXLS_ContainsLatestBondName(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI as gov website is flaky, we have proper retry logic in source.go")
	}
	tmpDir := t.TempDir()
	xlsFile := filepath.Join(tmpDir, "bonds.xls")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := downloadLatestBondXLS(ctx, xlsFile)
	if err != nil {
		t.Fatalf("DownloadLatestBondXLS() error = %v", err)
	}

	xlsxFile, err := xlsconv.ToXLSX(xlsFile)
	if err != nil {
		t.Fatalf("ToXLSX() error = %v", err)
	}

	repo, err := LoadFromXLSX(slog.New(slog.NewTextHandler(os.Stdout, nil)), xlsxFile)
	if err != nil {
		t.Fatalf("LoadFromXLSX() error = %v", err)
	}

	warsawTime := time.Now().In(tz.UnifiedTimezone)
	warsawMonth := warsawTime.Month()

	expectedEDOName := fmt.Sprintf("EDO%02d%02d", warsawMonth, (warsawTime.Year()+10)%100)

	bond, err := repo.Lookup(expectedEDOName)
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}

	if bond.Name != expectedEDOName {
		t.Fatalf("Lookup() expected %s, got %s", expectedEDOName, bond.Name)
	}

	if bond.MonthsToMaturity != 120 {
		t.Fatalf("Lookup() expected buyout in 10 years, got %d months", bond.MonthsToMaturity)
	}
}
