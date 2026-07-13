package paginated

type Paginated[T any] struct {
	Results     []T    `json:"results"`
	Offset      uint64 `json:"offset"`
	Limit       int64  `json:"limit"`
	TotalRecord int64  `json:"total_record"`
}

// Remap remapeia os itens em uma estrutura de Paginated para outro tipo definido pelo usuário
// e retorna uma nova estrutura de Paginated com os itens remapeados.
//
// Parâmetros:
//   - paginated *Paginated[T]: A estrutura de Paginated contendo os itens a serem remapeados.
//   - f func(T) Y: A função que define como cada item deve ser remapeado para o novo tipo Y.
//
// Retorno:
//   - *Paginated[Y]: Uma nova estrutura de Paginated contendo os itens remapeados para o tipo Y.
func Remap[T, Y any](paginated *Paginated[T], f func(T) Y) *Paginated[Y] {
	items := make([]Y, 0, len(paginated.Results))

	for i := range paginated.Results {
		items = append(items, f(
			paginated.Results[i],
		))
	}

	return &Paginated[Y]{
		Results:     items,
		Offset:      paginated.Offset,
		Limit:       paginated.Limit,
		TotalRecord: paginated.TotalRecord,
	}
}

// RemapWithError remapeia os resultados de uma lista paginada aplicando uma função especificada e capturando erros.
//
// Parâmetros:
//   - paginated (*Paginated[T]): A lista paginada a ser remapeada.
//   - f (func(T) (Y, error)): A função a ser aplicada a cada item da lista paginada.
//
// Retorna:
//   - (*Paginated[Y]): Uma lista paginada remapeada com os resultados da aplicação da função.
//   - error: Retorna um erro se ocorrer algum erro durante o processo de remapeamento.
func RemapWithError[T, Y any](paginated *Paginated[T], f func(T) (Y, error)) (*Paginated[Y], error) {
	items := make([]Y, 0, len(paginated.Results))

	for i := range paginated.Results {
		item, err := f(
			paginated.Results[i],
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return &Paginated[Y]{
		Results:     items,
		Offset:      paginated.Offset,
		Limit:       paginated.Limit,
		TotalRecord: paginated.TotalRecord,
	}, nil
}

// UpdateItems aplica uma função de atualização a todos os itens de uma lista paginada.
//
// Parâmetros:
//   - paginated (*Paginated[T]): Lista paginada de itens a serem atualizados.
//   - update (func(*T)): Função que modifica cada item da lista.
//
// Retorna:
//   - *Paginated[T]: Lista paginada com os itens atualizados.
func UpdateItems[T any](paginated *Paginated[T], update func(*T)) *Paginated[T] {
	for i := range paginated.Results {
		update(&paginated.Results[i])
	}

	return paginated
}
