package api

import (
	"fmt"
)

func Util_Format_Get2DigitNumber(num int64) string {
	if num < 10 {
		return fmt.Sprintf("0%d", num)
	}
	return fmt.Sprintf("%d", num)
}
