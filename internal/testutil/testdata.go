package testutil

import (
	"path/filepath"
	"runtime"
)

func TestDataDirectory() string {
	_, filename, _, _ := runtime.Caller(1)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "testdata")
}
