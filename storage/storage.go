package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
)

type Storage interface {
	// WithExtValidation retorna uma nova instância de Storage com validação de extensão configurada.
	//
	// Parâmetros:
	//   - exts (...string): Lista de extensões permitidas.
	//
	// Retorna:
	//   - Storage: Nova instância com a validação aplicada.
	WithExtValidation(exts ...string) Storage

	// Read abre um arquivo para leitura e retorna um io.ReadCloser.
	//
	// Parâmetros:
	//   - ctx (context.Context): O contexto da solicitação.
	//   - path (string): Caminho completo do arquivo.
	//
	// Retorna:
	//   - io.ReadCloser: Stream para leitura.
	//   - error: Erro se falhar ao abrir.
	Read(ctx context.Context, path string) (io.ReadCloser, error)

	// ReadAll lê todo o conteúdo de um arquivo e retorna os bytes.
	//
	// Parâmetros:
	//   - ctx (context.Context): O contexto da solicitação.
	//   - path (string): Caminho completo do arquivo.
	//
	// Retorna:
	//   - []byte: Conteúdo do arquivo.
	//   - error: Erro se falhar na leitura.
	ReadAll(ctx context.Context, path string) ([]byte, error)

	// Create grava o conteúdo do reader em um arquivo no caminho indicado.
	// Cria diretórios pai se necessário.
	//
	// Parâmetros:
	//   - ctx (context.Context): O contexto da solicitação.
	//   - path (string): Caminho completo do arquivo de destino.
	//   - r (io.Reader): Fonte dos bytes a serem gravados.
	//
	// Retorna:
	//   - error: Erro ocorrido durante a operação, ou nil se gravou com sucesso.
	Create(ctx context.Context, path string, r io.Reader) error

	// Remove remove um arquivo.
	//
	// Parâmetros:
	//   - ctx (context.Context): O contexto da solicitação.
	//   - path (string): Caminho completo do arquivo.
	//
	// Retorna:
	//   - error: Erro se falhar na exclusão.
	Remove(ctx context.Context, path string) error
}

type setup func(ctx context.Context, options StorageOptions) (Storage, error)

var (
	ErrDriverNotFound      = errors.New("driver não encontrado")
	ErrExtensionNotAllowed = errors.New("extensão não permitida")
	ErrInvalidOptions      = errors.New("opções inválidas")

	drivers = map[DriverName]setup{
		Local: newLocal,
		S3:    newS3,
	}
)

// NewStorage instancia um backend de armazenamento com base no driver informado.
//
// Parâmetros:
//   - ctx (context.Context): O contexto da solicitação.
//   - options (StorageOptions): Opções de configuração do backend.
//
// Retorna:
//   - Storage: Instância do backend configurado.
//   - error: `ErrDriverNotFound` se o driver não estiver registrado, ou erro de inicialização do driver.
func NewStorage(ctx context.Context, options StorageOptions) (Storage, error) {
	if options.DriverName == "" {
		return nil, fmt.Errorf("storage: %w: DriverName é obrigatório", ErrInvalidOptions)
	}

	fn, ok := drivers[options.DriverName]
	if !ok {
		return nil, fmt.Errorf("storage(%s): %w", options.DriverName, ErrDriverNotFound)
	}

	s, err := fn(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("storage(%s): %w", options.DriverName, err)
	}

	return s, nil
}

func newLocal(_ context.Context, opts StorageOptions) (Storage, error) {
	localOpts := opts.Local
	localOpts.ValidExts = slices.Concat(opts.ValidExts, opts.Local.ValidExts)

	return NewLocal(localOpts)
}

func newS3(ctx context.Context, opts StorageOptions) (Storage, error) {
	s3Opts := opts.S3
	s3Opts.ValidExts = slices.Concat(opts.ValidExts, opts.S3.ValidExts)

	return NewS3(ctx, s3Opts)
}
