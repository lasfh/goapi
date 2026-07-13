package resperr

import (
	"net/http"

	"github.com/lasfh/goapi/logs/resperr"
)

var (
	ErrDecodingData           = resperr.New(http.StatusBadRequest, "Erro ao decodificar dados.")
	ErrEncodingData           = resperr.New(http.StatusBadRequest, "Erro ao codificar dados.")
	ErrInternalError          = resperr.New(http.StatusInternalServerError, "Erro interno desconhecido.")
	ErrResourceNotFound       = resperr.New(http.StatusNotFound, "O recurso não foi encontrado.")
	ErrArquivoTamanhoExcedido = resperr.Newf(http.StatusBadRequest, "Um ou mais arquivos excedem o limite de %s por arquivo.")
	ErrTamanhoTotalExcedido   = resperr.Newf(http.StatusBadRequest, "O tamanho total dos arquivos excede o limite de %s.")
)
