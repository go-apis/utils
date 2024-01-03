package xcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type Crypt interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(encrypted []byte) ([]byte, error)
}

type crypt struct {
	block cipher.Block
}

// Encrypt will AES-encrypt the given byte slice
func (c *crypt) Encrypt(data []byte) ([]byte, error) {
	aesgcm, err := cipher.NewGCM(c.block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aesgcm.Seal(nonce, nonce, data, nil), nil
}

// DecryptBytes will AES-decrypt the given byte slice
func (c *crypt) Decrypt(encrypted []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(c.block)
	if err != nil {
		return nil, err
	}

	// Detach nonce from incoming cipher
	nonce := encrypted[:gcm.NonceSize()]
	data := encrypted[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func NewCrypt(key []byte) (Crypt, error) {
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, fmt.Errorf("Invalid KEY to set for crypt; must be 16, 24, or 32 bytes (got %d)", keyLen)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &crypt{
		block: block,
	}, nil
}
