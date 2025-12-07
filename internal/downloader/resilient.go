package downloader

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrNoValidFile = errors.New("no valid file found")
)

const (
	activeFile = "active"
)

type ValidatorFunc func(ctx context.Context, file string) error
type DownloadFunc func(ctx context.Context, outputFile string) error

type ResilientFileStore struct {
	dir       string
	validator ValidatorFunc
	download  DownloadFunc
}

func NewResilientFileDownloader(
	dir string,
	validator ValidatorFunc,
	download DownloadFunc,
) *ResilientFileStore {
	return &ResilientFileStore{
		dir:       dir,
		validator: validator,
		download:  download,
	}
}

func (d *ResilientFileStore) latestValidFile(ctx context.Context) (string, error) {
	if _, err := os.Stat(filepath.Join(d.dir, activeFile)); err != nil {
		return "", ErrNoValidFile
	}
	return filepath.Join(d.dir, activeFile), nil
}

func randFileSuffix() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return time.Now().Format("20060102150405")
	}
	return hex.EncodeToString(bytes)
}

func (d *ResilientFileStore) tryDownload(ctx context.Context) error {
	suffix := randFileSuffix()
	candidate := filepath.Join(d.dir, "candidate-"+suffix)
	if err := d.download(ctx, candidate); err != nil {
		return fmt.Errorf("error downloading file: %w", err)
	}

	if err := d.validator(ctx, candidate); err != nil {
		_ = os.Remove(candidate)
		return fmt.Errorf("error validating file: %w", err)
	}

	slog.Info("new valid file downloaded, replacing existing one")
	if err := os.Rename(candidate, filepath.Join(d.dir, activeFile)); err != nil {
		return fmt.Errorf("error moving file: %w", err)
	}
	return nil
}

func (d *ResilientFileStore) DownloadWithFallback(ctx context.Context) (string, error) {
	if err := d.tryDownload(ctx); err != nil {
		slog.Warn("error downloading file", "error", err)
	}

	latestValid, err := d.latestValidFile(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting latest valid file: %w", err)
	}

	return latestValid, nil
}
