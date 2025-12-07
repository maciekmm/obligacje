package downloader

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResilientFileStore_DownloadWithFallback(t *testing.T) {
	type fields struct {
		setupDir  func(t *testing.T, dir string)
		validator ValidatorFunc
		download  DownloadFunc
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "successful download and validation",
			fields: fields{
				setupDir: func(t *testing.T, dir string) {},
				download: func(ctx context.Context, outputFile string) error {
					return os.WriteFile(outputFile, []byte("new content"), 0644)
				},
				validator: func(ctx context.Context, file string) error {
					return nil
				},
			},
			want:    "active",
			wantErr: false,
		},
		{
			name: "download fails, no fallback",
			fields: fields{
				setupDir: func(t *testing.T, dir string) {},
				download: func(ctx context.Context, outputFile string) error {
					return errors.New("download failed")
				},
				validator: func(ctx context.Context, file string) error {
					return nil
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "download fails, fallback exists",
			fields: fields{
				setupDir: func(t *testing.T, dir string) {
					err := os.WriteFile(filepath.Join(dir, "active"), []byte("old content"), 0644)
					if err != nil {
						t.Fatal(err)
					}
				},
				download: func(ctx context.Context, outputFile string) error {
					return errors.New("download failed")
				},
				validator: func(ctx context.Context, file string) error {
					return nil
				},
			},
			want:    "active",
			wantErr: false,
		},
		{
			name: "validation fails, no fallback",
			fields: fields{
				setupDir: func(t *testing.T, dir string) {},
				download: func(ctx context.Context, outputFile string) error {
					return os.WriteFile(outputFile, []byte("bad content"), 0644)
				},
				validator: func(ctx context.Context, file string) error {
					return errors.New("validation failed")
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "validation fails, fallback exists",
			fields: fields{
				setupDir: func(t *testing.T, dir string) {
					err := os.WriteFile(filepath.Join(dir, "active"), []byte("old content"), 0644)
					if err != nil {
						t.Fatal(err)
					}
				},
				download: func(ctx context.Context, outputFile string) error {
					return os.WriteFile(outputFile, []byte("bad content"), 0644)
				},
				validator: func(ctx context.Context, file string) error {
					return errors.New("validation failed")
				},
			},
			want:    "active",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.fields.setupDir != nil {
				tt.fields.setupDir(t, dir)
			}

			d := NewResilientFileDownloader(dir, tt.fields.validator, tt.fields.download)
			got, err := d.DownloadWithFallback(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("ResilientFileStore.DownloadWithFallback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				wantPath := filepath.Join(dir, tt.want)
				if got != wantPath {
					t.Errorf("ResilientFileStore.DownloadWithFallback() = %v, want %v", got, wantPath)
				}
				// Verify file actually exists
				if _, err := os.Stat(got); os.IsNotExist(err) {
					t.Errorf("ResilientFileStore.DownloadWithFallback() returned path %v but file does not exist", got)
				}
			}
		})
	}
}
