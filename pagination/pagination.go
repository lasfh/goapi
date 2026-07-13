package pagination

import (
	"net/url"
	"slices"
	"strconv"
)

const (
	// Ignorar a validação do limite máximo de itens por página.
	IgnoreMaxLimit int64 = -1

	defaultLimit int64 = 5
	maxLimit     int64 = 50
)

var (
	skipKey  = "offset"
	limitKey = "limit"
	orderKey = "order"
	sortKey  = "sort"
)

// ListingOrderType representa o tipo de ordenação (ascendente ou descendente).
type ListingOrderType int8

const (
	// AscendingOrder representa a ordenação ascendente.
	AscendingOrder ListingOrderType = 1

	// DescendingOrder representa a ordenação descendente.
	DescendingOrder ListingOrderType = -1
)

// IsValid verifica se o tipo de ordenação é válido.
func (l ListingOrderType) IsValid() bool {
	switch l {
	case AscendingOrder, DescendingOrder:
		return true
	}

	return false
}

type Options struct {
	// Limite máximo de valores por página.
	// Se for igual a IgnoreMaxLimit (-1), a validação será ignorada.
	// Padrão: 50
	MaxLimit int64

	// Valor padrão para limite por página.
	// Padrão: 5
	DefaultLimit int64

	// Valores padrão, caso definido.
	Default *DefaultValues
}

// HasSortAndOrder verifica se o campo "sort" especificado possui valores personalizações.
func (o *Options) hasSortAndOrder(sort string) int {
	// Verifica se o receptor ou o campo Default é nulo.
	if o.Default == nil {
		// Retorna -1 se o receptor ou o Default não estiverem definidos.
		return -1
	}

	// Procura o índice do campo de ordenação especificado dentro do slice SortAndOrder.
	return slices.IndexFunc(o.Default.SortAndOrder, func(item SortAndOrder) bool {
		return item.Field == sort
	})
}

// DefaultValues contém valores padrão para ordenação.
type DefaultValues struct {
	// Lista de valores padrão para determinados nomes de campo.
	// Permite personalizar o nome do campo obtido de "sort" a ser
	// utilizado no banco de dados (path) e definir um valor padrão de ordenação exclusivo.
	SortAndOrder []SortAndOrder

	Sort       string
	CustomSort func(url.Values) string
	Order      ListingOrderType
}

// GetSortAndOrder retorne os valores para classificar e ordenar com base nos valores padrão da lista.
func (v *DefaultValues) getSortAndOrder(index int, sort string, order ListingOrderType) (string, ListingOrderType) {
	if v.SortAndOrder[index].Path != "" {
		sort = v.SortAndOrder[index].FieldName()
	}

	// Se não houver "order" ou for inválido, use o padrão.
	if !order.IsValid() {
		order = v.SortAndOrder[index].Order
	}

	return sort, order
}

// SortAndOrder contém os valores para ordenar e personalizar o caminho de um campo
type SortAndOrder struct {
	// Valor do "sort" específico que deseja personalizar.
	Field string

	// Nome real do campo no banco de dados. Se estiver vazio, será utilizado o valor presente em "Field".
	Path string

	// Valor padrão exclusivo para ordenação.
	Order ListingOrderType
}

// FieldName retorna o nome do campo no banco de dados.
func (s SortAndOrder) FieldName() string {
	if s.Path != "" {
		return s.Path
	}

	return s.Field
}

// Pagination representa os parâmetros de paginação e ordenação.
type Pagination struct {
	Limit int64
	Skip  int64
	Sort  string
	Order ListingOrderType
}

// ByQuery realiza a leitura de parâmetros de consulta (query) e opções adicionais para construir
// um objeto de paginação (Pagination) com base nesses parâmetros.
//
// Parâmetros:
//   - query: Um objeto url.Values contendo os parâmetros de consulta, geralmente recebidos de uma solicitação HTTP.
//   - opts: Opções adicionais que podem ser fornecidas para personalizar o processo de construção do objeto de paginação.
//
// Retorno:
//   - Um objeto do tipo Pagination contendo informações sobre limites, ordenação e paginação
//     com base nos parâmetros de consulta.
//   - Um erro que indica se houve algum problema durante o processo de construção, sendo nulo em caso de sucesso.
//
// Exemplo de Uso:
//
//	paging, err := pagination.ByQuery(query, pagination.Options{
//		Default: &pagination.DefaultValues{
//			Sort:  "codigo",
//			Order: pagination.AscendingOrder,
//		},
//	})
//	if err != nil {
//		httpview.SendError(
//			w,
//			httperror.ErrPaginationByQuery.WithWarn(err),
//		)
//
//		return
//	}
func ByQuery(query url.Values, opts ...Options) Pagination {
	options := getOptions(opts...)

	limit := getLimit(
		query.Get(limitKey),
		options,
	)

	pagination := Pagination{
		Limit: limit,
		Skip: getSkip(
			query, limit,
		),
	}

	sort := getSort(query, options)
	if sort == "" {
		return pagination
	}

	order := parseListingOrderType(
		query.Get(orderKey),
	)

	// Verifica se possui valores padrão (nome do caminho personalizado e/ou ordenação) para o campo "sort" fornecido.
	if index := options.hasSortAndOrder(sort); index >= 0 {
		pagination.Sort, pagination.Order = options.Default.getSortAndOrder(
			index, sort, order,
		)
	} else {
		pagination.Sort = sort
	}

	// Verifica se já foi definido um valor para a "order".
	// Os valores obtidos em "url.Values" ou na lista "SortAndOrder" têm precedência, se válidos.
	if !pagination.Order.IsValid() {
		pagination.Order = getOrder(
			order, options,
		)
	}

	return pagination
}

// getOptions obtém as opções de paginação a partir de parâmetros opcionais.
func getOptions(opts ...Options) Options {
	options := Options{
		MaxLimit:     maxLimit,
		DefaultLimit: defaultLimit,
	}

	for _, opt := range opts {
		if opt.MaxLimit != 0 {
			options.MaxLimit = opt.MaxLimit
		}

		if opt.DefaultLimit != 0 {
			options.DefaultLimit = opt.DefaultLimit
		}

		if opt.Default != nil {
			options.Default = opt.Default
		}
	}

	return options
}

// getLimit obtém o valor do limite com base nos parâmetros de consulta.
func getLimit(s string, opts Options) int64 {
	if s == "" {
		return opts.DefaultLimit
	}

	limit, err := strconv.ParseInt(s, 10, 64)
	if err != nil || limit < 1 {
		return opts.DefaultLimit
	}

	// Se verdadeiro, a validação de limite máximo será ignorada.
	if opts.MaxLimit == IgnoreMaxLimit {
		return limit
	}

	if limit > opts.MaxLimit {
		return opts.DefaultLimit
	}

	return limit
}

// getPage converte a string de página para int64, retornando 1 em caso de erro.
func getPage(s string) int64 {
	if s == "" {
		return 1
	}

	page, err := strconv.ParseInt(s, 10, 64)
	if err != nil || page <= 0 {
		return 1
	}

	return page
}

// getSkip calcula o valor de skip para a paginação.
func getSkip(query url.Values, limit int64) int64 {
	if query.Has("page") {
		page := getPage(
			query.Get("page"),
		)

		return (page - 1) * limit
	}

	if value := query.Get(skipKey); value != "" {
		skip, _ := strconv.ParseInt(value, 10, 64)

		return skip
	}

	return 0
}

// Sort retorna o campo "sort" baseado nos parâmetros da query ou nas opções padrão.
func getSort(query url.Values, opts Options) string {
	sort := query.Get(sortKey)

	// Se o valor não for encontrado na query string, verifica as opções padrão.
	if sort == "" && opts.Default != nil {
		// Se houver uma função CustomSort definida nas opções padrão, invoca essa função.
		if opts.Default.CustomSort != nil {
			return opts.Default.CustomSort(query)
		}

		return opts.Default.Sort
	}

	return sort
}

// getOrder obtém o tipo de ordenação com base nos parâmetros de consulta.
func getOrder(order ListingOrderType, opts Options) ListingOrderType {
	if !order.IsValid() {
		if opts.Default != nil && opts.Default.Order.IsValid() {
			return opts.Default.Order
		}

		return AscendingOrder
	}

	return order
}

// parseListingOrderType converte uma string em um tipo de ordenação.
func parseListingOrderType(s string) ListingOrderType {
	switch s {
	case "1", "asc", "ASC":
		return AscendingOrder
	case "-1", "desc", "DESC":
		return DescendingOrder
	}

	return 0
}
