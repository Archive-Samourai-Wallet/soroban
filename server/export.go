package server

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/cretz/bine/torutil/ed25519"
)

func ExportHiddenServiceSecret(seed string) ([]byte, error) {
	if len(seed) == 0 {
		return nil, errors.New("Invalid Seed")
	}
	str, err := hex.DecodeString(seed)
	if err != nil {
		return nil, err
	}

	var result bytes.Buffer
	result.WriteString("== ed25519v1-secret: type0 ==")
	for result.Len() < 32 {
		result.WriteByte(0)
	}

	pair := ed25519.FromCryptoPrivateKey(str)
	priv := pair.PrivateKey()

	result.Write(priv)
	return result.Bytes(), nil
}
