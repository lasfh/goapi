package queryfilter

import (
	"errors"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/lasfh/goapi/types/daterange"
)

func TestConditionalIsNegation(t *testing.T) {
	tests := []struct {
		name string
		cond Conditional
		want bool
	}{
		{name: "equal não é negação", cond: Equal, want: false},
		{name: "contains não é negação", cond: Contains, want: false},
		{name: "not contain é negação", cond: NotContain, want: true},
		{name: "is empty não é negação", cond: IsEmpty, want: false},
		{name: "not empty é negação", cond: NotEmpty, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cond.IsNegation(); got != tt.want {
				t.Errorf("IsNegation() = %v, esperado %v", got, tt.want)
			}
		})
	}
}

func TestConditionalFromString(t *testing.T) {
	tests := []struct {
		name    string
		cond    conditional
		want    Conditional
		wantErr bool
	}{
		{name: "vazio equivale a equal", cond: undefined, want: Equal},
		{name: "equal", cond: equal, want: Equal},
		{name: "in", cond: contains, want: Contains},
		{name: "not-in", cond: notContain, want: NotContain},
		{name: "is-null", cond: isEmpty, want: IsEmpty},
		{name: "not-null", cond: notEmpty, want: NotEmpty},
		{name: "inválido", cond: conditional("xpto"), want: Equal, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cond.Conditional()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Conditional() erro = %v, esperado erro = %v", err, tt.wantErr)
			}

			if tt.wantErr && !errors.Is(err, ErrInvalidConditional) {
				t.Errorf("Conditional() erro = %v, esperado ErrInvalidConditional", err)
			}

			if got != tt.want {
				t.Errorf("Conditional() = %v, esperado %v", got, tt.want)
			}
		})
	}
}

func TestConditionalNotUseValue(t *testing.T) {
	tests := []struct {
		name       string
		cond       conditional
		want       Conditional
		wantNotUse bool
	}{
		{name: "is-null não usa valor", cond: isEmpty, want: IsEmpty, wantNotUse: true},
		{name: "not-null não usa valor", cond: notEmpty, want: NotEmpty, wantNotUse: true},
		{name: "equal usa valor", cond: equal, want: 0, wantNotUse: false},
		{name: "in usa valor", cond: contains, want: 0, wantNotUse: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, notUse := tt.cond.NotUseValue()
			if got != tt.want || notUse != tt.wantNotUse {
				t.Errorf("NotUseValue() = (%v, %v), esperado (%v, %v)", got, notUse, tt.want, tt.wantNotUse)
			}
		})
	}
}

func TestFieldType(t *testing.T) {
	tests := []struct {
		name        string
		fieldType   FieldType
		wantInteger bool
		wantDate    bool
	}{
		{name: "text", fieldType: Text},
		{name: "bool", fieldType: Bool},
		{name: "date range", fieldType: DateRange, wantDate: true},
		{name: "date time", fieldType: DateTime, wantDate: true},
		{name: "int8", fieldType: Int8, wantInteger: true},
		{name: "int16", fieldType: Int16, wantInteger: true},
		{name: "int32", fieldType: Int32, wantInteger: true},
		{name: "int64", fieldType: Int64, wantInteger: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fieldType.IsInteger(); got != tt.wantInteger {
				t.Errorf("IsInteger() = %v, esperado %v", got, tt.wantInteger)
			}

			if got := tt.fieldType.IsNumber(); got != tt.wantInteger {
				t.Errorf("IsNumber() = %v, esperado %v", got, tt.wantInteger)
			}

			if got := tt.fieldType.IsDate(); got != tt.wantDate {
				t.Errorf("IsDate() = %v, esperado %v", got, tt.wantDate)
			}
		})
	}
}

func TestMultipleValuesIs(t *testing.T) {
	if NoMultipleValues.Is() {
		t.Error("NoMultipleValues.Is() = true, esperado false")
	}

	if !ByArray.Is() {
		t.Error("ByArray.Is() = false, esperado true")
	}
}

func TestAllowedFieldFieldName(t *testing.T) {
	tests := []struct {
		name  string
		field AllowedField
		want  string
	}{
		{
			name:  "sem path usa field",
			field: AllowedField{Field: "name"},
			want:  "name",
		},
		{
			name:  "com path usa path",
			field: AllowedField{Field: "user", Path: "u.user_id"},
			want:  "u.user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.field.FieldName(); got != tt.want {
				t.Errorf("FieldName() = %q, esperado %q", got, tt.want)
			}
		})
	}
}

func TestMake(t *testing.T) {
	tests := []struct {
		name          string
		query         url.Values
		allowedFields []AllowedField
		want          Fields
		wantErr       error
	}{
		{
			name:          "query vazia",
			query:         url.Values{},
			allowedFields: []AllowedField{{Field: "name"}},
			want:          Fields{},
		},
		{
			name:          "campo não permitido é ignorado",
			query:         url.Values{"other": {"value"}},
			allowedFields: []AllowedField{{Field: "name"}},
			want:          Fields{},
		},
		{
			name:          "valor vazio é ignorado",
			query:         url.Values{"name": {""}},
			allowedFields: []AllowedField{{Field: "name"}},
			want:          Fields{},
		},
		{
			name:          "campo com IgnoreIf é ignorado",
			query:         url.Values{"name": {"jose"}},
			allowedFields: []AllowedField{{Field: "name", IgnoreIf: true}},
			want:          Fields{},
		},
		{
			name:          "texto simples com conditional padrão",
			query:         url.Values{"name": {"jose"}},
			allowedFields: []AllowedField{{Field: "name"}},
			want: Fields{
				{Field: "name", Value: "jose", Type: Text, Conditional: Equal},
			},
		},
		{
			name:          "path substitui o nome do campo",
			query:         url.Values{"user": {"jose"}},
			allowedFields: []AllowedField{{Field: "user", Path: "u.name"}},
			want: Fields{
				{Field: "u.name", Value: "jose", Type: Text, Conditional: Equal},
			},
		},
		{
			name: "conditional in",
			query: url.Values{
				"name":             {"jose"},
				"conditional-name": {"in"},
			},
			allowedFields: []AllowedField{{Field: "name"}},
			want: Fields{
				{Field: "name", Value: "jose", Type: Text, Conditional: Contains},
			},
		},
		{
			name: "conditional com sufixo -filter",
			query: url.Values{
				"name":                    {"jose"},
				"conditional-name-filter": {"not-in"},
			},
			allowedFields: []AllowedField{{Field: "name"}},
			want: Fields{
				{Field: "name", Value: "jose", Type: Text, Conditional: NotContain},
			},
		},
		{
			name: "conditional is-null dispensa valor",
			query: url.Values{
				"conditional-name": {"is-null"},
			},
			allowedFields: []AllowedField{{Field: "name"}},
			want: Fields{
				{Field: "name", Type: Text, Conditional: IsEmpty},
			},
		},
		{
			name: "conditional not-null preserva ZeroIsEmpty",
			query: url.Values{
				"conditional-age": {"not-null"},
			},
			allowedFields: []AllowedField{{Field: "age", Type: Int32, ZeroIsEmpty: true}},
			want: Fields{
				{Field: "age", Type: Int32, Conditional: NotEmpty, ZeroIsEmpty: true},
			},
		},
		{
			name: "conditional inválida retorna erro",
			query: url.Values{
				"name":             {"jose"},
				"conditional-name": {"xpto"},
			},
			allowedFields: []AllowedField{{Field: "name"}},
			wantErr:       ErrInvalidConditional,
		},
		{
			name:          "texto normalizado",
			query:         url.Values{"name": {"josé"}},
			allowedFields: []AllowedField{{Field: "name", NormalizeText: true}},
			want: Fields{
				{Field: "name", Value: "jose", Type: Text, Conditional: Equal},
			},
		},
		{
			name:          "texto com múltiplos valores",
			query:         url.Values{"name": {"josé", "maria"}},
			allowedFields: []AllowedField{{Field: "name", NormalizeText: true, MultipleValues: ByArray}},
			want: Fields{
				{Field: "name", Value: []string{"jose", "maria"}, Type: Text, Conditional: Equal, MultipleValues: true},
			},
		},
		{
			name:          "inteiro simples",
			query:         url.Values{"age": {"30"}},
			allowedFields: []AllowedField{{Field: "age", Type: Int64}},
			want: Fields{
				{Field: "age", Value: int64(30), Type: Int64, Conditional: Equal},
			},
		},
		{
			name:          "inteiro inválido retorna erro",
			query:         url.Values{"age": {"abc"}},
			allowedFields: []AllowedField{{Field: "age", Type: Int64}},
			wantErr:       strconv.ErrSyntax,
		},
		{
			name:          "inteiro fora do range do tipo retorna erro",
			query:         url.Values{"age": {"130"}},
			allowedFields: []AllowedField{{Field: "age", Type: Int8}},
			wantErr:       strconv.ErrRange,
		},
		{
			name:          "inteiros com múltiplos valores ignora inválidos",
			query:         url.Values{"age": {"1", "2", "abc"}},
			allowedFields: []AllowedField{{Field: "age", Type: Int64, MultipleValues: ByArray}},
			want: Fields{
				{Field: "age", Value: []int64{1, 2}, Type: Int64, Conditional: Equal, MultipleValues: true},
			},
		},
		{
			name:          "faixa de datas",
			query:         url.Values{"created_at": {"2024-01-10/2024-01-20"}},
			allowedFields: []AllowedField{{Field: "created_at", Type: DateRange}},
			want: Fields{
				{
					Field: "created_at",
					Value: daterange.DateRange{
						time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 20, 23, 59, 59, 999999999, time.UTC),
					},
					Type:        DateRange,
					Conditional: Equal,
				},
			},
		},
		{
			name:          "faixa de datas inválida retorna erro",
			query:         url.Values{"created_at": {"10/01/2024"}},
			allowedFields: []AllowedField{{Field: "created_at", Type: DateRange}},
			wantErr:       daterange.ErrParsingStartDateField,
		},
		{
			name:          "faixa de datas não suporta múltiplos valores",
			query:         url.Values{"created_at": {"2024-01-10"}},
			allowedFields: []AllowedField{{Field: "created_at", Type: DateRange, MultipleValues: ByArray}},
			wantErr:       ErrNoSupportMultipleValues,
		},
		{
			name:          "booleano",
			query:         url.Values{"active": {"true"}},
			allowedFields: []AllowedField{{Field: "active", Type: Bool}},
			want: Fields{
				{Field: "active", Value: true, Type: Bool, Conditional: Equal},
			},
		},
		{
			name:          "booleano inválido retorna erro",
			query:         url.Values{"active": {"xpto"}},
			allowedFields: []AllowedField{{Field: "active", Type: Bool}},
			wantErr:       strconv.ErrSyntax,
		},
		{
			name:          "booleano não suporta múltiplos valores",
			query:         url.Values{"active": {"true"}},
			allowedFields: []AllowedField{{Field: "active", Type: Bool, MultipleValues: ByArray}},
			wantErr:       ErrNoSupportMultipleValues,
		},
		{
			name: "vários campos combinados",
			query: url.Values{
				"name":             {"jose"},
				"age":              {"30"},
				"conditional-name": {"in"},
			},
			allowedFields: []AllowedField{
				{Field: "name"},
				{Field: "age", Type: Int32},
			},
			want: Fields{
				{Field: "name", Value: "jose", Type: Text, Conditional: Contains},
				{Field: "age", Value: int64(30), Type: Int32, Conditional: Equal},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Make(tt.query, tt.allowedFields)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Make() erro = nil, esperado %v", tt.wantErr)
				}

				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Make() erro = %v, esperado %v", err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("Make() erro inesperado = %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Make() = %#v, esperado %#v", got, tt.want)
			}
		})
	}
}
