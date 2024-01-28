package util9s

import (
	"fmt"
	"strconv"
)

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
