package password

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"code.cacheflow.internal/util/secrets"
)

// EncryptWithPepper encrypts the input data with the pepper key.
func EncryptWithPepper(data []byte, version *int64) (string, error) {
	symmetricKey, found := secrets.GetSymmetricKeyWithVersion(int32(*version))
	if !found {
		return "", errors.New("symmetric key not found")
	}
	block, err := aes.NewCipher([]byte(symmetricKey))
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], data)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptWithPepper decrypts the input data with the pepper key.
func DecryptWithPepper(encrypted string, version *int64) ([]byte, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		fmt.Println("Error decoding base64")
		return nil, err
	}

	// Get the symmetric key
	symmetricKey, found := secrets.GetSymmetricKeyWithVersion(int32(*version))
	if !found {
		fmt.Println("Symmetric key not found")
		return nil, errors.New("symmetric key not found")
	}

	block, err := aes.NewCipher([]byte(symmetricKey))
	if err != nil {
		fmt.Println("Error creating new cipher")
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		fmt.Println("Ciphertext too short")
		return nil, errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
