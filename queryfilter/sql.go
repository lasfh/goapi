package queryfilter

import (
	"fmt"
	"strings"
)

type CustomToSQLFunc func(field Field) (string, []any)

// ToSQLEqual converte o campo em um filtro para condicional "igual".
//
// Retorna:
//   - string: Cláusula SQL correspondente ao campo e seus valores.
//   - []any: Argumentos a serem substituídos nos placeholders da cláusula.
func (f Field) ToSQLEqual() (string, []any) {
	var query strings.Builder

	query.WriteString(f.Field)

	if f.MultipleValues {
		query.WriteString(" IN (?)")
	} else {
		query.WriteString(" = ?")
	}

	return query.String(), []any{f.Value}
}

// ToSQLContains converte o campo em um filtro para condicional "contém".
//
// Retorna:
//   - string: Cláusula SQL correspondente ao campo e seus valores.
//   - []any: Argumentos a serem substituídos nos placeholders da cláusula.
func (f Field) ToSQLContains() (string, []any) {
	if f.Type != Text {
		return f.ToSQLEqual()
	}

	if f.MultipleValues {
		var query strings.Builder
		query.WriteRune('(')

		values, _ := f.Value.([]string)
		args := make([]any, 0, len(values))

		for index, value := range values {
			query.WriteString("LOWER(")
			query.WriteString(f.Field)
			query.WriteString(") LIKE LOWER(?)")

			if index < len(values)-1 {
				query.WriteString(" OR ")
			}

			args = append(args, "%"+value+"%")
		}

		query.WriteRune(')')

		return query.String(), args
	}

	query := fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", f.Field)

	return query, []any{"%" + f.Value.(string) + "%"}
}

// ToSQLNotContains converte o campo em um filtro para condicional "não contém".
//
// Retorna:
//   - string: Cláusula SQL correspondente ao campo e seus valores.
//   - []any: Argumentos a serem substituídos nos placeholders da cláusula.
func (f Field) ToSQLNotContains() (string, []any) {
	if f.Type != Text {
		if f.MultipleValues {
			return f.Field + " NOT IN (?)", []any{f.Value}
		}

		return f.Field + " != ?", []any{f.Value}
	}

	if f.MultipleValues {
		var query strings.Builder
		query.WriteRune('(')

		values, _ := f.Value.([]string)
		args := make([]any, 0, len(values))

		for index, value := range values {
			query.WriteString("LOWER(")
			query.WriteString(f.Field)
			query.WriteString(") NOT LIKE LOWER(?)")

			if index < len(values)-1 {
				query.WriteString(" AND ")
			}

			args = append(args, "%"+value+"%")
		}

		query.WriteRune(')')

		return query.String(), args
	}

	query := fmt.Sprintf("LOWER(%s) NOT LIKE LOWER(?)", f.Field)

	return query, []any{"%" + f.Value.(string) + "%"}
}

// ToSQLIsEmpty converte o campo em um filtro para condicional "está vazio".
//
// Retorna:
//   - string: Cláusula SQL correspondente ao campo e seus valores.
//   - []any: Argumentos a serem substituídos nos placeholders da cláusula.
func (f Field) ToSQLIsEmpty() (string, []any) {
	if f.Type == Text || (f.ZeroIsEmpty && f.Type.IsInteger()) {
		var query strings.Builder
		query.WriteRune('(')
		query.WriteString(f.Field)
		query.WriteString(" IS NULL OR ")
		query.WriteString(f.Field)

		if f.Type == Text {
			query.WriteString(" = ''")
		} else {
			query.WriteString(" = 0")
		}

		query.WriteRune(')')

		return query.String(), nil
	}

	return fmt.Sprintf("%s IS NULL", f.Field), nil
}

// ToSQLNotEmpty converte o campo em um filtro para condicional "não está vazio".
//
// Retorna:
//   - string: Cláusula SQL correspondente ao campo e seus valores.
//   - []any: Argumentos a serem substituídos nos placeholders da cláusula.
func (f Field) ToSQLNotEmpty() (string, []any) {
	if f.Type == Text || (f.ZeroIsEmpty && f.Type.IsInteger()) {
		var query strings.Builder
		query.WriteRune('(')
		query.WriteString(f.Field)
		query.WriteString(" IS NOT NULL AND ")
		query.WriteString(f.Field)

		if f.Type == Text {
			query.WriteString(" != ''")
		} else {
			query.WriteString(" != 0")
		}

		query.WriteRune(')')

		return query.String(), nil
	}

	return fmt.Sprintf("%s IS NOT NULL", f.Field), nil
}

// ToSQL converte o campo Field para uma cláusula SQL,
// representando uma condição de filtro para uma consulta.
func (f Field) ToSQL() (string, []any) {
	if f.CustomToSQL != nil {
		return f.CustomToSQL(f)
	}

	switch f.Conditional {
	case Contains:
		return f.ToSQLContains()
	case NotContain:
		return f.ToSQLNotContains()
	case IsEmpty:
		return f.ToSQLIsEmpty()
	case NotEmpty:
		return f.ToSQLNotEmpty()
	}

	return f.ToSQLEqual()
}

// ToSQLQuery converte uma lista de campos (Fields) em uma cláusula SQL,
// representando uma consulta completa de filtro para uma tabela.
func (f Fields) ToSQLQuery() (string, []any) {
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

	return strings.Join(parts, " AND "), args
}
