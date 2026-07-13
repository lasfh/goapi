package sliceutils

import (
	"slices"
	"testing"
)

func TestUnique(t *testing.T) {
	t.Run("Strings", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []string
			expected []string
		}{
			{
				name:     "Slice vazio",
				input:    []string{},
				expected: []string{},
			},
			{
				name:     "Sem duplicados",
				input:    []string{"a", "b", "c"},
				expected: []string{"a", "b", "c"},
			},
			{
				name:     "Com duplicados",
				input:    []string{"a", "b", "a", "c", "b"},
				expected: []string{"a", "b", "c"},
			},
			{
				name:     "Todos iguais",
				input:    []string{"x", "x", "x"},
				expected: []string{"x"},
			},
			{
				name:     "Mantém a ordem da primeira ocorrência",
				input:    []string{"c", "a", "c", "b", "a"},
				expected: []string{"c", "a", "b"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Unique(tt.input)
				if !slices.Equal(result, tt.expected) {
					t.Errorf("Unique(%v) = %v, esperado %v", tt.input, result, tt.expected)
				}
			})
		}
	})

	t.Run("Inteiros", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []int
			expected []int
		}{
			{
				name:     "Slice nil",
				input:    nil,
				expected: []int{},
			},
			{
				name:     "Com duplicados",
				input:    []int{1, 2, 2, 3, 1, 4},
				expected: []int{1, 2, 3, 4},
			},
			{
				name:     "Elemento único",
				input:    []int{7},
				expected: []int{7},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Unique(tt.input)
				if !slices.Equal(result, tt.expected) {
					t.Errorf("Unique(%v) = %v, esperado %v", tt.input, result, tt.expected)
				}
			})
		}
	})
}
