package compress

import (
	"github.com/klauspost/compress/zstd"
)

var (
	encoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	decoder, _ = zstd.NewReader(nil)
)

// Compress compresses data using Zstd.
func Compress(src []byte) []byte {
	return encoder.EncodeAll(src, make([]byte, 0, len(src)/2))
}

// Decompress decompresses Zstd data.
func Decompress(src []byte) ([]byte, error) {
	return decoder.DecodeAll(src, nil)
}
