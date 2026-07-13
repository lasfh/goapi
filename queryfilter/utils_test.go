package queryfilter

import (
	"net/url"
	"testing"
)

func TestFirstNonEmptyValue(t *testing.T) {
	tests := []struct {
		name   string
		query  url.Values
		fields []string
		want   string
	}{
		{
			name:   "query vazia",
			query:  url.Values{},
			fields: []string{"search", "q"},
			want:   "",
		},
		{
			name:   "primeiro campo preenchido",
			query:  url.Values{"search": {"jose"}, "q": {"maria"}},
			fields: []string{"search", "q"},
			want:   "jose",
		},
		{
			name:   "usa o próximo campo quando o primeiro está vazio",
			query:  url.Values{"search": {""}, "q": {"maria"}},
			fields: []string{"search", "q"},
			want:   "maria",
		},
		{
			name:   "valor somente com espaços é considerado vazio",
			query:  url.Values{"search": {"   "}},
			fields: []string{"search"},
			want:   "",
		},
		{
			name:   "valor com espaços nas bordas é aparado",
			query:  url.Values{"search": {"  jose  "}},
			fields: []string{"search"},
			want:   "jose",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstNonEmptyValue(tt.query, tt.fields...); got != tt.want {
				t.Errorf("firstNonEmptyValue() = %q, esperado %q", got, tt.want)
			}
		})
	}
}
