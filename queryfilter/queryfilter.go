package queryfilter

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"

	"github.com/lasfh/goapi/strutils"
	"github.com/lasfh/goapi/types/daterange"
)

var (
	ErrInvalidConditional      = errors.New("condicional inválida")
	ErrNoSupportMultipleValues = errors.New("não há suporte para vários valores")
)

const (
	conditionalFieldPrefix = "conditional-"
	filterFieldSufix       = "-filter"
)

// conditional tipo que representa uma condicional (em uint8) para o filtro.
type Conditional uint8

const (
	Equal      Conditional = iota // Igual
	Contains                      // Contém
	NotContain                    // Não contém
	IsEmpty                       // Está vazio
	NotEmpty                      // Não está vazio
)

// IsNegation verifica se a condição é uma das condições de negação.
//
// Parâmetros:
//   - c (Conditional): A condição a ser verificada.
//
// Retorna:
//   - bool: Retorna true se a condição for uma negação, caso contrário, retorna false.
func (c Conditional) IsNegation() bool {
	switch c {
	case NotContain, NotEmpty:
		return true
	}

	return false
}

// conditional tipo que representa uma condicional (em string) para o filtro vindo do frontend.
type conditional string

const (
	equal      conditional = "equal"    // Igual
	contains   conditional = "in"       // Contém
	notContain conditional = "not-in"   // Não contém
	isEmpty    conditional = "is-null"  // Está vazio
	notEmpty   conditional = "not-null" // Não está vazio

	// Mesmo comportamento do condicional "Igual".
	undefined conditional = ""
)

// NotUseValue informa se o tipo condicional necessita do campo valor.
func (c conditional) NotUseValue() (Conditional, bool) {
	switch c {
	case isEmpty:
		return IsEmpty, true
	case notEmpty:
		return NotEmpty, true
	}

	return 0, false
}

// Conditional converte um Conditional (string) em conditional (uint8),
// pode retornar um erro se for inválido.
func (c conditional) Conditional() (Conditional, error) {
	switch c {
	case undefined, equal:
		return Equal, nil
	case contains:
		return Contains, nil
	case notContain:
		return NotContain, nil
	case isEmpty:
		return IsEmpty, nil
	case notEmpty:
		return NotEmpty, nil
	}

	return Equal, ErrInvalidConditional
}

// FieldType representa os tipos de valores disponíveis no Field.
type FieldType uint8

const (
	Text      FieldType = 0  // String
	Bool      FieldType = 1  // Boolean
	DateRange FieldType = 2  // Faixa de datas
	DateTime  FieldType = 3  // DateTime somente para IsEmpty/NotEmpty no momento.
	Int8      FieldType = 8  // int8
	Int16     FieldType = 16 // int16
	Int32     FieldType = 32 // int32
	Int64     FieldType = 64 // int64
)

// IsNumber retorna verdadeiro se o tipo definido para o campo for do tipo numérico.
func (f FieldType) IsNumber() bool {
	return f.IsInteger()
}

// IsInteger retorna verdadeiro se o tipo definido para o campo for do tipo inteiro.
func (f FieldType) IsInteger() bool {
	switch f {
	case Int8, Int16, Int32, Int64:
		return true
	}

	return false
}

// IsDate retorna verdadeiro se o tipo definido for data.
func (f FieldType) IsDate() bool {
	switch f {
	case DateRange, DateTime:
		return true
	}

	return false
}

// MultipleValues define um tipo para representar diferentes formas de múltiplos valores.
type MultipleValues uint8

const (
	NoMultipleValues MultipleValues = iota

	// (Recomendado) Indica que múltiplos valores são recebidos como array (ex.: ?user=1&user=2&user=3).
	ByArray
)

// Is retorna verdadeiro se a opção selecionada permite múltiplos valores.
func (m MultipleValues) Is() bool {
	return m != NoMultipleValues
}

type AllowedField struct {
	// Nome do campo recebido no filtro.
	Field string

	// Nome real do campo a ser filtrado no banco de dados. Se estiver vazio, será utilizado o valor presente em ‘Field’.
	Path string

	// O tipo padrão é string (Str).
	Type FieldType

	// Considera o número ZERO como vazio. Usado nas condições IsEmpty/NotEmpty quando o tipo é numérico.
	ZeroIsEmpty bool

	// Opções: ByArray, NoMultipleValues (valor padrão). Recurso válido para todos os tipos, exceto 'DateRange' e 'Bool'.
	MultipleValues MultipleValues

	// Válido somente para o tipo 'Str'
	NormalizeText bool

	// (SQL) Substitui a função padrão "ToSQL" para casos em que ela não pode atender ao requisito.
	CustomToSQL CustomToSQLFunc

	// Quando verdadeiro, o campo é ignorado durante o processamento da query.
	IgnoreIf bool
}

// FieldName retorna o nome do campo no banco de dados.
func (a AllowedField) FieldName() string {
	if a.Path != "" {
		return a.Path
	}

	return a.Field
}

// Field informações obtidas de um campo para montagem de filtro.
type Field struct {
	Field                   string
	Value                   any
	Type                    FieldType
	Conditional             Conditional
	WithExpr                bool
	ZeroIsEmpty             bool
	NotStrictlyEmptyInLists bool
	MultipleValues          bool
	CustomToSQL             CustomToSQLFunc
}

// Fields tipo representando um slice de Field
type Fields []Field

// Make realiza uma análise dos parâmetros de consulta (query) e filtra os campos permitidos,
// retornando uma lista de campos filtrados (Fields) e um possível erro.
// Parâmetros:
//   - query (url.Values): Um objeto contendo os parâmetros da consulta HTTP.
//   - allowedFields ([]AllowedField): Uma lista que especifica quais campos são permitidos
//     na consulta e seu tipo correspondente.
//
// Retorno:
//   - Fields: Uma lista de campos filtrados com seus valores e condições correspondentes.
//   - error: Um erro, se ocorrer algum problema durante o processamento da consulta.
func Make(query url.Values, allowedFields []AllowedField) (Fields, error) {
	var (
		cap    = min(len(query), len(allowedFields))
		fields = make(Fields, 0, cap)
	)

	for _, field := range allowedFields {
		if field.IgnoreIf {
			continue
		}

		// Se nenhuma condição for encontrada, o padrão (Equal) será usado.
		cond := Equal

		// Nome padrão para o campo de condicional
		keyConditional := conditionalFieldPrefix + field.Field

		rawCond := query.Get(keyConditional)
		if rawCond == "" {
			// Em alguns caso é necessário adicionar o sufixo "-filter"
			// no campo "name" da estrutura de filtro do frontend.
			rawCond = query.Get(keyConditional + filterFieldSufix)
		}

		if rawCond != "" {
			condStr := conditional(rawCond)

			if conditional, notUse := condStr.NotUseValue(); notUse {
				fields = append(fields, Field{
					Field:       field.FieldName(),
					Type:        field.Type,
					Conditional: conditional,
					ZeroIsEmpty: field.ZeroIsEmpty,
					CustomToSQL: field.CustomToSQL,
				})

				continue
			}

			c, err := condStr.Conditional()
			if err != nil {
				return nil, fmt.Errorf("%s (%s): %w", field.Field, condStr, err)
			}

			cond = c
		}

		values := query[field.Field]
		if len(values) == 0 || (len(values) == 1 && values[0] == "") {
			continue
		}

		fieldValue := Field{
			Field:          field.FieldName(),
			Type:           field.Type,
			Conditional:    cond,
			ZeroIsEmpty:    field.ZeroIsEmpty,
			MultipleValues: field.MultipleValues.Is(),
			CustomToSQL:    field.CustomToSQL,
		}

		if field.Type.IsInteger() {
			value, err := valueByInt(field, values)
			if err != nil {
				return nil, err
			}

			fieldValue.Value = value
			fields = append(fields, fieldValue)

			continue
		}

		if field.Type == DateRange {
			value, err := valueByDateRange(field, values)
			if err != nil {
				return nil, err
			}

			fieldValue.Value = value
			fields = append(fields, fieldValue)

			continue
		}

		if field.Type == Bool {
			value, err := valueByBool(field, values)
			if err != nil {
				return nil, err
			}

			fieldValue.Value = value
			fields = append(fields, fieldValue)

			continue
		}

		value, err := valueByStr(field, values)
		if err != nil {
			return nil, err
		}

		fieldValue.Value = value
		fields = append(fields, fieldValue)
	}

	return slices.Clip(fields), nil
}

// valueByStr processa os valores de um campo permitido e os retorna no formato adequado.
//
// Parâmetros:
//   - field (AllowedField): Campo permitido contendo configurações de processamento dos valores.
//   - values ([]string): Lista de valores a serem processados.
//
// Retorna:
//   - any: Valor processado conforme as configurações do campo.
//   - error: Erro ocorrido durante o processamento, se houver.
func valueByStr(field AllowedField, values []string) (any, error) {
	if field.MultipleValues == ByArray {
		if field.NormalizeText {
			for index := range values {
				value, err := strutils.Normalize(
					values[index],
				)
				if err != nil {
					return nil, fmt.Errorf("%s[%s]: %w", field.Field, values[index], err)
				}

				values[index] = value
			}
		}

		return values, nil
	}

	if field.NormalizeText {
		v, err := strutils.Normalize(
			values[0],
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", field.Field, err)
		}

		return v, nil
	}

	return values[0], nil
}

// valueByInt processa os valores de um campo permitido e os converte para números inteiros.
//
// Parâmetros:
//   - field (AllowedField): Campo permitido contendo configurações de processamento dos valores numéricos.
//   - values ([]string): Lista de valores em formato de string a serem convertidos.
//
// Retorna:
//   - any: Valor numérico processado conforme as configurações do campo.
//   - error: Erro ocorrido durante a conversão, se houver.
func valueByInt(field AllowedField, values []string) (any, error) {
	if field.MultipleValues == ByArray {
		return strutils.ToIntValues[int64](
			values, int(field.Type),
		), nil
	}

	valueInt, err := strconv.ParseInt(
		values[0], 10, int(field.Type),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", field.Field, err)
	}

	return valueInt, nil
}

// valueByDateRange processa um valor de intervalo de datas para um campo permitido.
//
// Parâmetros:
//   - field (AllowedField): Campo permitido contendo configurações do intervalo de datas.
//   - values ([]string): Lista contendo um único valor representando o intervalo de datas.
//
// Retorna:
//   - any: Intervalo de datas processado.
//   - error: Erro ocorrido durante a conversão, se houver.
func valueByDateRange(field AllowedField, values []string) (any, error) {
	// Para o cenário atual, não é justificado o suporte a vários intervalos de datas.
	if field.MultipleValues.Is() {
		return nil, fmt.Errorf("%s (DateRange): %w", field.Field, ErrNoSupportMultipleValues)
	}

	dates, err := daterange.Parse(values[0])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", field.Field, err)
	}

	return dates, nil
}

// valueByBool converte um valor de string para booleano para um campo permitido.
//
// Parâmetros:
//   - field (AllowedField): Campo permitido contendo configurações de validação.
//   - values ([]string): Lista contendo um único valor a ser convertido.
//
// Retorna:
//   - any: Valor booleano convertido.
//   - error: Erro ocorrido durante a conversão, se houver.
func valueByBool(field AllowedField, values []string) (any, error) {
	if field.MultipleValues.Is() {
		return nil, fmt.Errorf("%s (Bool): %w", field.Field, ErrNoSupportMultipleValues)
	}

	boolValue, err := strconv.ParseBool(values[0])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", field.Field, err)
	}

	return boolValue, nil
}
