package redis

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
)

func parseValue(value string) (uint64, string) {
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

func formatValue(value, nonce string) string {
	// add nonce prefix
	return fmt.Sprintf("%s_%s", nonce, value)
}

func hash(domain, prefix, value string) string {
	return fmt.Sprintf("%s:%x", prefix, sha256.Sum256([]byte(domain+value)))

}
func nonceHash(domain, nonce string) string {
	return hash(domain, "n", nonce)
}

func keyHash(domain, key string) string {
	return hash(domain, "k", key)
}
