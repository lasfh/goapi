package example

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/lasfh/goapi/examples/api/internal/resperr"
	"github.com/lasfh/goapi/http/response"
)

type mockExampleRepo struct {
	FindFunc func(ctx context.Context) ([]string, error)
}

func (m mockExampleRepo) Find(ctx context.Context) ([]string, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx)
	}

	return []string{}, nil
}

func TestFind(t *testing.T) {
	t.Run("retorna a lista do repositório", func(t *testing.T) {
		want := []string{"a", "b", "c"}

		s := NewService(mockExampleRepo{
			FindFunc: func(ctx context.Context) ([]string, error) {
				return want, nil
			},
		})

		got, err := s.Find(context.Background())
		if err != nil {
			t.Fatalf("Find() erro inesperado = %v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Find() = %v, esperava %v", got, want)
		}
	})

	t.Run("erro do repositório vira erro interno", func(t *testing.T) {
		repoErr := errors.New("falha no banco")

		s := NewService(mockExampleRepo{
			FindFunc: func(ctx context.Context) ([]string, error) {
				return nil, repoErr
			},
		})

		got, err := s.Find(context.Background())
		if err == nil {
			t.Fatal("Find() erro = nil, esperava erro")
		}

		if got != nil {
			t.Errorf("Find() = %v, esperava nil", got)
		}

		if !errors.Is(err, resperr.ErrInternalError) {
			t.Errorf("Find() erro = %v, esperava resperr.ErrInternalError", err)
		}

		if !errors.Is(err, repoErr) {
			t.Errorf("Find() erro = %v, esperava conter o erro original do repositório", err)
		}

		var respErr response.ResponseError
		if !errors.As(err, &respErr) {
			t.Fatalf("Find() erro = %T, esperava response.ResponseError", err)
		}

		if respErr.StatusCode() != http.StatusInternalServerError {
			t.Errorf("StatusCode() = %d, esperava %d", respErr.StatusCode(), http.StatusInternalServerError)
		}
	})
}
