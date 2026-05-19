package decimal

import "testing"

func TestAmountStringPreservesIntegerTrailingZeroes(t *testing.T) {
	for _, value := range []string{"10", "30", "100", "200"} {
		amount, err := Parse(value)
		if err != nil {
			t.Fatalf("Parse(%q) error = %v", value, err)
		}
		if amount.String() != value {
			t.Fatalf("Parse(%q).String() = %q, want %q", value, amount.String(), value)
		}
	}
}

func TestAmountStringTrimsFractionalTrailingZeroes(t *testing.T) {
	amount, err := Parse("10.5000")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if amount.String() != "10.5" {
		t.Fatalf("String() = %q, want 10.5", amount.String())
	}
}

func TestAmountAverage(t *testing.T) {
	left, err := Parse("20.5")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	right, err := Parse("21.5")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	average := left.Add(right).Average(2)
	if average != "21" {
		t.Fatalf("Average() = %q, want 21", average)
	}
}

func TestParseRejectsMalformedDecimals(t *testing.T) {
	for _, value := range []string{".", "-", "1.", ".1", "1.2.3", "1%%"} {
		if _, err := Parse(value); err == nil {
			t.Fatalf("Parse(%q) error = nil, want validation", value)
		}
	}
}
