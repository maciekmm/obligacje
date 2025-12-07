package obligacje

import (
	"context"
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

	bondsLoader := periodical.NewLoader[bond.Repository](12*time.Hour, func() (bond.Repository, error) {
		file, err := files.DownloadWithFallback(context.Background())
		if err != nil {
			return nil, err
		}
		return bondxls.LoadFromXLSX(logger, file)
	}, periodical.ErrBehaviorKeepOld)

	return &BondSource{
		bondsLoader: bondsLoader,
	}, nil
}

func (s *BondSource) Close() error {
	s.bondsLoader.Stop()
	return nil
}

func (s *BondSource) Lookup(series string) (bond.Bond, error) {
	cur, err := s.bondsLoader.Current()
	if err != nil {
		return bond.Bond{}, err
	}
	return cur.Lookup(series)
}
