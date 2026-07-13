package storage

import (
	"errors"
	"reflect"
	"testing"
)

func TestOptions_checkExt(t *testing.T) {
	extOptions := func(exts ...string) *Options {
		return &Options{ValidExts: exts}
	}

	tests := []struct {
		name    string
		opts    *Options
		path    string
		wantErr bool
		errIs   error // erro que deve estar "wrapping" na chamada
	}{
		{
			name:    "nenhuma extensão válida – tudo permitido",
			opts:    &Options{}, // ValidExts vazia
			path:    "any/file.txt",
			wantErr: false,
		},
		{
			name:    "extensão permitida em maiúsculas",
			opts:    extOptions(".jpg", ".png"),
			path:    "photos/IMG_001.JPG",
			wantErr: false,
		},
		{
			name:    "extensão não permitida – erro com wrapper",
			opts:    extOptions(".jpg", ".png"),
			path:    "docs/document.pdf",
			wantErr: true,
			errIs:   ErrExtensionNotAllowed,
		},
		{
			name:    "extensão não permitida – caminho sem extensão",
			opts:    extOptions(".txt"),
			path:    "folder/noextension",
			wantErr: true,
			errIs:   ErrExtensionNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.checkExt(tt.path)

			if tt.wantErr && err == nil {
				t.Fatalf("Esperava um erro, mas obtive nil.")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if tt.wantErr {
				// verifica se o erro contém o wrapper esperado
				if !errors.Is(err, tt.errIs) {
					t.Fatalf("O erro esperado para wrap: %v, mas foi obtido: %v.", tt.errIs, err)
				}
			}
		})
	}
}

func TestNormalizeExtensions(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Adiciona ponto quando falta",
			input:    []string{"jpg", "png"},
			expected: []string{".jpg", ".png"},
		},
		{
			name:     "Mantém ponto quando já existe",
			input:    []string{".txt", ".pdf"},
			expected: []string{".txt", ".pdf"},
		},
		{
			name:     "Misto",
			input:    []string{"jpg", ".PNG", "TXT"},
			expected: []string{".jpg", ".png", ".txt"},
		},
		{
			name:     "Duplicados",
			input:    []string{"jpg", ".jpg", ".JPG"},
			expected: []string{".jpg"},
		},
		{
			name:     "Lista vazia",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeExtensions(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("normalizeExtensions() = %v, esperado %v", got, tt.expected)
			}
		})
	}
}
