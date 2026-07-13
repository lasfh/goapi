package queryfilter

import (
	"reflect"
	"testing"
)

func TestFieldToSQL(t *testing.T) {
	tests := []struct {
		name      string
		field     Field
		wantQuery string
		wantArgs  []any
	}{
		{
			name:      "equal simples",
			field:     Field{Field: "name", Value: "jose", Type: Text, Conditional: Equal},
			wantQuery: "name = ?",
			wantArgs:  []any{"jose"},
		},
		{
			name:      "equal com múltiplos valores",
			field:     Field{Field: "age", Value: []int64{1, 2}, Type: Int64, Conditional: Equal, MultipleValues: true},
			wantQuery: "age IN (?)",
			wantArgs:  []any{[]int64{1, 2}},
		},
		{
			name:      "contains em texto",
			field:     Field{Field: "name", Value: "jose", Type: Text, Conditional: Contains},
			wantQuery: "LOWER(name) LIKE LOWER(?)",
			wantArgs:  []any{"%jose%"},
		},
		{
			name:      "contains em texto com múltiplos valores",
			field:     Field{Field: "name", Value: []string{"jose", "maria"}, Type: Text, Conditional: Contains, MultipleValues: true},
			wantQuery: "(LOWER(name) LIKE LOWER(?) OR LOWER(name) LIKE LOWER(?))",
			wantArgs:  []any{"%jose%", "%maria%"},
		},
		{
			name:      "contains em tipo não textual vira equal",
			field:     Field{Field: "age", Value: int64(30), Type: Int64, Conditional: Contains},
			wantQuery: "age = ?",
			wantArgs:  []any{int64(30)},
		},
		{
			name:      "not contain em texto",
			field:     Field{Field: "name", Value: "jose", Type: Text, Conditional: NotContain},
			wantQuery: "LOWER(name) NOT LIKE LOWER(?)",
			wantArgs:  []any{"%jose%"},
		},
		{
			name:      "not contain em texto com múltiplos valores",
			field:     Field{Field: "name", Value: []string{"jose", "maria"}, Type: Text, Conditional: NotContain, MultipleValues: true},
			wantQuery: "(LOWER(name) NOT LIKE LOWER(?) AND LOWER(name) NOT LIKE LOWER(?))",
			wantArgs:  []any{"%jose%", "%maria%"},
		},
		{
			name:      "not contain em tipo não textual",
			field:     Field{Field: "age", Value: int64(30), Type: Int64, Conditional: NotContain},
			wantQuery: "age != ?",
			wantArgs:  []any{int64(30)},
		},
		{
			name:      "not contain em tipo não textual com múltiplos valores",
			field:     Field{Field: "age", Value: []int64{1, 2}, Type: Int64, Conditional: NotContain, MultipleValues: true},
			wantQuery: "age NOT IN (?)",
			wantArgs:  []any{[]int64{1, 2}},
		},
		{
			name:      "is empty em texto",
			field:     Field{Field: "name", Type: Text, Conditional: IsEmpty},
			wantQuery: "(name IS NULL OR name = '')",
			wantArgs:  nil,
		},
		{
			name:      "is empty em inteiro com ZeroIsEmpty",
			field:     Field{Field: "age", Type: Int32, Conditional: IsEmpty, ZeroIsEmpty: true},
			wantQuery: "(age IS NULL OR age = 0)",
			wantArgs:  nil,
		},
		{
			name:      "is empty em inteiro sem ZeroIsEmpty",
			field:     Field{Field: "age", Type: Int32, Conditional: IsEmpty},
			wantQuery: "age IS NULL",
			wantArgs:  nil,
		},
		{
			name:      "is empty em data",
			field:     Field{Field: "created_at", Type: DateTime, Conditional: IsEmpty},
			wantQuery: "created_at IS NULL",
			wantArgs:  nil,
		},
		{
			name:      "not empty em texto",
			field:     Field{Field: "name", Type: Text, Conditional: NotEmpty},
			wantQuery: "(name IS NOT NULL AND name != '')",
			wantArgs:  nil,
		},
		{
			name:      "not empty em inteiro com ZeroIsEmpty",
			field:     Field{Field: "age", Type: Int32, Conditional: NotEmpty, ZeroIsEmpty: true},
			wantQuery: "(age IS NOT NULL AND age != 0)",
			wantArgs:  nil,
		},
		{
			name:      "not empty em data",
			field:     Field{Field: "created_at", Type: DateTime, Conditional: NotEmpty},
			wantQuery: "created_at IS NOT NULL",
			wantArgs:  nil,
		},
		{
			name: "custom to sql tem precedência",
			field: Field{
				Field:       "name",
				Value:       "jose",
				Type:        Text,
				Conditional: Equal,
				CustomToSQL: func(field Field) (string, []any) {
					return "custom_fn(name, ?)", []any{field.Value}
				},
			},
			wantQuery: "custom_fn(name, ?)",
			wantArgs:  []any{"jose"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs := tt.field.ToSQL()

			if gotQuery != tt.wantQuery {
				t.Errorf("ToSQL() query = %q, esperado %q", gotQuery, tt.wantQuery)
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("ToSQL() args = %#v, esperado %#v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestFieldsToSQLQuery(t *testing.T) {
	tests := []struct {
		name      string
		fields    Fields
		wantQuery string
		wantArgs  []any
	}{
		{
			name:      "sem campos",
			fields:    Fields{},
			wantQuery: "",
			wantArgs:  []any{},
		},
		{
			name: "um campo",
			fields: Fields{
				{Field: "name", Value: "jose", Type: Text, Conditional: Equal},
			},
			wantQuery: "name = ?",
			wantArgs:  []any{"jose"},
		},
		{
			name: "vários campos unidos com AND",
			fields: Fields{
				{Field: "name", Value: "jose", Type: Text, Conditional: Contains},
				{Field: "age", Value: int64(30), Type: Int64, Conditional: Equal},
				{Field: "deleted_at", Type: DateTime, Conditional: IsEmpty},
			},
			wantQuery: "LOWER(name) LIKE LOWER(?) AND age = ? AND deleted_at IS NULL",
			wantArgs:  []any{"%jose%", int64(30)},
		},
		{
			name: "campo com query vazia é ignorado",
			fields: Fields{
				{
					Field: "name",
					CustomToSQL: func(field Field) (string, []any) {
						return "", nil
					},
				},
				{Field: "age", Value: int64(30), Type: Int64, Conditional: Equal},
			},
			wantQuery: "age = ?",
			wantArgs:  []any{int64(30)},
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
