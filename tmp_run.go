package main

import (
	"fmt"
	"os"

	"github.com/ast-jean/audiophash/cmd/audiophash"
	"github.com/ast-jean/audiophash/pkg/config"
)

func main() {
	data, err := os.ReadFile("test/fixtures/base/guitar_test.wav")
	if err != nil {
		panic(err)
	}

	cfg := config.DefaultConfig(44100)
	hash, err := audiophash.AudioPHashBytes(data, &cfg, "wav")
	if err != nil {
		panic(err)
	}
	fmt.Println("Hash:", hash)

	data2, err := os.ReadFile("test/fixtures/base/b.wav")
	if err != nil {
		panic(err)
	}

	cfg2 := config.DefaultConfig(44100)
	hash2, err := audiophash.AudioPHashBytes(data2, &cfg2, "wav")
	if err != nil {
		panic(err)
	}
	fmt.Println("Hash:", hash2)
}
