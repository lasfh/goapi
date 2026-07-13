package queryfilter

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/lasfh/goapi/strutils"
)

type AllowedContext struct {
	// Nome do contexto.
	Context string

	// Lista de campos para busca do contexto.
	Fields []FieldFromContext

	// Quando verdadeiro, o campo é ignorado durante o processamento da query.
	IgnoreIf bool
}

type FieldFromContext struct {
	// Nome real do campo a ser filtrado no banco de dados. Se estiver vazio, será utilizado o valor presente em "Context".
	Path string

	// O tipo padrão é string (Str).
	Type FieldType

	// (SQL) Substitui a função padrão "ToSQL" para casos em que ela não pode atender ao requisito.
	CustomToSQL CustomToSQLFunc
}

// FieldsFromSearch tipo representando um slice de Field
type FieldsFromSearch Fields

// FieldsAndSearch conjunto de valores para montagem de consulta.
type FieldsAndSearch struct {
	Fields            Fields
	FieldsFromContext FieldsFromSearch
}

// MakeSearch realiza uma pesquisa com contexto com base nos parâmetros de consulta fornecidos.
// Ele analisa os parâmetros de consulta para obter o termo de pesquisa e o contexto (se fornecido) e retorna
// os campos correspondentes permitidos pelo contexto para realizar a pesquisa.
//
// Parâmetros:
//   - query (url.Values): Os parâmetros de consulta contendo o termo de pesquisa e o contexto.
//   - allowedContexts ([]AllowedContext): Os contextos permitidos e os campos associados a cada contexto.
//
// Retorna:
//   - FieldsFromSearch: Os campos correspondentes para realizar a pesquisa.
//   - error: Retorna um erro se houver algum problema ao analisar os parâmetros de consulta.
func MakeSearch(query url.Values, allowedContexts []AllowedContext) (FieldsFromSearch, error) {
	search := firstNonEmptyValue(query, "search", "q")
	if search == "" {
		return nil, nil
	}

	isNumeric := strutils.IsNumeric(search)

	// Se o contexto for "", "tudo" ou "all" é considerado como não definido.
	context := firstNonEmptyValue(query, "context", "search-context")
	if context == "tudo" || context == "all" {
		context = ""
	}

	var fields FieldsFromSearch

	for _, fromContext := range allowedContexts {
		if fromContext.IgnoreIf || (context != "" && fromContext.Context != context) {
			continue
		}

		for _, field := range fromContext.Fields {
			fieldValue := Field{
				Field:       field.Path,
				Type:        field.Type,
				Conditional: Contains,
				CustomToSQL: field.CustomToSQL,
			}

			if field.Type == Text || field.Type == DateRange {
				fieldValue.Value = search

				fields = append(fields, fieldValue)

				continue
			}

			if isNumeric {
				valueInt, err := strconv.ParseInt(search, 10, int(field.Type))
				if err != nil {
					return nil, fmt.Errorf("%s: %w", fromContext.Context, err)
				}

				fieldValue.Conditional = Equal
				fieldValue.Value = valueInt

				fields = append(fields, fieldValue)
			}
		}
	}

	return fields, nil
}
