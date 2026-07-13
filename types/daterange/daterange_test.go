package daterange

import (
	"errors"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input         string
		expectedStart time.Time
		expectedEnd   time.Time
		expectedError error
	}{
		{
			"2022-01-01",
			time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.January, 1, 23, 59, 59, 999999999, time.UTC),
			nil,
		},
		{
			"2022-01-01/2022-01-10",
			time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.January, 10, 23, 59, 59, 999999999, time.UTC),
			nil,
		},
		{
			"2022-02-30/2022-01-10",
			time.Time{},
			time.Time{},
			ErrParsingStartDateField,
		},
		{
			"invalid-date",
			time.Time{},
			time.Time{},
			ErrParsingStartDateField,
		},
		{
			"2022-01-01/invalid-date",
			time.Time{},
			time.Time{},
			ErrParsingEndDateField,
		},
	}

	for _, test := range tests {
		dates, err := Parse(test.input)

		if !dates.Start().Equal(test.expectedStart) || !dates.End().Equal(test.expectedEnd) || !errors.Is(err, test.expectedError) {
			t.Errorf("Para entrada %s, esperado (%v, %v, %v), mas obtido (%v, %v, %v)",
				test.input, test.expectedStart, test.expectedEnd, test.expectedError, dates.Start(), dates.End(), err)
		}
	}

	_, err := Parse("2022-01-01/2022-01-10", time.Local)
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}
}

func TestDateRange_IsEmpty(t *testing.T) {
	// Teste para uma DateRange vazia
	emptyRange := DateRange{}
	if !emptyRange.IsEmpty() {
		t.Errorf("Esperado verdadeiro para um DateRange vazio, mas obtido falso")
	}

	// Teste para uma DateRange não vazia
	nonEmptyRange := DateRange{
		time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC),
	}
	if nonEmptyRange.IsEmpty() {
		t.Errorf("Esperado falso para um DateRange não vazio, mas foi verdadeiro")
	}
}
