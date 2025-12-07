package xlsconv

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ToXLSX converts an XLS file to XLSX format using LibreOffice.
// converted file is placed in the same directory as the input file.
func ToXLSX(xlsFile string) (string, error) {
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "xlsx", filepath.Base(xlsFile))
	cmd.Dir = filepath.Dir(xlsFile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to convert file: %w, output: %s", err, string(output))
	}

	baseName := strings.TrimSuffix(filepath.Base(xlsFile), filepath.Ext(xlsFile))
	outputFileName := baseName + ".xlsx"

	return filepath.Join(cmd.Dir, outputFileName), nil
}
