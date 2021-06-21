package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"hash"

	"github.com/dchest/cmac"
)

func NewAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm, nil
}

// nonceSize must return the nonce size of the AEAD returned by newAEAD
func NonceSize() int {
	return 12
}

// nonceSize must return the tag size of the AEAD returned by newAEAD
func TagSize() int {
	return 16
}

func InitMac(key []byte) (hash.Hash, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cmac.New(block)
}