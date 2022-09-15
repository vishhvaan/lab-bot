package functions

import (
	"encoding/binary"
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
