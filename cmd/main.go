package main

import (
	"os"
	"time"

	"log/slog"

	"github.com/maciekmm/obligacje"
	"github.com/maciekmm/obligacje/calculator"
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

	for {
		time.Sleep(5 * time.Second)
		calc := calculator.NewCalculator(source)
		price, err := calc.Calculate("EDO0834", time.Now(), time.Now())
		if err != nil {
			slog.Error("failed to calculate", "err", err)
		}
		slog.Info("calculated", "price", price)
	}
}
