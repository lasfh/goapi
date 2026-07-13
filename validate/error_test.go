package validate

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/lasfh/goapi/logs/resperr"
)

func TestValidateError(t *testing.T) {
	innerErr := errors.New("erro interno")
	err := &validateError{
		ResponseError: resperr.New(
			http.StatusBadRequest,
			"Mensagem de erro",
		).WithDebugContext(context.TODO(), innerErr),
		details: map[string]string{
			"campo1": "detalhe 1",
			"campo2": "detalhe 2",
		},
	}

	if got := err.Error(); got != "erro interno" {
		t.Errorf("Error() = %q, esperado %q", got, "erro interno")
	}

	// if got := err.Unwrap(); !errors.Is(got, innerErr) {
	// 	t.Errorf("Unwrap() = %v, esperado %v", got, innerErr)
	// }

	if got := err.StatusCode(); got != http.StatusBadRequest {
		t.Errorf("StatusCode() = %d, esperado %d", got, http.StatusBadRequest)
	}

	if got := err.Message(); got != "Mensagem de erro" {
		t.Errorf("Message() = %q, esperado %q", got, "Mensagem de erro")
	}

	details := err.Details()
	if len(details) != 2 || details["campo1"] != "detalhe 1" || details["campo2"] != "detalhe 2" {
		t.Errorf("Details() = %v, esperado map com campo1 e campo2", details)
	}
}

func TestValidateErrorWithoutInnerError(t *testing.T) {
	err := &validateError{
		ResponseError: resperr.New(
			http.StatusBadRequest,
			"Mensagem de erro sem erro interno",
		),
	}

	if got := err.Error(); got != "Mensagem de erro sem erro interno" {
		t.Errorf("Error() = %q, esperado %q", got, "Mensagem de erro sem erro interno")
	}

	if got := err.Unwrap(); got != nil {
		t.Errorf("Unwrap() = %v, esperado nil", got)
	}

	if got := err.Details(); got != nil {
		t.Errorf("Details() = %v, esperado nil", got)
	}
}
