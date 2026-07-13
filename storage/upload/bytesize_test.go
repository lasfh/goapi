package upload

import "testing"

func TestByteSizeString(t *testing.T) {
	tests := []struct {
		name  string
		input ByteSize
		want  string
	}{
		{"bytes", 512, "512 B"},
		{"kilobytes", 2 * KB, "2 KB"},
		{"megabytes", 5 * MB, "5 MB"},
		{"gigabytes", 3 * GB, "3 GB"},
		{"terabytes", 1 * TB, "1 TB"},
		{"zero", 0, "0 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.String()
			if got != tt.want {
				t.Errorf("ByteSize(%d).String() = %q, esperado %q", tt.input, got, tt.want)
			}
		})
	}
}
