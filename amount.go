package util9s

import (
	"fmt"
	"math"
	"strconv"
)

// Money is an amount in fen (minor units). Prefer this over float yuan compares.
type Money int64

// Fen returns the fen value.
func (m Money) Fen() int64 { return int64(m) }

// YuanString formats fen as a 2-decimal yuan string.
func (m Money) YuanString() string { return Fen2Yuan(int64(m)) }

// Fen2Yuan 分转元 （整形转字符串）
func Fen2Yuan(amount int64) string {
	return fmt.Sprintf("%.2f", float64(amount)/100)
}

// FenS2Yuan 分转元(字符串转字符串）
func FenS2Yuan(amount string) string {
	parseInt, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return amount
	}
	return fmt.Sprintf("%.2f", float64(parseInt)/100)
}

// FenSF2Yuan 分转元(字符串转字符串）
func FenSF2Yuan(amount string) string {
	f, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return amount
	}
	return fmt.Sprintf("%.2f", f/100)
}

// YuanToFen converts yuan (float) to fen using math.Round (half away from ±∞ via Round).
// Prefer storing/comparing fen end-to-end; use this only at the float boundary.
func YuanToFen(yuan float64) int64 {
	return int64(math.Round(yuan * 100))
}

// MustMatchPayAmount reports whether paidFen equals dueFen within tolFen (tolFen < 0 → 0).
// Both sides are int64 fen — never compare float yuan directly.
func MustMatchPayAmount(paidFen, dueFen, tolFen int64) bool {
	if tolFen < 0 {
		tolFen = 0
	}
	diff := paidFen - dueFen
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolFen
}

// MatchPayAmountDefault is MustMatchPayAmount with 1 fen tolerance (common for round-trip).
func MatchPayAmountDefault(paidFen, dueFen int64) bool {
	return MustMatchPayAmount(paidFen, dueFen, 1)
}
