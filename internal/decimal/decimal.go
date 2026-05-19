package decimal

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

var decimalPattern = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?%?$`)

type Amount struct {
	value *big.Int
	scale int
}

func Parse(value string) (Amount, error) {
	cleanValue := strings.ReplaceAll(strings.TrimSpace(value), ",", "")
	if cleanValue == "" {
		return Amount{}, fmt.Errorf("value is required")
	}
	if !decimalPattern.MatchString(cleanValue) {
		return Amount{}, fmt.Errorf("invalid decimal %q", value)
	}
	cleanValue = strings.TrimSuffix(cleanValue, "%")
	sign := 1
	if strings.HasPrefix(cleanValue, "-") {
		sign = -1
		cleanValue = strings.TrimPrefix(cleanValue, "-")
	}
	parts := strings.Split(cleanValue, ".")
	digits := parts[0]
	scale := 0
	if len(parts) == 2 {
		scale = len(parts[1])
		digits += parts[1]
	}
	amount := new(big.Int)
	if _, ok := amount.SetString(digits, 10); !ok {
		return Amount{}, fmt.Errorf("invalid decimal %q", value)
	}
	if sign < 0 {
		amount.Neg(amount)
	}
	return Amount{value: amount, scale: scale}, nil
}

func (a Amount) Add(b Amount) Amount {
	if a.value == nil {
		return b
	}
	if b.value == nil {
		return a
	}
	scale := max(a.scale, b.scale)
	left := scaleDecimal(a.value, scale-a.scale)
	right := scaleDecimal(b.value, scale-b.scale)
	return Amount{value: left.Add(left, right), scale: scale}
}

func (a Amount) Average(count int) string {
	if count <= 0 || a.value == nil {
		return "0"
	}
	ratio := new(big.Rat).SetFrac(a.value, big.NewInt(1))
	if a.scale > 0 {
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(a.scale)), nil)
		ratio.Quo(ratio, new(big.Rat).SetInt(divisor))
	}
	ratio.Quo(ratio, new(big.Rat).SetInt64(int64(count)))
	return trimDecimalString(ratio.FloatString(6))
}

func (a Amount) String() string {
	if a.value == nil {
		return "0"
	}
	value := new(big.Int).Set(a.value)
	sign := ""
	if value.Sign() < 0 {
		sign = "-"
		value.Abs(value)
	}
	text := value.String()
	if a.scale > 0 {
		for len(text) <= a.scale {
			text = "0" + text
		}
		split := len(text) - a.scale
		text = text[:split] + "." + text[split:]
		text = trimDecimalString(text)
	}
	if text == "" || text == "0" {
		return "0"
	}
	return sign + text
}

func scaleDecimal(value *big.Int, places int) *big.Int {
	scaled := new(big.Int).Set(value)
	if places <= 0 {
		return scaled
	}
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	return scaled.Mul(scaled, multiplier)
}

func trimDecimalString(text string) string {
	if !strings.Contains(text, ".") {
		return text
	}
	text = strings.TrimRight(text, "0")
	return strings.TrimRight(text, ".")
}
