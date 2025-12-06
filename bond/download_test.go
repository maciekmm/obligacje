package bond

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maciekmm/obligacje/internal/xlsconv"
)

func TestDownloadLatestBondXLS_FindsFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "bonds.xls")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := DownloadLatestBondXLS(ctx, outputFile)
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
	tmpDir := t.TempDir()
	xlsFile := filepath.Join(tmpDir, "bonds.xls")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := DownloadLatestBondXLS(ctx, xlsFile)
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

func TestDownloadLatestBondXLS_ContainsLatestBondSeries(t *testing.T) {
	tmpDir := t.TempDir()
	xlsFile := filepath.Join(tmpDir, "bonds.xls")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := DownloadLatestBondXLS(ctx, xlsFile)
	if err != nil {
		t.Fatalf("DownloadLatestBondXLS() error = %v", err)
	}

	xlsxFile, err := xlsconv.ToXLSX(xlsFile)
	if err != nil {
		t.Fatalf("ToXLSX() error = %v", err)
	}

	repo, err := LoadFromXLSX(xlsxFile)
	if err != nil {
		t.Fatalf("LoadFromXLSX() error = %v", err)
	}

	warsawTZ, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatalf("LoadLocation() error = %v", err)
	}

	warsawTime := time.Now().In(warsawTZ)
	warsawMonth := warsawTime.Month()

	expectedEDOSeries := fmt.Sprintf("EDO%02d%02d", warsawMonth, (warsawTime.Year()+10)%100)

	bond, err := repo.Lookup(expectedEDOSeries)
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}

	if bond.Series != expectedEDOSeries {
		t.Fatalf("Lookup() expected %s, got %s", expectedEDOSeries, bond.Series)
	}

	if bond.BuyoutInMonths != 120 {
		t.Fatalf("Lookup() expected buyout in 10 years, got %d months", bond.BuyoutInMonths)
	}
}
