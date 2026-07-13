package queryfilter

import (
	"errors"
	"net/url"
	"reflect"
	"strconv"
	"testing"
)

func TestMakeSearch(t *testing.T) {
	allowedContexts := []AllowedContext{
		{
			Context: "cliente",
			Fields: []FieldFromContext{
				{Path: "c.name"},
				{Path: "c.id", Type: Int64},
			},
		},
		{
			Context: "produto",
			Fields: []FieldFromContext{
				{Path: "p.description"},
			},
		},
	}

	tests := []struct {
		name            string
		query           url.Values
		allowedContexts []AllowedContext
		want            FieldsFromSearch
		wantErr         error
	}{
		{
			name:            "sem termo de busca",
			query:           url.Values{},
			allowedContexts: allowedContexts,
			want:            nil,
		},
		{
			name:            "termo em branco é ignorado",
			query:           url.Values{"search": {"   "}},
			allowedContexts: allowedContexts,
			want:            nil,
		},
		{
			name:            "busca textual em todos os contextos",
			query:           url.Values{"search": {"jose"}},
			allowedContexts: allowedContexts,
			want: FieldsFromSearch{
				{Field: "c.name", Value: "jose", Type: Text, Conditional: Contains},
				{Field: "p.description", Value: "jose", Type: Text, Conditional: Contains},
			},
		},
		{
			name:            "busca numérica inclui campos inteiros com equal",
			query:           url.Values{"search": {"42"}},
			allowedContexts: allowedContexts,
			want: FieldsFromSearch{
				{Field: "c.name", Value: "42", Type: Text, Conditional: Contains},
				{Field: "c.id", Value: int64(42), Type: Int64, Conditional: Equal},
				{Field: "p.description", Value: "42", Type: Text, Conditional: Contains},
			},
		},
		{
			name: "contexto restringe os campos",
			query: url.Values{
				"search":  {"jose"},
				"context": {"produto"},
			},
			allowedContexts: allowedContexts,
			want: FieldsFromSearch{
				{Field: "p.description", Value: "jose", Type: Text, Conditional: Contains},
			},
		},
		{
			name: "contexto tudo equivale a todos",
			query: url.Values{
				"search":  {"jose"},
				"context": {"tudo"},
			},
			allowedContexts: allowedContexts,
			want: FieldsFromSearch{
				{Field: "c.name", Value: "jose", Type: Text, Conditional: Contains},
				{Field: "p.description", Value: "jose", Type: Text, Conditional: Contains},
			},
		},
		{
			name: "aliases q e search-context",
			query: url.Values{
				"q":              {"jose"},
				"search-context": {"cliente"},
			},
			allowedContexts: allowedContexts,
			want: FieldsFromSearch{
				{Field: "c.name", Value: "jose", Type: Text, Conditional: Contains},
			},
		},
		{
			name:  "contexto com IgnoreIf é ignorado",
			query: url.Values{"search": {"jose"}},
			allowedContexts: []AllowedContext{
				{
					Context:  "cliente",
					Fields:   []FieldFromContext{{Path: "c.name"}},
					IgnoreIf: true,
				},
			},
			want: nil,
		},
		{
			name:  "busca numérica fora do range do tipo retorna erro",
			query: url.Values{"search": {"130"}},
			allowedContexts: []AllowedContext{
				{
					Context: "cliente",
					Fields:  []FieldFromContext{{Path: "c.tipo", Type: Int8}},
				},
			},
			wantErr: strconv.ErrRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeSearch(tt.query, tt.allowedContexts)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("MakeSearch() erro = %v, esperado %v", err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("MakeSearch() erro inesperado = %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeSearch() = %#v, esperado %#v", got, tt.want)
			}
		})
	}
}
