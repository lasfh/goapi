package queryfilter

import "strings"

// ToSQLQuery converte uma lista de campos (Field) em uma cláusula SQL,
// representando uma consulta completa de filtro com OR entre os campos.
func (f FieldsFromSearch) ToSQLQuery() (string, []any) {
	parts := make([]string, 0, len(f))
	args := make([]any, 0)

	for _, field := range f {
		query, fieldArgs := field.ToSQL()
		if query == "" {
			continue
		}

		parts = append(parts, query)
		args = append(args, fieldArgs...)
	}

	if len(parts) == 0 {
		return "", nil
	}

	return "(" + strings.Join(parts, " OR ") + ")", args
}

// ToSQLQuery converte os valores dos campos 'Fields' e 'FieldsFromContext' em uma cláusula SQL,
// representando uma consulta completa de filtro.
func (f FieldsAndSearch) ToSQLQuery() (string, []any) {
	query, args := f.Fields.ToSQLQuery()

	if f.FieldsFromContext != nil {
		contextQuery, contextArgs := f.FieldsFromContext.ToSQLQuery()
		if contextQuery != "" {
			if query != "" {
				query += " AND " + contextQuery
			} else {
				query = contextQuery
			}

			args = append(args, contextArgs...)
		}
	}

	return query, args
}
