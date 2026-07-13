package daterange

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrParsingStartDateField = errors.New("erro ao analisar o campo de data de início")
	ErrParsingEndDateField   = errors.New("erro ao analisar o campo de data de término")
)

// DateRange é um tipo que representa uma faixa de datas, com um início e um fim.
type DateRange [2]time.Time

// Start retorna a data de início da faixa.
func (d DateRange) Start() time.Time {
	return d[0]
}

// End retorna a data final da faixa.
func (d DateRange) End() time.Time {
	return d[1]
}

// IsEmpty verifica se a faixa de datas está vazio, ou seja, ambos os limites são zero.
func (d DateRange) IsEmpty() bool {
	return d[0].IsZero() && d[1].IsZero()
}

// Parse analisa uma string de intervalo de datas no formato "YYYY-MM-DD/YYYY-MM-DD" ou "YYYY-MM-DD"
// e retorna um DateRange correspondente. Se apenas uma data for fornecida, ela é tratada como a data de início
// do intervalo, e o final do intervalo é definido para o final do dia da data de início no fuso horário especificado.
// Se ambas as datas forem fornecidas, elas são interpretadas como a data de início e de término do intervalo, respectivamente.
// Parâmetros:
//   - dateRangeString: A string que representa o intervalo de datas no formato especificado.
//   - loc: Opcional. O fuso horário utilizado para interpretar as datas. Se nenhum fuso horário for fornecido,
//     o UTC será usado como padrão.
//
// Retorno:
//   - DateRange: Um DateRange representando o intervalo de datas analisado.
//   - error: Um erro, se ocorrer algum problema durante a análise da string de intervalo de datas.
//     Os erros possíveis incluem ErroParsingStartDateField e ErrParsingEndDateField.
func Parse(dateRangeString string, loc ...*time.Location) (DateRange, error) {
	var location *time.Location

	// Use o fuso horário fornecido ou use o fuso horário local se nenhum for fornecido
	if len(loc) > 0 && loc[0] != nil {
		location = loc[0]
	} else {
		location = time.UTC
	}

	trimmed := strings.TrimSpace(dateRangeString)
	parts := strings.Split(trimmed, "/")

	if len(parts) != 2 {
		dataInicial, err := time.ParseInLocation("2006-01-02", trimmed, location)
		if err != nil {
			return DateRange{}, fmt.Errorf("%w: %w", ErrParsingStartDateField, err)
		}

		dataFinal := time.Date(
			dataInicial.Year(),
			dataInicial.Month(),
			dataInicial.Day(),
			23, 59, 59, 999999999,
			location,
		)

		return DateRange{dataInicial, dataFinal}, nil
	}

	dataInicial, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(parts[0]), location)
	if err != nil {
		return DateRange{}, fmt.Errorf("%w: %w", ErrParsingStartDateField, err)
	}

	dataFinal, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(parts[1]), location)
	if err != nil {
		return DateRange{}, fmt.Errorf("%w: %w", ErrParsingEndDateField, err)
	}

	dataFinal = time.Date(
		dataFinal.Year(),
		dataFinal.Month(),
		dataFinal.Day(),
		23, 59, 59, 999999999,
		location,
	)

	return DateRange{dataInicial, dataFinal}, nil
}
