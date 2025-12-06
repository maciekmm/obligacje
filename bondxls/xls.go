package bondxls

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/maciekmm/obligacje/bond"
	"github.com/xuri/excelize/v2"
)

var (
	supportedSeries = []string{
		"OTS", "TOS", "DOS",
		"ROR", "DOR",
		"COI", "EDO", "ROS", "ROD",
	}
)

func interestRecalculation(seriesPrefix string) bond.InterestRecalculation {
	switch seriesPrefix {
	case "OTS", "TOS", "DOS":
		return bond.InterestRecalculationNone
	case "ROR", "DOR":
		return bond.InterestRecalculationMonthly
	case "COI", "EDO", "ROS", "ROD":
		return bond.InterestRecalculationYearly
	default:
		return bond.InterestRecalculationUnknown
	}
}

type XLSXRepository struct {
	bonds map[string]bond.Bond
}

func (r *XLSXRepository) Lookup(series string) (bond.Bond, error) {
	bnd, ok := r.bonds[series]
	if !ok {
		return bond.Bond{}, bond.ErrSeriesNotFound
	}
	return bnd, nil
}

func LoadFromXLSX(file string) (*XLSXRepository, error) {
	repo := &XLSXRepository{
		bonds: make(map[string]bond.Bond),
	}

	xls, err := excelize.OpenFile(file)
	if err != nil {
		return nil, fmt.Errorf("error loading excel: %w", err)
	}
	defer xls.Close()

	for _, seriesPrefix := range supportedSeries {
		if bonds, err := parseSheet(xls, seriesPrefix); err != nil {
			return nil, fmt.Errorf("error loading sheet %s: %w", seriesPrefix, err)
		} else {
			for series, bond := range bonds {
				repo.bonds[series] = bond
			}
			slog.Info("loaded bonds", "bonds_no", len(bonds), "series", seriesPrefix)
		}
	}
	return repo, nil
}

func parseSheet(xls *excelize.File, seriesPrefix string) (map[string]bond.Bond, error) {
	bonds := make(map[string]bond.Bond)

	rows, err := xls.GetRows(seriesPrefix)
	if err != nil {
		return nil, fmt.Errorf("error getting rows: %w", err)
	}

	var headers []string
	for i, row := range rows {
		if len(row) < 2 {
			slog.Debug("skipping short row", "sheet", seriesPrefix, "row", i+1)
			continue
		}
		if i == 0 {
			headers = row
			// if there are merged cells, fill them
			for i, header := range headers {
				if header == "" && i > 0 {
					headers[i] = headers[i-1]
				}
			}
			continue
		}
		if !strings.HasPrefix(row[0], seriesPrefix) {
			if i == 1 {
				slog.Debug("found second header row, appending values", "sheet", seriesPrefix, "row", i+1, "seriesPrefix", row[0])
				// second header row
				for j, cell := range row {
					if cell != "" {
						headers[j] += " " + strings.TrimSpace(cell)
					}
				}
			} else {
				slog.Debug("skipping row", "sheet", seriesPrefix, "row", i+1, "series", row[0])
			}
			continue
		}
		bond := bond.Bond{}
		for j, cell := range row {
			if j >= len(headers) {
				slog.Warn("extra cell in row", "sheet", seriesPrefix, "row", i+1, "cell_index", j, "value", cell)
				break
			}
			if cell == "" {
				continue
			}
			header := headers[j]
			switch {
			case header == "Seria":
				bond.Series = cell
			case header == "Kod ISIN":
				bond.ISIN = cell
			case header == "Data wykupu":
				parts := strings.Split(cell, " ")
				if len(parts) < 2 {
					slog.Warn("invalid buyout period format", "sheet", seriesPrefix, "row", i+1, "cell_index", j, "value", cell)
					continue
				}
				periodValue, err := strconv.Atoi(parts[0])
				if err != nil {
					slog.Warn("error parsing buyout period value", "sheet", seriesPrefix, "row", i+1, "cell_index", j, "value", cell, "error", err)
					continue
				}
				if parts[1] == "rok" || parts[1] == "lat/a" {
					bond.BuyoutInMonths = periodValue * 12
				} else if parts[1] == "miesięcy" || parts[1] == "miesiąc" || parts[1] == "miesiące" {
					bond.BuyoutInMonths = periodValue
				}
			case header == "Cena emisyjna":
				if price, err := parsePrice(cell); err == nil {
					bond.Price = price
				} else {
					slog.Warn("error parsing price", "sheet", seriesPrefix, "row", i+1, "cell_index", j, "value", cell, "error", err)
				}
			case header == "Cena zamiany":
				if price, err := parsePrice(cell); err == nil {
					bond.ExchangePrice = price
				} else {
					slog.Warn("error parsing exchange price", "sheet", seriesPrefix, "row", i+1, "cell_index", j, "value", cell, "error", err)
				}
			case strings.HasPrefix(header, "Oprocentowanie"):
				if percentage, err := parsePercentage(cell); err == nil {
					bond.InterestPeriods = append(bond.InterestPeriods, percentage)
				} else {
					slog.Warn("error parsing interest percentage", "sheet", seriesPrefix, "row", i+1, "cell_index", j, "value", cell, "error", err)
				}
			case strings.HasPrefix(header, "Marża"):
				if percentage, err := parsePercentage(cell); err == nil {
					bond.MarginPercentage = percentage
				} else {
					slog.Warn("error parsing margin percentage", "sheet", seriesPrefix, "row", i+1, "cell_index", j, "value", cell, "error", err)
				}
			}
		}
		bond.InterestRecalculation = interestRecalculation(seriesPrefix)
		bonds[bond.Series] = bond
	}

	return bonds, nil
}

func strDecimalAsInt(cell string) (int, error) {
	parts := strings.Split(cell, ".")
	if len(parts) != 2 || len(parts[1]) != 2 {
		return 0, fmt.Errorf("invalid format: %s", cell)
	}
	if priceWithPences, err := strconv.Atoi(parts[0] + parts[1]); err != nil {
		return 0, fmt.Errorf("error parsing decimal %s: %w", cell, err)
	} else {
		return priceWithPences, nil
	}
}

func parsePrice(cell string) (bond.Price, error) {
	if cell == "-" {
		return 0, nil
	}
	price, err := strDecimalAsInt(cell)
	return bond.Price(price), err
}

func parsePercentage(cell string) (bond.Percentage, error) {
	noPercentageSign := strings.TrimSuffix(cell, "%")
	price, err := strDecimalAsInt(noPercentageSign)
	return bond.Percentage(price), err
}
