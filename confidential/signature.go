package confidential

import (
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/nacl/sign"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	log "github.com/sirupsen/logrus"
)

func toNaclPubKey(publicKey string) *[32]byte {
	var result [32]byte
	key, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil
	}
	copy(result[:], key[:32])
	return &result
}

// VerifySignature check signature with publicKey and message
// Support Nacl and Ecdsa Algorithms
func VerifySignature(info ConfidentialEntry, publicKey, message, algorithm, signature string) error {
	if len(info.Prefix) == 0 || len(info.Algorithm) == 0 || len(info.PublicKey) == 0 {
		return nil
	}
	log.WithField("Info", info).Debug("Verify Signature")
	switch info.Algorithm {
	case AlgorithmNacl:
		if info.Algorithm != algorithm {
			return errors.New("algorithm not maching")
		}
		if info.PublicKey != publicKey {
			return errors.New("publicKey not maching")
		}

		verified := verifyNaclSignature(publicKey, message, signature)
		if !verified {
			return errors.New("invalid singature")
		}

		log.Debug("Signature verified")
		return nil

	case AlgorithmEcdsa:
		if info.PublicKey != publicKey {
			return errors.New("publicKey not maching")
		}

		verified := verifyEcdsaSignature(publicKey, message, signature)
		if !verified {
			return errors.New("invalid singature")
		}

		log.Debug("Signature verified")
		return nil

	default:
		return errors.New("unknown signature algorithm")
	}
}

func signMessage(privateKey, message string) string {
	wif, err := btcutil.DecodeWIF(privateKey)
	if err != nil {
		panic(err)
	}

	privKey := wif.PrivKey
	pubKey := privKey.PubKey()
	log.Printf("%v %d\n", hex.EncodeToString(pubKey.SerializeCompressed()), len(pubKey.SerializeCompressed()))

	messageHash := chainhash.DoubleHashB([]byte(message))
	signature := ecdsa.Sign(privKey, messageHash)

	return hex.EncodeToString(signature.Serialize())
}

func verifyNaclSignature(publicKey, message, signature string) bool {
	signedMessage, _ := hex.DecodeString(signature)
	signedMessage = append(signedMessage, []byte(message)...)

	_, verified := sign.Open(nil, signedMessage, toNaclPubKey(publicKey))
	return verified
}

func verifyEcdsaSignature(publicKey, message, signature string) bool {
	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		panic(err)
	}
	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		panic(err)
	}

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		panic(err)
	}
	sign, err := ecdsa.ParseSignature(sigBytes)
	if err != nil {
		panic(err)
	}

	messageHash := chainhash.DoubleHashB([]byte(message))

	return sign.Verify(messageHash, pubKey)
}
