package example

import (
	"context"
	"net/http"

	"github.com/lasfh/goapi/http/response"
)

type ExampleService interface {
	Find(ctx context.Context) ([]string, error)
}

type handler struct {
	exampleService ExampleService
}

func NewHandler(
	exampleService ExampleService,
) *handler {
	return &handler{
		exampleService: exampleService,
	}
}

// Find retorna a lista de exemplos.
//
// @Summary Listar exemplos
// @Description Retorna a lista de exemplos cadastrados.
// @Tags Example
// @Produce json
// @Security Auth
// @Success 200 {array} string "Lista de exemplos"
// @Failure 500 {object} response.responseError "Erro interno"
// @Router /example [get]
func (h *handler) Find(w http.ResponseWriter, r *http.Request) {
	data, err := h.exampleService.Find(
		r.Context(),
	)
	if err != nil {
		response.SendError(w, err)

		return
	}

	response.EncodeOK(w, data)
}
