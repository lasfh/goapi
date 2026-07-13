package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewLocal(t *testing.T) {
	original := LocalOptions{
		Options: Options{
			ValidExts: []string{".txt"},
		},
		DirPerm:  0o777,
		FilePerm: 0o644,
	}

	st, err := NewLocal(original)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	fs, ok := st.(*local)
	if !ok {
		t.Fatalf("o armazenamento retornado não implementa *local")
	}

	if got := fs.DirPerm; got != 0o777 {
		t.Errorf("DirPerm = %o, esperado %o", got, 0o777)
	}

	if got := fs.FilePerm; got != 0o644 {
		t.Errorf("FilePerm = %o, esperado %o", got, 0o644)
	}

	if got := fs.ValidExts; !reflect.DeepEqual(got, original.ValidExts) {
		t.Errorf("ValidExts = %v, esperado %v", got, original.ValidExts)
	}
}

func TestLocal_WithExtValidation(t *testing.T) {
	fs := &local{
		Options: Options{
			ValidExts: []string{".txt"},
		},
	}

	newStg := fs.WithExtValidation(".jpg", "png") // "png" sem ponto para testar normalização

	newFs, ok := newStg.(*local)
	if !ok {
		t.Fatalf("o armazenamento retornado não implementa *local")
	}

	// Verifica se a nova instância tem as extensões corretas e NORMALIZADAS
	expectedNew := []string{".jpg", ".png"}
	if !reflect.DeepEqual(newFs.ValidExts, expectedNew) {
		t.Errorf("Nova instância: ValidExts = %v, esperado %v", newFs.ValidExts, expectedNew)
	}

	// Verifica se a instância original permaneceu inalterada
	expectedOriginal := []string{".txt"}
	if !reflect.DeepEqual(fs.ValidExts, expectedOriginal) {
		t.Errorf("Instância original: ValidExts = %v, esperado %v", fs.ValidExts, expectedOriginal)
	}

	// Verifica se são instâncias diferentes (ponteiros diferentes)
	if fs == newFs {
		t.Error("WithExtValidation deve retornar uma nova instância, mas retornou a mesma")
	}
}

func TestLocal_Create(t *testing.T) {
	t.Run("o arquivo é criado com sucesso e os dados são gravados.", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "dir", "file.txt")

		fs := &local{
			DirPerm:  0o755,
			FilePerm: 0o644,
		}

		content := []byte("Olá, mundo!")
		reader := bytes.NewReader(content)

		if err := fs.Create(context.Background(), path, reader); err != nil {
			t.Fatalf("esperado nil, obtido %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("falha ao ler o arquivo criado: %v", err)
		}

		if string(data) != string(content) {
			t.Fatalf("conteúdo do arquivo diferente: obtido %q, esperado %q", data, content)
		}
	})

	t.Run("falha quando checkExt retorna erro", func(t *testing.T) {
		fs := &local{
			Options: Options{
				ValidExts: []string{".txt"},
			},
			DirPerm:  0o755,
			FilePerm: 0o644,
		}

		err := fs.Create(context.Background(), "file.bad", bytes.NewReader(nil))
		if err == nil {
			t.Fatal("erro esperado, obtido nil")
		}
	})

	t.Run("falha se não for possível criar o diretório.", func(t *testing.T) {
		tmp := t.TempDir()

		// cria diretório somente leitura
		dir := filepath.Join(tmp, "locked")

		os.WriteFile(dir, []byte("x"), 0o644)

		path := filepath.Join(dir, "dir", "file.txt")

		fs := &local{
			Options: Options{
				ValidExts: []string{".txt"},
			},
			DirPerm:  0o755,
			FilePerm: 0o644,
		}

		err := fs.Create(context.Background(), path, bytes.NewReader(nil))
		if err == nil {
			t.Fatalf("erro esperado, obtido nil")
		}
	})

	t.Run("remove o arquivo se a cópia falhar.", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "file.txt")

		fs := &local{
			DirPerm:  0o755,
			FilePerm: 0o644,
		}

		reader := io.NopCloser(errReader{})

		err := fs.Create(context.Background(), path, reader)
		if err == nil {
			t.Fatalf("erro esperado, obtido nil")
		}

		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("o arquivo deveria ter sido removido em caso de falha.")
		}
	})
}

type errReader struct{}

func (errReader) Read(p []byte) (int, error) { return 0, errors.New("copy error") }
func (errReader) Close() error               { return nil }
