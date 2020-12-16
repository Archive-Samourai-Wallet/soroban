package common

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
)

func ParseValue(value string) (uint64, string) {
	toks := strings.SplitN(value, "_", 2)
	if len(toks) != 2 {
		return 0, value
	}
	// separate nonce prefix
	nonce, err := strconv.ParseUint(toks[0], 10, 64)
	if err != nil {
		return 0, value
	}

	return nonce, toks[1]
}

func FormatValue(counter int64, value string) string {
	// add counter prefix
	return fmt.Sprintf("%d_%s", counter, value)
}

func Hash(domain, prefix, value string) string {
	return fmt.Sprintf("%s:%x", prefix, sha256.Sum256([]byte(domain+value)))

}

func KeyHash(domain, key string) string {
	return Hash(domain, "k", key)
}

func CountHash(domain, count string) string {
	return Hash(domain, "c", count)
}

func ValueHash(domain, value string) string {
	return Hash(domain, "v", value)
}
