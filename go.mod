module github.com/ast-jean/audiophash

go 1.22

require (
	// Audio decoding and encoding
	github.com/go-audio/audio v1.0.0
	github.com/go-audio/wav v1.0.0
	github.com/go-audio/aiff v1.0.0

	// FFT operations
	gonum.org/v1/gonum v0.14.0

	// For tests
	github.com/stretchr/testify v1.9.0
)