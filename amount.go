package util9s

import "fmt"

// Fen2Yuan 分转元
func Fen2Yuan(amount int64) string {
	return fmt.Sprintf("%.2f", float64(amount)/100)
}
