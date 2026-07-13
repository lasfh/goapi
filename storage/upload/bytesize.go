package upload

import "fmt"

type ByteSize int64

const (
	_  ByteSize = 1 << (10 * iota)
	KB
	MB
	GB
	TB
)

// String retorna a representação textual do tamanho em bytes na unidade mais adequada.
//
// Retorna:
//   - (string): O tamanho formatado com a unidade correspondente (B, KB, MB, GB ou TB).
func (b ByteSize) String() string {
	switch {
	case b >= TB:
		return fmt.Sprintf("%.0f TB", float64(b)/float64(TB))
	case b >= GB:
		return fmt.Sprintf("%.0f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.0f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.0f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
