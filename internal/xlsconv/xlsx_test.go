package xlsconv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/maciekmm/obligacje/internal/testutil"
)

func TestToXLSX(t *testing.T) {
	testDataDir := testutil.TestDataDirectory()
	inputFile := filepath.Join(testDataDir, "data.xls")

	outputFile, err := ToXLSX(inputFile)
	if err != nil {
		t.Fatalf("ToXLSX() error = %v", err)
	}

	if outputFile != filepath.Join(testDataDir, "data.xlsx") {
		t.Fatalf("ToXLSX() returned unexpected output file path: %s", outputFile)
	}

	t.Cleanup(func() {
		if err := os.Remove(outputFile); err != nil {
			t.Logf("Failed to clean up output file: %v", err)
		}
	})

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Output file does not exist: %s", outputFile)
	}

	t.Logf("Successfully converted %s to %s", inputFile, outputFile)
}
