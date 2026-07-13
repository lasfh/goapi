package paginated

import (
	"errors"
	"reflect"
	"testing"
)

func TestRemap(t *testing.T) {
	// Resultados esperados após o remapeamento
	expectedPaginated := &Paginated[int]{
		Results:     []int{2, 3, 4},
		Offset:      0,
		Limit:       3,
		TotalRecord: 3,
	}

	// Chamada da função a ser testada
	result := Remap(&Paginated[int]{
		Results:     []int{1, 2, 3},
		Offset:      0,
		Limit:       3,
		TotalRecord: 3,
	}, func(input int) int {
		// Implementação de exemplo: simplesmente adiciona 1 ao input
		return input + 1
	})

	// Verifica se os resultados remapeados correspondem ao esperado
	if !reflect.DeepEqual(result, expectedPaginated) {
		t.Errorf("Incompatibilidade de resultado. Esperado: %v, obtido: %v", expectedPaginated, result)
	}
}

func TestRemapWithError(t *testing.T) {
	// Teste básico para verificar se a função está remapeando corretamente os resultados
	t.Run("Sem erro", func(t *testing.T) {
		// Resultados esperados após o remapeamento
		expectedPaginated := &Paginated[int]{
			Results:     []int{2, 3, 4},
			Offset:      0,
			Limit:       3,
			TotalRecord: 3,
		}

		// Chamada da função a ser testada
		result, err := RemapWithError(&Paginated[int]{
			Results:     []int{1, 2, 3},
			Offset:      0,
			Limit:       3,
			TotalRecord: 3,
		}, func(input int) (int, error) {
			// Implementação de exemplo: simplesmente adiciona 1 ao input
			return input + 1, nil
		})
		if err != nil {
			t.Errorf("Erro inesperado: %v", err)
		}

		// Verifica se os resultados remapeados correspondem ao esperado
		if !reflect.DeepEqual(result, expectedPaginated) {
			t.Errorf("Incompatibilidade de resultado. Esperado: %v, obtido: %v", expectedPaginated, result)
		}
	})

	// Teste para verificar se a função retorna um erro se a função de mapeamento falhar
	t.Run("Com erro", func(t *testing.T) {
		// Chamada da função a ser testada
		_, err := RemapWithError(&Paginated[int]{
			Results:     []int{1, 2, 3},
			Offset:      0,
			Limit:       3,
			TotalRecord: 3,
		}, func(input int) (int, error) {
			return 0, errors.New("erro de exemplo")
		})
		if err == nil {
			t.Errorf("Erro esperado, mas obteve zero")
		}
	})
}
