package pagination

import (
	"net/url"
	"reflect"
	"testing"
)

func TestByQuery(t *testing.T) {
	tests := []struct {
		name      string
		query     url.Values
		opts      []Options
		wantLimit int64
		wantSkip  int64
		wantSort  string
		wantOrder ListingOrderType
	}{
		{
			name:      "query vazia – valores padrão",
			query:     url.Values{},
			opts:      []Options{},
			wantLimit: defaultLimit,
			wantSkip:  0,
			wantSort:  "",
			wantOrder: ListingOrderType(0), // inválido
		},
		{
			name: "limit e skip explícitos",
			query: url.Values{
				limitKey: {"10"},
				skipKey:  {"20"},
			},
			opts:      []Options{},
			wantLimit: 10,
			wantSkip:  20,
			wantSort:  "",
			wantOrder: ListingOrderType(0),
		},
		{
			name: "page e limit",
			query: url.Values{
				limitKey: {"10"},
				"page":   {"3"},
			},
			opts:      []Options{},
			wantLimit: 10,
			wantSkip:  20,
			wantSort:  "",
			wantOrder: ListingOrderType(0),
		},
		{
			name: "sort e order válidos",
			query: url.Values{
				sortKey:  {"codigo"},
				orderKey: {"desc"},
				limitKey: {"25"},
			},
			opts:      []Options{},
			wantLimit: 25,
			wantSkip:  0,
			wantSort:  "codigo",
			wantOrder: DescendingOrder,
		},
		{
			name: "sort com order inválido – fallback para options.Default.Order",
			query: url.Values{
				sortKey:  {"codigo"},
				orderKey: {"invalid"},
			},
			opts: []Options{
				{
					Default: &DefaultValues{
						Order: DescendingOrder, // valor padrão a ser usado
					},
				},
			},
			wantLimit: defaultLimit,
			wantSkip:  0,
			wantSort:  "codigo",
			wantOrder: DescendingOrder,
		},
		{
			name: "sort em SortAndOrder com order inválido – usa o order da lista",
			query: url.Values{
				sortKey:  {"codigo"},
				orderKey: {"invalid"},
			},
			opts: []Options{
				{
					Default: &DefaultValues{
						SortAndOrder: []SortAndOrder{
							{Field: "codigo", Path: "item.code", Order: DescendingOrder},
						},
					},
				},
			},
			wantLimit: defaultLimit,
			wantSkip:  0,
			wantSort:  "item.code",
			wantOrder: DescendingOrder,
		},
		{
			name: "order válido no query – precedência sobre Default.Order",
			query: url.Values{
				sortKey:  {"codigo"},
				orderKey: {"asc"},
				limitKey: {"15"},
			},
			opts: []Options{
				{
					Default: &DefaultValues{
						Order: DescendingOrder,
					},
				},
			},
			wantLimit: 15,
			wantSkip:  0,
			wantSort:  "codigo",
			wantOrder: AscendingOrder,
		},
		{
			name: "sort em SortAndOrder e order válido no query",
			query: url.Values{
				sortKey:  {"codigo"},
				orderKey: {"desc"},
			},
			opts: []Options{
				{
					Default: &DefaultValues{
						SortAndOrder: []SortAndOrder{
							{Field: "codigo", Order: DescendingOrder},
							{Field: "data", Order: AscendingOrder},
						},
					},
				},
			},
			wantLimit: defaultLimit,
			wantSkip:  0,
			wantSort:  "codigo",
			wantOrder: DescendingOrder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Constrói a paginação a partir da query e das opções
			got := ByQuery(tt.query, tt.opts...)

			if got.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, esperado %d", got.Limit, tt.wantLimit)
			}
			if got.Skip != tt.wantSkip {
				t.Errorf("Skip = %d, esperado %d", got.Skip, tt.wantSkip)
			}
			if got.Sort != tt.wantSort {
				t.Errorf("Sort = %q, esperado %q", got.Sort, tt.wantSort)
			}
			if got.Order != tt.wantOrder {
				t.Errorf("Order = %v, esperado %v", got.Order, tt.wantOrder)
			}
		})
	}
}

func TestGetOptions(t *testing.T) {
	// Consta que maxLimit e defaultLimit já existem no pacote.
	tests := []struct {
		name string
		in   []Options
		want Options
	}{
		{
			name: "nenhuma sobrescrita – usa valores padrões",
			in:   []Options{},
			want: Options{
				MaxLimit:     maxLimit,
				DefaultLimit: defaultLimit,
				Default:      nil,
			},
		},
		{
			name: "sobrescreve apenas MaxLimit",
			in: []Options{
				{MaxLimit: 2000},
			},
			want: Options{
				MaxLimit:     2000,
				DefaultLimit: defaultLimit,
				Default:      nil,
			},
		},
		{
			name: "sobrescreve apenas DefaultLimit",
			in: []Options{
				{DefaultLimit: 50},
			},
			want: Options{
				MaxLimit:     maxLimit,
				DefaultLimit: 50,
				Default:      nil,
			},
		},
		{
			name: "sobrescreve apenas Default",
			in: []Options{
				{
					Default: &DefaultValues{},
				},
			},
			want: Options{
				MaxLimit:     maxLimit,
				DefaultLimit: defaultLimit,
				Default:      &DefaultValues{},
			},
		},
		{
			name: "múltiplas sobrescritas",
			in: []Options{
				{MaxLimit: 500},
				{DefaultLimit: 30},
				{Default: &DefaultValues{}},
			},
			want: Options{
				MaxLimit:     500,
				DefaultLimit: 30,
				Default:      &DefaultValues{},
			},
		},
		{
			name: "valor zero não sobrescreve",
			in: []Options{
				{MaxLimit: 0, DefaultLimit: 0, Default: nil},
			},
			want: Options{
				MaxLimit:     maxLimit,
				DefaultLimit: defaultLimit,
				Default:      nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getOptions(tt.in...)

			if got.MaxLimit != tt.want.MaxLimit {
				t.Errorf("MaxLimit = %d, esperado %d", got.MaxLimit, tt.want.MaxLimit)
			}
			if got.DefaultLimit != tt.want.DefaultLimit {
				t.Errorf("DefaultLimit = %d, esperado %d", got.DefaultLimit, tt.want.DefaultLimit)
			}
			if (got.Default == nil) != (tt.want.Default == nil) {
				t.Errorf("Default = %v, esperado %v", got.Default, tt.want.Default)
			}

			if got.Default != nil && tt.want.Default != nil {
				if !reflect.DeepEqual(got.Default, tt.want.Default) {
					t.Errorf("Default = %+v, esperado %+v", got.Default, tt.want.Default)
				}
			}
		})
	}
}

func TestGetLimit(t *testing.T) {
	// Valores padrão para uso comum
	opts := Options{
		DefaultLimit: 100,  // limite padrão
		MaxLimit:     1000, // limite máximo aceito
	}

	tests := []struct {
		name string
		s    string  // string de entrada
		opts Options // opções a serem usadas
		want int64   // resultado esperado
	}{
		{"string vazia", "", opts, opts.DefaultLimit},
		{"string inválida", "abc", opts, opts.DefaultLimit},
		{"valor negativo", "-5", opts, opts.DefaultLimit},
		{"zero", "0", opts, opts.DefaultLimit},
		{"valor mínimo (1)", "1", opts, 1},
		{"dentro do limite máximo", "500", opts, 500},
		{"acima do limite máximo", "2000", opts, opts.DefaultLimit},
		{"ignorar limite máximo (IgnoreMaxLimit)", "2000", Options{DefaultLimit: 100, MaxLimit: IgnoreMaxLimit}, 2000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getLimit(tc.s, tc.opts)
			if got != tc.want {
				t.Fatalf("getLimit(%q, %+v) = %d; esperado %d", tc.s, tc.opts, got, tc.want)
			}
		})
	}
}

func TestGetPage(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int64
	}{
		{"string vazia", "", 1},
		{"não numérico", "abc", 1},
		{"zero", "0", 1},
		{"negativo", "-5", 1},
		{"positivo válido", "5", 5},
		{"max int64 + 1 (overflow)", "9223372036854775808", 1},
		{"notação exponencial", "1e3", 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getPage(tc.in)
			if got != tc.want {
				t.Errorf("getPage(%q) = %d; esperado %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestGetSkip(t *testing.T) {
	limit := int64(10)

	tests := []struct {
		name     string
		query    url.Values
		wantSkip int64
	}{
		{
			name:     "nem page nem skip",
			query:    url.Values{},
			wantSkip: 0,
		},
		{
			name:     "página válida",
			query:    url.Values{"page": []string{"3"}},
			wantSkip: (3 - 1) * limit,
		},
		{
			name:     "página inválida (não numérica)",
			query:    url.Values{"page": []string{"abc"}},
			wantSkip: 0,
		},
		{
			name:     "página zero (tratada como 1)",
			query:    url.Values{"page": []string{"0"}},
			wantSkip: 0,
		},
		{
			name:     "skip válido",
			query:    url.Values{skipKey: []string{"15"}},
			wantSkip: 15,
		},
		{
			name:     "skip inválido (não numérico)",
			query:    url.Values{skipKey: []string{"xyz"}},
			wantSkip: 0,
		},
		{
			name:     "tanto page quanto skip são válidas – page tem precedência.",
			query:    url.Values{"page": []string{"4"}, skipKey: []string{"100"}},
			wantSkip: (4 - 1) * limit,
		},
		{
			name:     "page (inválida) e skip",
			query:    url.Values{"page": []string{"-1"}, skipKey: []string{"50"}},
			wantSkip: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getSkip(tc.query, limit)
			if got != tc.wantSkip {
				t.Errorf("getSkip(%v, %d) = %d; esperado %d",
					tc.query, limit, got, tc.wantSkip)
			}
		})
	}
}

func TestGetSort(t *testing.T) {
	// Função mock que será usada no teste “CustomSort”
	customSortFunc := func(v url.Values) string {
		return "custom-sorted"
	}

	tests := []struct {
		name  string
		query url.Values
		opts  Options
		want  string
	}{
		{
			name:  "sort presente na query",
			query: url.Values{sortKey: []string{"name"}}, // sortKey está definido no pacote original
			opts:  Options{},
			want:  "name",
		},
		{
			name:  "sort vazio na query → usa Default.Sort",
			query: url.Values{sortKey: []string{""}},
			opts: Options{
				Default: &DefaultValues{Sort: "default-sort"},
			},
			want: "default-sort",
		},
		{
			name:  "sort ausente → usa Default.Sort",
			query: url.Values{}, // sem “sort”
			opts: Options{
				Default: &DefaultValues{Sort: "default-sort"},
			},
			want: "default-sort",
		},
		{
			name:  "sort ausente → usa Default.CustomSort",
			query: url.Values{}, // sem “sort”
			opts: Options{
				Default: &DefaultValues{CustomSort: customSortFunc},
			},
			want: "custom-sorted",
		},
		{
			name:  "sort ausente e Default nil → string vazio",
			query: url.Values{}, // sem “sort”
			opts:  Options{},
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getSort(tc.query, tc.opts)
			if got != tc.want {
				t.Errorf("getSort(%v, %v) = %q; esperado %q", tc.query, tc.opts, got, tc.want)
			}
		})
	}
}

func TestParseListingOrderType(t *testing.T) {
	// Caso de teste 1: String "1", deve retornar AscendingOrder
	result := parseListingOrderType("1")
	expected := AscendingOrder
	if result != expected {
		t.Errorf("Resultado inesperado. Esperado: %d, obtido: %d", expected, result)
	}

	// Caso de teste 2: String "-1", deve retornar DescendingOrder
	result = parseListingOrderType("-1")
	expected = DescendingOrder
	if result != expected {
		t.Errorf("Resultado inesperado. Esperado: %d, obtido: %d", expected, result)
	}

	// Caso de teste 3: String inválida, deve ser inválido
	result = parseListingOrderType("0")
	if result.IsValid() {
		t.Error("Esperava-se que fosse inválido, mas não ocorreu")
	}
}

func TestGetOrder(t *testing.T) {
	invalidOrder := ListingOrderType(0)

	tests := []struct {
		name  string
		order ListingOrderType
		opts  Options
		want  ListingOrderType
	}{
		{
			name:  "order válido – AscendingOrder",
			order: AscendingOrder,
			opts:  Options{}, // não importa
			want:  AscendingOrder,
		},
		{
			name:  "order válido – DescendingOrder",
			order: DescendingOrder,
			opts:  Options{},
			want:  DescendingOrder,
		},
		{
			name:  "order inválido – opts.Default nil → padrão",
			order: invalidOrder,
			opts:  Options{},
			want:  AscendingOrder, // default
		},
		{
			name:  "order inválido – Default.Order válido",
			order: invalidOrder,
			opts: Options{
				Default: &DefaultValues{ // estrutura que o pacote usa
					Order: DescendingOrder,
				},
			},
			want: DescendingOrder,
		},
		{
			name:  "order inválido – Default.Order inválido → padrão",
			order: invalidOrder,
			opts: Options{
				Default: &DefaultValues{ // valor inválido
					Order: invalidOrder,
				},
			},
			want: AscendingOrder,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getOrder(tc.order, tc.opts)
			if got != tc.want {
				t.Fatalf("Obtido %d, esperado %d", got, tc.want)
			}
		})
	}
}

func BenchmarkByQuery(b *testing.B) {
	query := url.Values{
		"sort":  {"nome_classe"},
		"order": {"1"},
		"limit": {"5"},
		"skip":  {"0"},
	}

	// Verifica o desempenho de execução da função 'ByQuery'.
	for i := 0; i < b.N; i++ {
		ByQuery(query, Options{
			DefaultLimit: 50,
			MaxLimit:     100,
			Default: &DefaultValues{
				SortAndOrder: []SortAndOrder{
					{
						Field: "nome",
						Path:  "itens.nome",
						Order: DescendingOrder,
					},
					{
						Field: "nome1",
						Path:  "itens.nome1",
						Order: AscendingOrder,
					},
				},
			},
		})
	}
}
