package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalOptions struct {
	Options

	DirPerm  os.FileMode
	FilePerm os.FileMode
}

type local struct {
	Options

	DirPerm  os.FileMode
	FilePerm os.FileMode
}

func NewLocal(opts LocalOptions) (Storage, error) {
	opts.ValidExts = normalizeExtensions(opts.ValidExts)

	if opts.DirPerm == 0 {
		opts.DirPerm = 0o755
	}

	if opts.FilePerm == 0 {
		opts.FilePerm = 0o600
	}

	return &local{
		Options:  opts.Options,
		DirPerm:  opts.DirPerm,
		FilePerm: opts.FilePerm,
	}, nil
}

func (l *local) WithExtValidation(exts ...string) Storage {
	c := *l
	c.Options.ValidExts = normalizeExtensions(exts)

	return &c
}

func (l *local) ReadAll(ctx context.Context, path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (l *local) Read(ctx context.Context, path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (l *local) Create(ctx context.Context, path string, r io.Reader) error {
	if err := l.checkExt(path); err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := os.MkdirAll(dir, l.DirPerm); err != nil {
			return fmt.Errorf("criar diretório %s: %w", dir, err)
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, l.FilePerm)
	if err != nil {
		return fmt.Errorf("criar arquivo: %w", err)
	}

	defer file.Close()

	if _, err := io.Copy(file, r); err != nil {
		_ = file.Close()
		_ = os.Remove(path)

		return fmt.Errorf("copiar dados para arquivo: %w", err)
	}

	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(path)

		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

func (l *local) Remove(ctx context.Context, path string) error {
	return os.Remove(path)
}
