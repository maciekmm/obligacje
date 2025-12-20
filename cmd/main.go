package main

import (
	"net/http"
	"os"

	"log/slog"

	"github.com/maciekmm/obligacje"
	"github.com/maciekmm/obligacje/internal/server"
)

func main() {
	dir := "./data"
	if err := os.Mkdir(dir, 0755); err != nil && !os.IsExist(err) {
		panic(err)
	}

	source, err := obligacje.NewBondSource(slog.Default(), dir)
	if err != nil {
		panic(err)
	}
	defer source.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	srv := server.NewServer(source, logger)

	slog.Info("starting server on :8080")
	if err := http.ListenAndServe(":8080", srv); err != nil {
		panic(err)
	}
}
