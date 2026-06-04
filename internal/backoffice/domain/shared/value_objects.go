package shared

import (
	"fmt"
	"strings"
)

const defaultCurrency = "USD"

type Money struct {
	currency string
	cents    int64
}

func NewMoney(currency string, cents int64) (Money, error) {
	currency = strings.TrimSpace(strings.ToUpper(currency))
	if currency == "" {
		currency = defaultCurrency
	}
	if cents < 0 {
		return Money{}, fmt.Errorf("money amount must not be negative")
	}
	return Money{currency: currency, cents: cents}, nil
}

func ParseMoney(raw string) (Money, error) {
	negative := strings.Contains(raw, "-")
	cleaned := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '.' {
			return r
		}
		return -1
	}, raw)
	if cleaned == "" {
		return Money{}, fmt.Errorf("invalid money")
	}
	var dollars float64
	if _, err := fmt.Sscanf(cleaned, "%f", &dollars); err != nil {
		return Money{}, fmt.Errorf("invalid money")
	}
	cents := int64(dollars*100 + 0.5)
	if negative {
		cents = -cents
	}
	return Money{currency: defaultCurrency, cents: cents}, nil
}

func (m Money) Add(other Money) (Money, error) {
	if err := m.ensureSameCurrency(other); err != nil {
		return Money{}, err
	}
	return Money{currency: m.currency, cents: m.cents + other.cents}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if err := m.ensureSameCurrency(other); err != nil {
		return Money{}, err
	}
	return Money{currency: m.currency, cents: m.cents - other.cents}, nil
}

func (m Money) MulInt(qty int) Money {
	return Money{currency: m.currency, cents: m.cents * int64(qty)}
}

func (m Money) Format() string {
	sign := ""
	cents := m.cents
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("$%s%d.%02d", sign, cents/100, cents%100)
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) Cents() int64 {
	return m.cents
}

func (m Money) ensureSameCurrency(other Money) error {
	if m.currency != other.currency {
		return fmt.Errorf("money currency mismatch: %s != %s", m.currency, other.currency)
	}
	return nil
}
