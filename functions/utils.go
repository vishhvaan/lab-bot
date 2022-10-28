package functions

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

func Contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func GetKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Inttobyteslice(v int, length int) []byte {
	b := make([]byte, length)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func SHA256Sum(input string, outputLength int) (output string) {
	// if outputLength == 0 then send the whole sum
	sum := sha256.Sum256([]byte(input))
	output = hex.EncodeToString(sum[:])
	if outputLength != 0 {
		output = firstN(output, outputLength)
	}

	return output
}

func firstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}
