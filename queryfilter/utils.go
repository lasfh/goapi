package queryfilter

import (
	"net/url"
	"strings"
)

func firstNonEmptyValue(query url.Values, fields ...string) string {
	for index := range fields {
		value := strings.TrimSpace(
			query.Get(fields[index]),
		)
		if value != "" {
			return value
		}
	}

	return ""
}
