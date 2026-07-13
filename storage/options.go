package storage

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/lasfh/goapi/sliceutils"
)

type DriverName string

const (
	Local DriverName = "local"
	S3    DriverName = "s3"
)

// StorageOptions agrupa as opções de configuração dos backends de armazenamento suportados.
type StorageOptions struct {
	Options

	// DriverName Nome do driver a ser utilizado (ex.: Local, S3).
	DriverName DriverName

	// Local contém as opções de configuração para o backend de armazenamento local.
	Local LocalOptions

	// S3 contém as opções de configuração para o backend de armazenamento S3.
	S3 S3Options
}

type Options struct {
	// ValidExts lista de extensões de arquivo permitidas (ex.: ".jpg", ".png").
	ValidExts []string
}

// checkExt verifica se a extensão do arquivo está dentro do conjunto de extensões válidas.
//
// Parâmetros:
//   - path (string): Caminho completo ou nome do arquivo a ser verificado.
//
// Retorna:
//   - error: `ErrExtensionNotAllowed` se a extensão não estiver na lista
//     de extensões válidas; `nil` caso contrário.
func (s *Options) checkExt(path string) error {
	if len(s.ValidExts) == 0 {
		return nil
	}

	ext := strings.ToLower(
		filepath.Ext(path),
	)

	if !slices.Contains(s.ValidExts, ext) {
		return fmt.Errorf("%w: %s", ErrExtensionNotAllowed, ext)
	}

	return nil
}

func normalizeExtensions(exts []string) []string {
	normalized := make([]string, len(exts))
	for i, ext := range exts {
		if !strings.HasPrefix(ext, ".") {
			normalized[i] = "." + strings.ToLower(ext)

			continue
		}

		normalized[i] = strings.ToLower(ext)
	}

	return sliceutils.Unique(normalized)
}
