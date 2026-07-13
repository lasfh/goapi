package healthcheck

import (
	"net/http"

	"github.com/lasfh/goapi/http/response"
)

type handler struct{}

func NewHandler() *handler {
	return &handler{}
}

// Healthcheck retorna um handler HTTP para verificar a saúde da aplicação.
//
// @Summary Verifica a saúde da aplicação.
// @Description Retorna um status "Ok!" se a aplicação estiver rodando corretamente.
// @Tags Health
// @Produce plain
// @Success 200 {object} response.response
// @Router /health [get]
func (h *handler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	response.SendOK(w, "Ok!")
}
