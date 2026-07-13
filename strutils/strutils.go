package strutils

import (
	"strconv"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func Normalize(value string) (string, error) {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	s, _, err := transform.String(t, value)
	if err != nil {
		return "", err
	}

	return s, nil
}

func IsNumeric(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func ToIntValues[T ~int8 | ~int16 | ~int32 | ~int64](source []string, bitSize int) []T {
	items := make([]T, 0, len(source))

	for _, item := range source {
		value, err := strconv.ParseInt(item, 10, bitSize)
		if err != nil {
			continue
		}

		items = append(items, T(value))
	}

	return items
}
