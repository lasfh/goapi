package logfile

import (
	"errors"
	"os"
	"path"
	"testing"
	"time"
)

func TestOpenNewFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		rootPath    string
		filename    string
		expectedErr error
	}{
		{
			name:        "Criar novo arquivo",
			rootPath:    tempDir,
			filename:    "testfile.log",
			expectedErr: nil,
		},
		{
			name:        "Caminho raiz inválido",
			rootPath:    "/invalid/path",
			filename:    "testfile.log",
			expectedErr: os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := openNewFile(tt.rootPath, tt.filename)

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("openNewFile() erro = %v, esperado %v", err, tt.expectedErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("openNewFile() retornou erro inesperado: %v", err)
			}

			defer file.Close()

			if _, err := os.Stat(path.Join(tt.rootPath, tt.filename)); err != nil {
				t.Errorf("arquivo não foi criado no caminho esperado: %v", err)
			}
		})
	}
}

func TestNewFileName(t *testing.T) {
	tests := []struct {
		name     string
		year     int
		month    time.Month
		day      int
		expected string
	}{
		{
			name:     "Data padrão",
			year:     2024,
			month:    time.October,
			day:      5,
			expected: "log-2024-10-05.log",
		},
		{
			name:     "Mês de um dígito",
			year:     2023,
			month:    time.January,
			day:      15,
			expected: "log-2023-01-15.log",
		},
		{
			name:     "Ano bissexto",
			year:     2020,
			month:    time.February,
			day:      29,
			expected: "log-2020-02-29.log",
		},
		{
			name:     "Fim do ano",
			year:     2024,
			month:    time.December,
			day:      31,
			expected: "log-2024-12-31.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFileName(tt.year, tt.month, tt.day)
			if result != tt.expected {
				t.Errorf("Esperado: %s, obtido: %s", tt.expected, result)
			}
		})
	}
}

func TestTimeUntilMidnight(t *testing.T) {
	tests := []struct {
		name         string
		date         time.Time
		expectedTime time.Duration
	}{
		{
			name:         "Meio-dia",
			date:         time.Date(2024, 10, 4, 12, 0, 0, 0, time.UTC),
			expectedTime: 12 * time.Hour,
		},
		{
			name:         "Um minuto antes da meia-noite",
			date:         time.Date(2024, 10, 4, 23, 59, 0, 0, time.UTC),
			expectedTime: 1 * time.Minute,
		},
		{
			name:         "Exatamente à meia-noite",
			date:         time.Date(2024, 10, 4, 0, 0, 0, 0, time.UTC),
			expectedTime: 24 * time.Hour,
		},
		{
			name:         "Alguns segundos antes da meia-noite",
			date:         time.Date(2024, 10, 4, 23, 59, 50, 0, time.UTC),
			expectedTime: 10 * time.Second,
		},
		{
			name:         "De manhã cedo",
			date:         time.Date(2024, 10, 4, 6, 0, 0, 0, time.UTC),
			expectedTime: 18 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := timeUntilMidnight(tt.date); result != tt.expectedTime {
				t.Errorf("timeUntilMidnight() = %v, esperado %v", result, tt.expectedTime)
			}
		})
	}
}

func TestNewFileWriter(t *testing.T) {
	t.Run("Diretório existente", func(t *testing.T) {
		tempDir := t.TempDir()

		writer, err := NewFileWriter(tempDir, NewFileName)
		if err != nil {
			t.Fatalf("NewFileWriter() retornou erro inesperado: %v", err)
		}

		defer writer.Close()

		expectedFile := path.Join(tempDir, NewFileName(time.Now().Date()))
		if _, err := os.Stat(expectedFile); err != nil {
			t.Errorf("arquivo de log não foi criado: %v", err)
		}

		if writer.removeEmptyFile {
			t.Error("removeEmptyFile deveria ser false por padrão")
		}
	})

	t.Run("Cria diretório inexistente", func(t *testing.T) {
		logsPath := path.Join(t.TempDir(), "logs", "app")

		writer, err := NewFileWriter(logsPath, NewFileName)
		if err != nil {
			t.Fatalf("NewFileWriter() retornou erro inesperado: %v", err)
		}

		defer writer.Close()

		info, err := os.Stat(logsPath)
		if err != nil || !info.IsDir() {
			t.Errorf("diretório de logs não foi criado: %v", err)
		}
	})

	t.Run("Erro ao criar diretório", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("executando como root; permissões de escrita não são aplicadas")
		}

		parent := t.TempDir()
		if err := os.Chmod(parent, 0555); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_ = os.Chmod(parent, 0755)
		})

		if _, err := NewFileWriter(path.Join(parent, "logs"), NewFileName); err == nil {
			t.Error("NewFileWriter() deveria retornar erro quando o diretório não pode ser criado")
		}
	})

	t.Run("Caminho é um arquivo, não um diretório", func(t *testing.T) {
		filePath := path.Join(t.TempDir(), "arquivo.txt")
		if err := os.WriteFile(filePath, []byte("conteúdo"), 0644); err != nil {
			t.Fatal(err)
		}

		if _, err := NewFileWriter(filePath, NewFileName); err == nil {
			t.Error("NewFileWriter() deveria retornar erro para caminho que não é diretório")
		}
	})

	t.Run("Erro ao abrir o arquivo de log", func(t *testing.T) {
		invalidName := func(int, time.Month, int) string {
			return path.Join("subdir-inexistente", "log.log")
		}

		if _, err := NewFileWriter(t.TempDir(), invalidName); err == nil {
			t.Error("NewFileWriter() deveria retornar erro quando o arquivo não pode ser criado")
		}
	})

	t.Run("Com removeEmptyFile habilitado", func(t *testing.T) {
		writer, err := NewFileWriter(t.TempDir(), NewFileName, true)
		if err != nil {
			t.Fatalf("NewFileWriter() retornou erro inesperado: %v", err)
		}

		defer writer.Close()

		if !writer.removeEmptyFile {
			t.Error("removeEmptyFile deveria ser true")
		}
	})
}

func TestWrite(t *testing.T) {
	t.Run("Grava no arquivo", func(t *testing.T) {
		tempDir := t.TempDir()

		writer, err := NewFileWriter(tempDir, NewFileName)
		if err != nil {
			t.Fatal(err)
		}

		defer writer.Close()

		content := []byte("linha de log\n")

		n, err := writer.Write(content)
		if err != nil {
			t.Fatalf("Write() retornou erro inesperado: %v", err)
		}

		if n != len(content) {
			t.Errorf("Write() = %d bytes, esperado %d", n, len(content))
		}

		data, err := os.ReadFile(path.Join(tempDir, NewFileName(time.Now().Date())))
		if err != nil {
			t.Fatal(err)
		}

		if string(data) != string(content) {
			t.Errorf("conteúdo do arquivo = %q, esperado %q", data, content)
		}
	})

	t.Run("Arquivo não inicializado", func(t *testing.T) {
		writer := &fileWriter{}

		if _, err := writer.Write([]byte("teste")); err == nil {
			t.Error("Write() deveria retornar erro quando o arquivo não está inicializado")
		}
	})
}

func TestOpenNewFileMethod(t *testing.T) {
	t.Run("Substitui o arquivo atual", func(t *testing.T) {
		tempDir := t.TempDir()

		writer, err := NewFileWriter(tempDir, NewFileName)
		if err != nil {
			t.Fatal(err)
		}

		defer writer.Close()

		previousFile := writer.file

		if err := writer.openNewFile(); err != nil {
			t.Fatalf("openNewFile() retornou erro inesperado: %v", err)
		}

		if writer.file == previousFile {
			t.Error("openNewFile() deveria substituir o arquivo atual")
		}

		// O arquivo anterior deve ter sido fechado.
		if _, err := previousFile.Write([]byte("x")); err == nil {
			t.Error("o arquivo anterior deveria estar fechado")
		}
	})

	t.Run("Remove arquivo vazio anterior", func(t *testing.T) {
		tempDir := t.TempDir()

		writer, err := NewFileWriter(tempDir, NewFileName, true)
		if err != nil {
			t.Fatal(err)
		}

		defer writer.Close()

		previousName := writer.file.Name()

		if err := writer.openNewFile(); err != nil {
			t.Fatalf("openNewFile() retornou erro inesperado: %v", err)
		}

		// openNewFile abre o mesmo nome (mesma data) e depois remove o arquivo
		// vazio anterior, portanto o arquivo não deve mais existir no disco.
		if _, err := os.Stat(previousName); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("arquivo vazio anterior deveria ter sido removido, mas os.Stat retornou: %v", err)
		}
	})

	t.Run("Mantém arquivo anterior com conteúdo", func(t *testing.T) {
		tempDir := t.TempDir()

		writer, err := NewFileWriter(tempDir, NewFileName, true)
		if err != nil {
			t.Fatal(err)
		}

		defer writer.Close()

		if _, err := writer.Write([]byte("dados\n")); err != nil {
			t.Fatal(err)
		}

		previousName := writer.file.Name()

		if err := writer.openNewFile(); err != nil {
			t.Fatalf("openNewFile() retornou erro inesperado: %v", err)
		}

		if _, err := os.Stat(previousName); err != nil {
			t.Errorf("arquivo com conteúdo não deveria ter sido removido: %v", err)
		}
	})

	t.Run("Erro ao abrir novo arquivo", func(t *testing.T) {
		tempDir := t.TempDir()

		writer, err := NewFileWriter(tempDir, NewFileName)
		if err != nil {
			t.Fatal(err)
		}

		defer writer.Close()

		writer.rootPath = path.Join(tempDir, "removido")

		if err := writer.openNewFile(); err == nil {
			t.Error("openNewFile() deveria retornar erro quando o diretório não existe")
		}
	})
}

func TestClose(t *testing.T) {
	t.Run("Fecha o arquivo", func(t *testing.T) {
		writer, err := NewFileWriter(t.TempDir(), NewFileName)
		if err != nil {
			t.Fatal(err)
		}

		if err := writer.Close(); err != nil {
			t.Errorf("Close() retornou erro inesperado: %v", err)
		}

		if _, err := writer.file.Write([]byte("x")); err == nil {
			t.Error("o arquivo deveria estar fechado após Close()")
		}
	})

	t.Run("Remove arquivo vazio ao fechar", func(t *testing.T) {
		writer, err := NewFileWriter(t.TempDir(), NewFileName, true)
		if err != nil {
			t.Fatal(err)
		}

		fileName := writer.file.Name()

		if err := writer.Close(); err != nil {
			t.Errorf("Close() retornou erro inesperado: %v", err)
		}

		if _, err := os.Stat(fileName); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("arquivo vazio deveria ter sido removido ao fechar, mas os.Stat retornou: %v", err)
		}
	})

	t.Run("Mantém arquivo com conteúdo ao fechar", func(t *testing.T) {
		writer, err := NewFileWriter(t.TempDir(), NewFileName, true)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := writer.Write([]byte("dados\n")); err != nil {
			t.Fatal(err)
		}

		fileName := writer.file.Name()

		if err := writer.Close(); err != nil {
			t.Errorf("Close() retornou erro inesperado: %v", err)
		}

		if _, err := os.Stat(fileName); err != nil {
			t.Errorf("arquivo com conteúdo não deveria ter sido removido: %v", err)
		}
	})

	t.Run("Sem arquivo inicializado", func(t *testing.T) {
		writer := &fileWriter{stopOpenNewFiles: make(chan struct{})}

		if err := writer.Close(); err != nil {
			t.Errorf("Close() retornou erro inesperado: %v", err)
		}
	})
}
