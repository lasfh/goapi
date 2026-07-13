package upload

import (
	"errors"
	"mime/multipart"
	"testing"
)

func makeFileHeader(name string, size int64) *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: name,
		Size:     size,
	}
}

func TestValidateFiles(t *testing.T) {
	tests := []struct {
		name         string
		files        []*multipart.FileHeader
		maxFileSize  ByteSize
		maxTotalSize ByteSize
		wantErr      error
	}{
		{
			name:         "sem arquivos",
			files:        []*multipart.FileHeader{},
			maxFileSize:  5 * MB,
			maxTotalSize: 20 * MB,
			wantErr:      nil,
		},
		{
			name: "arquivo dentro do limite",
			files: []*multipart.FileHeader{
				makeFileHeader("doc.pdf", int64(2*MB)),
			},
			maxFileSize:  5 * MB,
			maxTotalSize: 20 * MB,
			wantErr:      nil,
		},
		{
			name: "arquivo excede limite individual",
			files: []*multipart.FileHeader{
				makeFileHeader("doc.pdf", int64(6*MB)),
			},
			maxFileSize:  5 * MB,
			maxTotalSize: 20 * MB,
			wantErr:      ErrArquivoTamanhoExcedido,
		},
		{
			name: "total excede limite",
			files: []*multipart.FileHeader{
				makeFileHeader("a.pdf", int64(5*MB)),
				makeFileHeader("b.pdf", int64(5*MB)),
				makeFileHeader("c.pdf", int64(5*MB)),
				makeFileHeader("d.pdf", int64(5*MB)),
				makeFileHeader("e.pdf", int64(5*MB)),
			},
			maxFileSize:  5 * MB,
			maxTotalSize: 20 * MB,
			wantErr:      ErrTamanhoTotalExcedido,
		},
		{
			name: "múltiplos arquivos dentro dos limites",
			files: []*multipart.FileHeader{
				makeFileHeader("a.pdf", int64(4*MB)),
				makeFileHeader("b.pdf", int64(4*MB)),
				makeFileHeader("c.pdf", int64(4*MB)),
			},
			maxFileSize:  5 * MB,
			maxTotalSize: 20 * MB,
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFiles(tt.files, tt.maxFileSize, tt.maxTotalSize)
			if tt.wantErr == nil && err != nil {
				t.Errorf("esperava nil, obteve: %v", err)
			}
			if tt.wantErr != nil && err == nil {
				t.Errorf("esperava erro %v, obteve nil", tt.wantErr)
			}
			if tt.wantErr != nil && err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("esperava erro %v, obteve: %v", tt.wantErr, err)
				}
			}
		})
	}
}
