package strutils

import (
	"slices"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "String vazia",
			input:    "",
			expected: "",
		},
		{
			name:     "Sem acentos",
			input:    "texto simples",
			expected: "texto simples",
		},
		{
			name:     "Acentos comuns",
			input:    "ação de validação",
			expected: "acao de validacao",
		},
		{
			name:     "Cedilha e til",
			input:    "coração",
			expected: "coracao",
		},
		{
			name:     "Maiúsculas acentuadas",
			input:    "ÀÉÎÕÜ",
			expected: "AEIOU",
		},
		{
			name:     "Acentos misturados com números e símbolos",
			input:    "café nº 1 - São Paulo!",
			expected: "cafe nº 1 - Sao Paulo!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Normalize(tt.input)
			if err != nil {
				t.Fatalf("Normalize(%q) retornou erro inesperado: %v", tt.input, err)
			}

			if result != tt.expected {
				t.Errorf("Normalize(%q) = %q, esperado %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "String vazia",
			input:    "",
			expected: false,
		},
		{
			name:     "Apenas dígitos",
			input:    "1234567890",
			expected: true,
		},
		{
			name:     "Um dígito",
			input:    "7",
			expected: true,
		},
		{
			name:     "Dígitos com letra",
			input:    "12a3",
			expected: false,
		},
		{
			name:     "Número decimal",
			input:    "12.3",
			expected: false,
		},
		{
			name:     "Número negativo",
			input:    "-123",
			expected: false,
		},
		{
			name:     "Dígitos com espaço",
			input:    "12 3",
			expected: false,
		},
		{
			name:     "Apenas letras",
			input:    "abc",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := IsNumeric(tt.input); result != tt.expected {
				t.Errorf("IsNumeric(%q) = %t, esperado %t", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToIntValues(t *testing.T) {
	t.Run("Valores válidos", func(t *testing.T) {
		result := ToIntValues[int64]([]string{"1", "20", "300"}, 64)
		expected := []int64{1, 20, 300}

		if !slices.Equal(result, expected) {
			t.Errorf("ToIntValues() = %v, esperado %v", result, expected)
		}
	})

	t.Run("Ignora valores inválidos", func(t *testing.T) {
		result := ToIntValues[int32]([]string{"1", "abc", "2", "3.5", "", "4"}, 32)
		expected := []int32{1, 2, 4}

		if !slices.Equal(result, expected) {
			t.Errorf("ToIntValues() = %v, esperado %v", result, expected)
		}
	})

	t.Run("Valores negativos", func(t *testing.T) {
		result := ToIntValues[int16]([]string{"-1", "0", "42"}, 16)
		expected := []int16{-1, 0, 42}

		if !slices.Equal(result, expected) {
			t.Errorf("ToIntValues() = %v, esperado %v", result, expected)
		}
	})

	t.Run("Ignora valores fora do intervalo do bitSize", func(t *testing.T) {
		result := ToIntValues[int8]([]string{"127", "128", "-128", "-129"}, 8)
		expected := []int8{127, -128}

		if !slices.Equal(result, expected) {
			t.Errorf("ToIntValues() = %v, esperado %v", result, expected)
		}
	})

	t.Run("Slice vazio", func(t *testing.T) {
		result := ToIntValues[int64](nil, 64)

		if len(result) != 0 {
			t.Errorf("ToIntValues() = %v, esperado slice vazio", result)
		}
	})
}
