package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
)

func parseRegionalPricePatchMoney(value string, formatError func() error) (play.Money, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 && len(parts) != 3 {
		return play.Money{}, formatError()
	}
	currency, err := play.NewCurrencyCode(parts[0])
	if err != nil {
		return play.Money{}, err
	}
	units, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return play.Money{}, fmt.Errorf("price units must be an integer: %w", err)
	}
	var nanos int64
	if len(parts) == 3 {
		nanos, err = strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return play.Money{}, fmt.Errorf("price nanos must be an integer: %w", err)
		}
	}
	return play.Money{CurrencyCode: currency.String(), Units: units, Nanos: nanos}, nil
}
