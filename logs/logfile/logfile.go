package logfile

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"sync"
	"time"
)

type FormatFileNameFunc func(year int, month time.Month, day int) string

// NewFileName gera o nome de um arquivo de log com base na data fornecida.
//
// Parâmetros:
//   - year (int): O ano para o nome do arquivo.
//   - month (time.Month): O mês para o nome do arquivo, formatado com dois dígitos.
//   - day (int): O dia para o nome do arquivo.
//
// Retorna:
//   - (string): Uma string formatada representando o nome do arquivo.
func NewFileName(year int, month time.Month, day int) string {
	return fmt.Sprintf("log-%d-%02d-%02d.log", year, month, day)
}

type fileWriter struct {
	mu               sync.Mutex
	file             *os.File
	rootPath         string
	stopOpenNewFiles chan struct{}
	formatFileName   FormatFileNameFunc
	removeEmptyFile  bool
}

// NewFileWriter cria uma nova instância de fileWriter para gerenciar logs em arquivo.
//
// Parâmetros:
//   - logsPath (string): O caminho para a pasta de logs.
//   - formatFileName (FormatFileNameFunc): Função que define o formato do nome dos arquivos de log.
//   - removeEmptyFile (bool, opcional): Define se arquivos vazios devem ser removidos automaticamente.
//
// Retorna:
//   - (*fileWriter): Um ponteiro para a instância de fileWriter.
//   - (error): Um erro se ocorrer durante a criação do arquivo ou diretório.
func NewFileWriter(
	logsPath string,
	formatFileName FormatFileNameFunc,
	removeEmptyFile ...bool,
) (*fileWriter, error) {
	info, err := os.Stat(logsPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(logsPath, os.ModePerm); err != nil {
			return nil, fmt.Errorf("logfile: %w", err)
		}
	} else if !info.IsDir() {
		return nil, fmt.Errorf("logfile: o caminho informado não é um diretório")
	}

	file, err := openNewFile(
		logsPath,
		formatFileName(
			time.Now().Date(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("logfile: %w", err)
	}

	writer := &fileWriter{
		rootPath:         logsPath,
		file:             file,
		stopOpenNewFiles: make(chan struct{}),
		formatFileName:   formatFileName,
	}

	if len(removeEmptyFile) > 0 {
		writer.removeEmptyFile = removeEmptyFile[0]
	}

	writer.waitAndOpenNewFile()

	return writer, nil
}

// Write grava os dados de log no arquivo se o modo de gravação estiver habilitado.
//
// Parâmetros:
//   - p ([]byte): O conteúdo a ser gravado.
//
// Retorna:
//   - (int): O número de bytes gravados.
//   - (error): Um erro se ocorrer durante a gravação no arquivo.
func (w *fileWriter) Write(p []byte) (int, error) {
	return w.writeToFile(p)
}

// writeToFile grava diretamente os dados no arquivo.
//
// Parâmetros:
//   - p ([]byte): O conteúdo a ser gravado.
//
// Retorna:
//   - (int): O número de bytes gravados.
//   - (error): Um erro se ocorrer durante a gravação no arquivo.
func (w *fileWriter) writeToFile(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return 0, fmt.Errorf("arquivo não inicializado")
	}

	return w.file.Write(p)
}

// waitAndOpenNewFile monitora o tempo e abre um novo arquivo de log à meia-noite.
func (w *fileWriter) waitAndOpenNewFile() {
	go func() {
		for {
			currentDate := time.Now()

			select {
			case <-time.After(
				timeUntilMidnight(currentDate),
			):
				if err := w.openNewFile(); err != nil {
					slog.Error(
						"logfile: erro ao abrir novo arquivo",
						slog.String("error", err.Error()),
					)
				}
			case <-w.stopOpenNewFiles:
				return
			}
		}
	}()
}

// openNewFile abre um novo arquivo de log baseado na data atual.
//
// Retorna:
//   - (error): Um erro se ocorrer ao abrir o novo arquivo.
func (w *fileWriter) openNewFile() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	file, err := openNewFile(
		w.rootPath,
		w.formatFileName(
			time.Now().Date(),
		),
	)
	if err != nil {
		return err
	}

	if w.file != nil {
		if w.removeEmptyFile {
			w.checkAndRemoveEmptyFile()
		}

		_ = w.file.Close()
	}

	w.file = file

	return nil
}

// checkAndRemoveEmptyFile remove o arquivo atual se ele estiver vazio.
func (w *fileWriter) checkAndRemoveEmptyFile() {
	info, err := w.file.Stat()
	if err != nil || info.Size() > 0 {
		return
	}

	_ = os.Remove(w.file.Name())
}

// Close encerra a gravação de logs e fecha o arquivo.
//
// Retorna:
//   - (error): Um erro se ocorrer ao fechar o arquivo.
func (w *fileWriter) Close() error {
	close(w.stopOpenNewFiles) // Sinaliza para a goroutine parar

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		if w.removeEmptyFile {
			w.checkAndRemoveEmptyFile()
		}

		return w.file.Close()
	}

	return nil
}

// openNewFile abre ou cria um arquivo no caminho especificado e permite a gravação e a anexação de dados.
//
// Parâmetros:
//   - rootPath (string): O caminho raiz onde o arquivo será criado ou aberto.
//   - filename (string): O nome do arquivo a ser criado ou aberto.
//
// Retorna:
//   - (*os.File, error): Um ponteiro para o arquivo aberto e um erro, caso ocorra algum problema na operação.
func openNewFile(rootPath string, filename string) (*os.File, error) {
	filepath := path.Join(
		rootPath, filename,
	)

	return os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func timeUntilMidnight(date time.Time) time.Duration {
	nextMidnight := time.Date(
		date.Year(),
		date.Month(),
		date.Day()+1,
		0, 0, 0, 0,
		date.Location(),
	)

	// Calcula a duração até a próxima meia-noite
	return nextMidnight.Sub(date)
}
