package hasher

import (
	"fmt"

	"golang.org/x/crypto/sha3"
)

func MultiHash(input []string) []string {
	result := make([]string, len(input))
	for i := 0; i < len(input); i++ {
		result[i] = Hash(input[i])
	}
	return result
}

func Hash(input string) string {
	h := sha3.New512()
	defer h.Reset()
	h.Write([]byte(input))
	return fmt.Sprintf("%x", h.Sum(nil))
}
