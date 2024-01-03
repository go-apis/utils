package xcrypt

import (
	"encoding/base64"
	"encoding/json"
)

func DecryptJson[T any](c Crypt, encrypted string, out T) error {
	raw, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return err
	}

	decrypted, err := c.Decrypt(raw)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(decrypted, out); err != nil {
		return err
	}
	return err
}

func EncryptJson[T any](c Crypt, data T) (string, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	encrypted, err := c.Encrypt(raw)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}
