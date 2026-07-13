package queryfilter

import (
	"reflect"
	"testing"
)

func TestFieldsFromSearchToSQLQuery(t *testing.T) {
	tests := []struct {
		name      string
		fields    FieldsFromSearch
		wantQuery string
		wantArgs  []any
	}{
		{
			name:      "sem campos",
			fields:    FieldsFromSearch{},
			wantQuery: "",
			wantArgs:  nil,
		},
		{
			name: "um campo",
			fields: FieldsFromSearch{
				{Field: "name", Value: "jose", Type: Text, Conditional: Contains},
			},
			wantQuery: "(LOWER(name) LIKE LOWER(?))",
			wantArgs:  []any{"%jose%"},
		},
		{
			name: "vários campos unidos com OR",
			fields: FieldsFromSearch{
				{Field: "name", Value: "42", Type: Text, Conditional: Contains},
				{Field: "id", Value: int64(42), Type: Int64, Conditional: Equal},
			},
			wantQuery: "(LOWER(name) LIKE LOWER(?) OR id = ?)",
			wantArgs:  []any{"%42%", int64(42)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs := tt.fields.ToSQLQuery()

			if gotQuery != tt.wantQuery {
				t.Errorf("ToSQLQuery() query = %q, esperado %q", gotQuery, tt.wantQuery)
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("ToSQLQuery() args = %#v, esperado %#v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestFieldsAndSearchToSQLQuery(t *testing.T) {
	tests := []struct {
		name      string
		fields    FieldsAndSearch
		wantQuery string
		wantArgs  []any
	}{
		{
			name:      "sem campos e sem busca",
			fields:    FieldsAndSearch{},
			wantQuery: "",
			wantArgs:  []any{},
		},
		{
			name: "somente filtros",
			fields: FieldsAndSearch{
				Fields: Fields{
					{Field: "age", Value: int64(30), Type: Int64, Conditional: Equal},
				},
			},
			wantQuery: "age = ?",
			wantArgs:  []any{int64(30)},
		},
		{
			name: "somente busca",
			fields: FieldsAndSearch{
				FieldsFromContext: FieldsFromSearch{
					{Field: "name", Value: "jose", Type: Text, Conditional: Contains},
				},
			},
			wantQuery: "(LOWER(name) LIKE LOWER(?))",
			wantArgs:  []any{"%jose%"},
		},
		{
			name: "filtros e busca unidos com AND",
			fields: FieldsAndSearch{
				Fields: Fields{
					{Field: "age", Value: int64(30), Type: Int64, Conditional: Equal},
				},
				FieldsFromContext: FieldsFromSearch{
					{Field: "name", Value: "jose", Type: Text, Conditional: Contains},
					{Field: "description", Value: "jose", Type: Text, Conditional: Contains},
				},
			},
			wantQuery: "age = ? AND (LOWER(name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?))",
			wantArgs:  []any{int64(30), "%jose%", "%jose%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs := tt.fields.ToSQLQuery()

			if gotQuery != tt.wantQuery {
				t.Errorf("ToSQLQuery() query = %q, esperado %q", gotQuery, tt.wantQuery)
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("ToSQLQuery() args = %#v, esperado %#v", gotArgs, tt.wantArgs)
			}
		})
	}
}
