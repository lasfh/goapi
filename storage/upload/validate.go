package upload

import (
	"errors"
	"fmt"
	"mime/multipart"
)

var (
	ErrArquivoTamanhoExcedido = errors.New("arquivo excede o limite de tamanho permitido")
	ErrTamanhoTotalExcedido   = errors.New("tamanho total dos arquivos excede o limite permitido")
)

// ValidateFiles valida uma lista de arquivos com base nos limites de tamanho individual e total.
//
// Parâmetros:
//   - files ([]*multipart.FileHeader): Lista de arquivos a serem validados.
//   - maxFileSize (ByteSize): Tamanho máximo permitido por arquivo.
//   - maxTotalSize (ByteSize): Tamanho máximo permitido para o conjunto de arquivos.
//
// Retorna:
//   - (error): Um erro se algum arquivo exceder o limite individual ou se o total ultrapassar o limite permitido.
func ValidateFiles(files []*multipart.FileHeader, maxFileSize, maxTotalSize ByteSize) error {
	var total ByteSize
	for _, f := range files {
		if ByteSize(f.Size) > maxFileSize {
			return fmt.Errorf("%w: %s (limite: %s)", ErrArquivoTamanhoExcedido, f.Filename, maxFileSize)
		}

		total += ByteSize(f.Size)
	}

	if total > maxTotalSize {
		return fmt.Errorf("%w: %s enviados, limite: %s", ErrTamanhoTotalExcedido, total, maxTotalSize)
	}

	return nil
}
