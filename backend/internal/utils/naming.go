package utils

import "fmt"

func Get2DigitNumber(num int64) string {
	if num < 10 {
		return fmt.Sprintf("0%d", num)
	}
	return fmt.Sprintf("%d", num)
}
