package client

import (
	"fmt"
	"testing"
)

type formatExpect struct {
	millis int64
	value  string
}

var currencyFormatUsd = &CurrencyFormat{
	IsoCode:          "USD",
	ExampleFormat:    "$35.42",
	SymbolFirst:      true,
	CurrencySymbol:   "$",
	DecimalDigits:    2,
	DisplaySymbol:    true,
	DecimalSeparator: ".",
	GroupSeparator:   ",",
}

func TestCurrencyFormat_Format_USD(t *testing.T) {
	pairs := []formatExpect{
		{100, "$0.10"},
		{85200, "$85.20"},
		{10250210, "$10,250.21"},
		{100250000, "$100,250.00"},
		{-15000, "-$15.00"},
	}

	for _, example := range pairs {
		formatted := currencyFormatUsd.Format(example.millis)
		if formatted != example.value {
			t.Error(fmt.Sprintf("Expected: %v but got: %v", example.value, formatted))
		}
	}
}
