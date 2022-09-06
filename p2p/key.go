package p2p

import (
	"crypto/rand"
	"errors"
	"io/ioutil"
	"os"

	"github.com/libp2p/go-libp2p-core/crypto"
)

func keyExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

func keySave(filename string, key crypto.PrivKey) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := key.Raw()
	if err != nil {
		return err
	}
	_, err = file.WriteString(crypto.ConfigEncodeKey(data))
	if err != nil {
		return err
	}
	return nil
}

func keyLoad(filename string) (crypto.PrivKey, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	data, err = crypto.ConfigDecodeKey(string(data))
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalSecp256k1PrivateKey(data)
}

func KeyLoadOrCreate(filename string) (priv crypto.PrivKey, err error) {
	if keyExists(filename) {
		priv, err = keyLoad(filename)
		return
	}

	priv, _, err = crypto.GenerateKeyPairWithReader(crypto.Secp256k1, 32, rand.Reader)
	if err != nil {
		return
	}

	err = keySave(filename, priv)
	if err != nil {
		return
	}

	return
}
