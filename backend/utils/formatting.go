package utils

import "strconv"

type intLike interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func FormatIntAsTwoDigitString[T intLike](num T) string {
	if num < 10 {
		return "0" + strconv.FormatInt(int64(num), 10)
	}
	return strconv.FormatInt(int64(num), 10)
}

func ParseFileSize(fileSize string) int64 {
	size, err := strconv.Atoi(fileSize)
	if err != nil {
		return 0
	}
	return int64(size)
}
