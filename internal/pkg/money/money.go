package money

import (
	"fmt"
	"strings"
)

type Money struct {
	Amount       int64  `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}

func New(amount int64, currency string) (Money, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if len(currency) != 3 {
		return Money{}, fmt.Errorf("invalid currency code: %s", currency)
	}
	if amount < 0 {
		return Money{}, fmt.Errorf("amount must be non-negative: %d", amount)
	}
	return Money{Amount: amount, CurrencyCode: currency}, nil
}

func Must(amount int64, currency string) Money {
	m, err := New(amount, currency)
	if err != nil {
		panic(err)
	}
	return m
}

func (m Money) Add(other Money) (Money, error) {
	if m.CurrencyCode != other.CurrencyCode {
		return Money{}, fmt.Errorf("currency mismatch: %s vs %s", m.CurrencyCode, other.CurrencyCode)
	}
	return Money{Amount: m.Amount + other.Amount, CurrencyCode: m.CurrencyCode}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if m.CurrencyCode != other.CurrencyCode {
		return Money{}, fmt.Errorf("currency mismatch: %s vs %s", m.CurrencyCode, other.CurrencyCode)
	}
	return Money{Amount: m.Amount - other.Amount, CurrencyCode: m.CurrencyCode}, nil
}

func (m Money) Mul(qty int) Money {
	return Money{Amount: m.Amount * int64(qty), CurrencyCode: m.CurrencyCode}
}

func (m Money) IsZero() bool { return m.Amount == 0 }

func (m Money) Equal(other Money) bool {
	return m.Amount == other.Amount && m.CurrencyCode == other.CurrencyCode
}

var Precision = map[string]int{
	"USD": 2, "EUR": 2, "GBP": 2, "CNY": 2, "CAD": 2, "AUD": 2, "JPY": 0, "KRW": 0, "VND": 0,
}

func (m Money) Format() string {
	prec, ok := Precision[m.CurrencyCode]
	if !ok {
		prec = 2
	}
	if prec == 0 {
		return fmt.Sprintf("%d %s", m.Amount, m.CurrencyCode)
	}
	divisor := int64(1)
	for i := 0; i < prec; i++ {
		divisor *= 10
	}
	whole := m.Amount / divisor
	frac := m.Amount % divisor
	return fmt.Sprintf("%d.%0*d %s", whole, prec, frac, m.CurrencyCode)
}
