package bondxls

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/maciekmm/obligacje/bond"
	"github.com/maciekmm/obligacje/tz"
	"github.com/xuri/excelize/v2"
)

var (
	supportedNames = []string{
		"TOS", "DOS",
		"ROR", "DOR",
		"COI", "EDO", "ROS", "ROD",
	}
)

const (
	dateFormat = "_2/01/2006"
)

func namePrefix(name string) string {
	if len(name) < 3 {
		return ""
	}
	return name[:3]
}

func interestRecalculation(name string) (bond.CouponPaymentsFrequency, error) {
	prefix := namePrefix(name)
	if prefix == "" {
		return bond.CouponPaymentsFrequencyUnknown, fmt.Errorf("invalid name prefix: %s", name)
	}
	switch prefix {
	case "TOS", "DOS":
		return bond.CouponPaymentsFrequencyYearly, nil
	case "ROR", "DOR":
		return bond.CouponPaymentsFrequencyMonthly, nil
	case "COI", "EDO", "ROS", "ROD":
		return bond.CouponPaymentsFrequencyYearly, nil
	// TODO: setting quarterly payment frequency for OTS doesn't work
	// as interest calculation for it is based on number of days (or fixed value) and not frequency of payments
	// case "OTS":
	// 	return bond.CouponPaymentsFrequencyQuarterly, nil
	default:
		return bond.CouponPaymentsFrequencyUnknown, fmt.Errorf("invalid name prefix: %s", prefix)
	}
}

type XLSXRepository struct {
	logger *slog.Logger
	bonds  map[string]bond.Bond
}

func (r *XLSXRepository) Lookup(name string) (bond.Bond, error) {
	bnd, ok := r.bonds[name]
	if !ok {
		return bond.Bond{}, bond.ErrNameNotFound
	}
	return bnd, nil
}

func LoadFromXLSX(logger *slog.Logger, file string) (*XLSXRepository, error) {
	repo := &XLSXRepository{
		logger: logger,
		bonds:  make(map[string]bond.Bond),
	}

	xls, err := excelize.OpenFile(file)
	if err != nil {
		return nil, fmt.Errorf("error loading excel: %w", err)
	}
	defer xls.Close()

	for _, namePrefix := range supportedNames {
		if bonds, err := parseSheet(logger, xls, namePrefix); err != nil {
			return nil, fmt.Errorf("error loading sheet %s: %w", namePrefix, err)
		} else {
			for name, bond := range bonds {
				repo.bonds[name] = bond
			}
			logger.Info("loaded bonds", "bonds_no", len(bonds), "name", namePrefix)
		}
	}
	return repo, nil
}

func parseSheet(logger *slog.Logger, xls *excelize.File, namePrefix string) (map[string]bond.Bond, error) {
	bonds := make(map[string]bond.Bond)

	rows, err := xls.GetRows(namePrefix)
	if err != nil {
		return nil, fmt.Errorf("error getting rows: %w", err)
	}

	var headers []string
	for i, row := range rows {
		if len(row) < 2 {
			logger.Debug("skipping short row", "sheet", namePrefix, "row", i+1)
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
		if !strings.HasPrefix(row[0], namePrefix) {
			if i == 1 {
				logger.Debug("found second header row, appending values", "sheet", namePrefix, "row", i+1, "namePrefix", row[0])
				// second header row
				for j, cell := range row {
					if cell != "" {
						headers[j] += " " + strings.TrimSpace(cell)
					}
				}
			} else {
				logger.Debug("skipping row", "sheet", namePrefix, "row", i+1, "name", row[0])
			}
			continue
		}

		bond, err := rowToBond(headers, row)
		if err != nil {
			logger.Warn("skipping row", "sheet", namePrefix, "row", i+1, "name", row[0], "error", err)
			continue
		}

		bonds[bond.Name] = bond
	}

	return bonds, nil
}

func rowToBond(headers, row []string) (bond.Bond, error) {
	bond := bond.Bond{}
	for j, cell := range row {
		if j >= len(headers) {
			return bond, fmt.Errorf("extra cell in row")
		}
		if cell == "" {
			continue
		}
		header := headers[j]
		switch {
		case header == "Seria":
			bond.Name = cell
		case header == "Kod ISIN":
			bond.ISIN = cell
		case header == "Data wykupu":
			parts := strings.Split(cell, " ")
			if len(parts) < 2 {
				return bond, fmt.Errorf("invalid buyout period format")
			}
			periodValue, err := strconv.Atoi(parts[0])
			if err != nil {
				return bond, fmt.Errorf("error parsing buyout period value: %w", err)
			}
			switch parts[1] {
			case "rok", "lat/a":
				bond.MonthsToMaturity = periodValue * 12
			case "miesięcy", "miesiąc", "miesiące":
				bond.MonthsToMaturity = periodValue
			default:
				return bond, fmt.Errorf("invalid buyout period format")
			}
		case header == "Cena emisyjna":
			if price, err := parsePrice(cell); err == nil {
				bond.FaceValue = price
			} else {
				return bond, fmt.Errorf("error parsing price: %w", err)
			}
		case header == "Cena zamiany":
			if price, err := parsePrice(cell); err == nil {
				bond.ExchangePrice = price
			} else {
				return bond, fmt.Errorf("error parsing exchange price: %w", err)
			}
		case strings.HasPrefix(header, "Oprocentowanie"):
			if percentage, err := parsePercentage(cell); err == nil {
				bond.InterestPeriods = append(bond.InterestPeriods, percentage)
			} else {
				return bond, fmt.Errorf("error parsing interest percentage: %w", err)
			}
		case strings.HasPrefix(header, "Marża"):
			if percentage, err := parsePercentage(cell); err == nil {
				bond.Margin = percentage
			} else {
				return bond, fmt.Errorf("error parsing margin percentage: %w", err)
			}
		case header == "Początek sprzedaży":
			if saleStart, err := time.ParseInLocation(dateFormat, cell, tz.UnifiedTimezone); err == nil {
				bond.SaleStart = saleStart
			}
		case header == "Koniec sprzedaży":
			if saleEnd, err := time.ParseInLocation(dateFormat, cell, tz.UnifiedTimezone); err == nil {
				bond.SaleEnd = saleEnd
			}
		}
	}
	recalc, err := interestRecalculation(bond.Name)
	if err != nil {
		return bond, fmt.Errorf("error parsing interest recalculation: %w", err)
	}
	bond.CouponPaymentsFrequency = recalc

	// sometimes sale start and sale end are not provided
	if bond.SaleStart.IsZero() {
		bond.SaleStart = nameToSaleStart(bond.Name, bond.MonthsToMaturity)
		if !bond.SaleStart.IsZero() && bond.SaleEnd.IsZero() {
			bond.SaleEnd = bond.SaleStart.AddDate(0, 1, -1)
		}
	}

	if bond.SaleEnd.IsZero() {
		bond.SaleEnd = bond.SaleStart.AddDate(0, 1, -1)
	}

	// If it's fixed interest bond, fill interest periods for each year
	namePrefix := namePrefix(bond.Name)
	if namePrefix == "TOS" || namePrefix == "DOS" {
		for i := 1; i < bond.MonthsToMaturity/12; i++ {
			bond.InterestPeriods = append(bond.InterestPeriods, bond.InterestPeriods[0])
		}
	}

	return bond, nil
}

func nameToSaleStart(name string, monthsToMaturity int) time.Time {
	if len(name) < 4 {
		return time.Time{}
	}
	suffix := name[len(name)-4:]
	month, err := strconv.Atoi(suffix[:2])
	if err != nil {
		return time.Time{}
	}
	year, err := strconv.Atoi(suffix[2:])
	if err != nil {
		return time.Time{}
	}

	maturity := time.Date(2000+year, time.Month(month), 1, 0, 0, 0, 0, tz.UnifiedTimezone)
	return maturity.AddDate(0, -monthsToMaturity, 0)
}

func parsePrice(cell string) (bond.Price, error) {
	if cell == "-" {
		return 0, nil
	}
	price, err := strconv.ParseFloat(cell, 64)
	return bond.Price(price), err
}

func parsePercentage(cell string) (bond.Percentage, error) {
	noPercentageSign := strings.TrimSuffix(cell, "%")
	price, err := strconv.ParseFloat(noPercentageSign, 64)
	return bond.Percentage(price / 100.0), err
}
