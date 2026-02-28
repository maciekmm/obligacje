package obligacje

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/maciekmm/obligacje/bond"
	"github.com/maciekmm/obligacje/bondxls"
	"github.com/maciekmm/obligacje/internal/downloader"
	"github.com/maciekmm/obligacje/internal/periodical"
)

func validateBondFile(ctx context.Context, file string) error {
	_, err := bondxls.LoadFromXLSX(slog.New(slog.DiscardHandler), file)
	if err != nil {
		return err
	}
	return nil
}

type BondSource struct {
	bondsLoader *periodical.Loader[bond.Repository]
}

func NewBondSource(logger *slog.Logger, dir string) (*BondSource, error) {
	files := downloader.NewResilientFileDownloader(dir, validateBondFile, bondxls.DownloadLatestAndConvert)

	const maxRetries = 3
	const initialDelay = 2 * time.Second

	loadFn := func() (bond.Repository, error) {
		var lastErr error
		for attempt := range maxRetries {
			file, err := files.DownloadWithFallback(context.Background())
			if err != nil {
				lastErr = err
			} else {
				repo, err := bondxls.LoadFromXLSX(logger, file)
				if err != nil {
					lastErr = err
				} else {
					return repo, nil
				}
			}
			if attempt < maxRetries-1 {
				delay := initialDelay * time.Duration(1<<attempt)
				logger.Error("failed to load bond data, retrying", "err", lastErr, "attempt", attempt+1, "next_retry_in", delay)
				time.Sleep(delay)
			}
		}
		return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
	}

	bondsLoader, err := periodical.NewLoader(12*time.Hour, loadFn, periodical.ErrBehaviorKeepOld)
	if err != nil {
		return nil, fmt.Errorf("initial bond data load failed: %w", err)
	}

	return &BondSource{
		bondsLoader: bondsLoader,
	}, nil
}

func (s *BondSource) Close() error {
	s.bondsLoader.Stop()
	return nil
}

func (s *BondSource) Lookup(name string) (bond.Bond, error) {
	cur, err := s.bondsLoader.Current()
	if err != nil {
		return bond.Bond{}, err
	}
	return cur.Lookup(name)
}
