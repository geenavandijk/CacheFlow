package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"sync"

	"crypto/x509"
	"encoding/hex"

	"encoding/pem"
	"errors"
	"io"

	"code.cacheflow.internal/util/secrets"

	"os"

	crypt "github.com/jmhobbs/struct-crypt"

	"github.com/charmbracelet/log"
)

const symmetricKey = `0aa915235d82958197f1ebb1d916c75f`

// MARK: Encrypt Data
func EncryptData(data string, version int32, overrideKey string) (string, error) {

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller: true, // Report the file name and line number
		ReportTimestamp: true, // Report the timestamp
		TimeFormat: "2006-01-02 15:04:05", // Set the time format
		Prefix: "ENCRYPT", // Set the prefix
	})

	publicKeySecret, found := secrets.GetPublicKeyWithVersion(version)

	if !found {
		logger.Error("failed to get public key: %d", version)
		return "", errors.New("failed to get public key")
	}

	var publicKey *rsa.PublicKey
	var err error

	if overrideKey != "" {
		publicKey, err = parsePublicKey(overrideKey)
		if err != nil {
			// log.WithFields(Fields{"error": err}).Error("failed to parse override key")
			logger.Error("failed to parse override key: %v", err)
			return "", err
		}
	} else {
		publicKey, err = parsePublicKey(publicKeySecret)
		if err != nil {
			// log.WithFields(Fields{"error": err}).Error("failed to parse public key")
			logger.Error("failed to parse public key: %v", err)
			return "", err
		}
	}

	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(data))
	if err != nil {
		logger.Error("failed to encrypt data: %v", err)
		return "", err
	}

	return hex.EncodeToString(encryptedData), nil
}

// MARK: Decrypt Data
func DecryptData(data string, version int32) (string, error) {
	
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller: true, // Report the file name and line number
		ReportTimestamp: true, // Report the timestamp
		TimeFormat: "2006-01-02 15:04:05", // Set the time format
		Prefix: "DECRYPT", // Set the prefix
	})

	privateKeySecret, found := secrets.GetPrivateKeyWithVersion(version)

	if !found {
		logger.Error("failed to get private key with version: %d", version)
		return "", errors.New("failed to get private key with version: " + string(version))
	}

	privateKey, err := parsePrivateKey(privateKeySecret)

	if err != nil {
		logger.Error("failed to parse private key: %v", err)
		return "", err
	}

	cipherText, err := hex.DecodeString(data)
	if err != nil {
		logger.Error("failed to decode hex string: %v", err)
		return "", err
	}

	decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
	if err != nil {
		logger.Error("failed to decrypt data: %v", err)
		return "", err
	}

	return string(decryptedData), nil
}

var cipherPool = sync.Pool{
    New: func() interface{} {
        return &aesCipher{}
    },
}

type aesCipher struct {
    block cipher.Block
}

// MARK: Symmetric Encryption
func SymmetricEncryption(data string, version int32, operationType string) (string, error) {

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true, // Report the file name and line number
		ReportTimestamp: true, // Report the timestamp
		TimeFormat:      "2006-01-02 15:04:05", // Set the time format
		Prefix:          "SYMMETRIC", // Set the prefix
	})

	// Retrieve the symmetric key from the cache
	key, found := secrets.GetSymmetricKeyWithVersion(version)

	if !found {
		logger.Error("failed to get symmetric key")
		return "", errors.New("failed to get symmetric key")
	}

	// Decode the hexadecimal key
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		logger.Error("failed to decode hex key: %v", err)
		return "", err
	}

	// Ensure the key size is 128, 192, or 256 bits
	if len(keyBytes) != 16 && len(keyBytes) != 24 && len(keyBytes) != 32 {
		logger.Error("invalid key size")
		return "", errors.New("invalid key size")
	}

	// Perform encryption or decryption based on the operation type
	if operationType == "encrypt" {
		return symmetricEncrypt(data, key)
	} else if operationType == "decrypt" {
		return symmetricDecrypt(data, key)
	} else {
		logger.Error("invalid operation type")
		return "", errors.New("invalid operation type")
	}
}


// Symmetric encrypt function
func symmetricEncrypt(data, key string) (string, error) {
	
	// logger := log.NewWithOptions(os.Stderr, log.Options{
	// 	ReportCaller: true, // Report the file name and line number
	// 	ReportTimestamp: true, // Report the timestamp
	// 	TimeFormat: "2006-01-02 15:04:05", // Set the time format
	// 	Prefix: "SYMMETRIC", // Set the prefix
	// })

    poolCipher := cipherPool.Get().(*aesCipher)
    defer cipherPool.Put(poolCipher)

    // Initialize the cipher block if necessary
    if poolCipher.block == nil {
        block, err := aes.NewCipher([]byte(key))
        if err != nil {
            return "", err
        }
        poolCipher.block = block
    }

    cipherText := make([]byte, aes.BlockSize+len(data))
    iv := cipherText[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }

    stream := cipher.NewCFBEncrypter(poolCipher.block, iv)
    stream.XORKeyStream(cipherText[aes.BlockSize:], []byte(data))

    return hex.EncodeToString(cipherText), nil
}

// Symmetric decrypt function
func symmetricDecrypt(data, key string) (string, error) {
	
	// logger := log.NewWithOptions(os.Stderr, log.Options{
	// 	ReportCaller: true, // Report the file name and line number
	// 	ReportTimestamp: true, // Report the timestamp
	// 	TimeFormat: "2006-01-02 15:04:05", // Set the time format
	// 	Prefix: "SYMMETRIC", // Set the prefix
	// })

    poolCipher := cipherPool.Get().(*aesCipher)
    defer cipherPool.Put(poolCipher)

    if poolCipher.block == nil {
        block, err := aes.NewCipher([]byte(key))
        if err != nil {
            return "", err
        }
        poolCipher.block = block
    }

    cipherText, err := hex.DecodeString(data)
    if err != nil {
        return "", err
    }

    if len(cipherText) < aes.BlockSize {
        return "", errors.New("cipherText too short")
    }

    iv := cipherText[:aes.BlockSize]
    cipherText = cipherText[aes.BlockSize:]

    stream := cipher.NewCFBDecrypter(poolCipher.block, iv)
    stream.XORKeyStream(cipherText, cipherText)

    return string(cipherText), nil
}

// MARK: JWT Encryption
type JWTEncryptionData struct {
	Token   string
	ID      string
	Version *int32
}

// JWTEncryption function to encrypt/decrypt data based on the operation type
func JWTEncryption(data string, iv []byte, operationType string, version *int32) (JWTEncryptionData, error) {
	// key := secrets.SymmetricKeyValue
	// keyVersion := secrets.SymmetricKeyValueLatestVersion

	key, found := secrets.GetSymmetricKeyWithVersion(*version)
	keyVersion := version

	if !found {
		return JWTEncryptionData{}, errors.New("failed to get symmetric key")
	}

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller: true, // Report the file name and line number
		ReportTimestamp: true, // Report the timestamp
		TimeFormat: "2006-01-02 15:04:05", // Set the time format
		Prefix: "JWT", // Set the prefix
	})

	if operationType == "encrypt" {
		encryptedData, err := encryptJWT(data, key, iv)
		if err != nil {
			logger.Error("failed to encrypt data: %v", err)
			return JWTEncryptionData{}, err
		}

		return JWTEncryptionData{Token: encryptedData, ID: hex.EncodeToString(iv), Version: keyVersion}, nil
	} else if operationType == "decrypt" {
		decryptedData, err := decryptJWT(data, key, iv)
		if err != nil {
			logger.Error("failed to decrypt data: %v", err)
			return JWTEncryptionData{}, err
		}

		return JWTEncryptionData{Token: decryptedData, ID: hex.EncodeToString(iv), Version: keyVersion}, nil
	} else {
		logger.Error("invalid operation type: %s", operationType)
		return JWTEncryptionData{}, errors.New("invalid operation type")
	}
}

func encryptJWT(plainText, key string, iv []byte) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(iv) != aes.BlockSize {
		return "", errors.New("IV length must be 16 bytes")
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	copy(cipherText[:aes.BlockSize], iv)

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], []byte(plainText))

	return hex.EncodeToString(cipherText), nil
}

func decryptJWT(cipherHex, key string, iv []byte) (string, error) {
	cipherText, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(iv) != aes.BlockSize {
		return "", errors.New("IV length must be 16 bytes")
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("cipherText too short")
	}

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], cipherText[aes.BlockSize:])

	return string(cipherText[aes.BlockSize:]), nil
}

// MARK: Email Encryption
// EmailVerificationData holds the data for email verification.
type EmailVerificationData struct {
	ObjectID          string `encrypt:"EncryptedObjectID"`
	EncryptedObjectID string `decrypt:"ObjectID"`
	Email             string `encrypt:"EncryptedEmail"`
	EncryptedEmail    string `decrypt:"Email"`
	Code              string `encrypt:"EncryptedCode"`
	EncryptedCode     string `decrypt:"Code"`
}

var secret [32]byte

func EncryptEmailVerificationData(data *EmailVerificationData) error {
	copy(secret[:], []byte(symmetricKey))
	transform := crypt.New(secret)
	return transform.Encrypt(data)
}

func DecryptEmailVerificationData(data *EmailVerificationData) error {
	copy(secret[:], []byte(symmetricKey))
	transform := crypt.New(secret)
	return transform.Decrypt(data)
}

// MARK: Parse Public Key
func parsePublicKey(key string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("key type is not RSA")
	}
}

// MARK: Parse Private Key
func parsePrivateKey(key string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try parsing as PKCS1
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	switch priv := priv.(type) {
	case *rsa.PrivateKey:
		return priv, nil
	default:
		return nil, errors.New("key type is not RSA")
	}
}

// MARK: Generate Keypair
// GenerateKey generates either a symmetric or asymmetric key.
func GenerateKey(keyType string) (interface{}, error) {
	data := "Hello World"

	if keyType == "symmetric" {
		return generateSymmetricKey(data)
	} else if keyType == "asymmetric" {
		return generateAsymmetricKey(data)
	} else {
		return nil, errors.New("invalid key type")
	}
}

// generateSymmetricKey generates a symmetric key and tests it.
func generateSymmetricKey(data string) (string, error) {
	key := make([]byte, 16) // Generate a 256-bit key
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	keyHex := hex.EncodeToString(key)

	encrypted, err := symmetricEncrypt(data, keyHex)
	if err != nil {
		return "", errors.New("failed to encrypt data")
	}
	log.Info("Encrypted:", encrypted)

	decrypted, err := symmetricDecrypt(encrypted, keyHex)
	if err != nil {
		return "", errors.New("failed to decrypt data")
	}
	log.Info("Decrypted:", decrypted)

	return keyHex, nil
}



// generateAsymmetricKey generates an asymmetric key pair and tests it.
func generateAsymmetricKey(data string) (map[string]string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	publicKey := &privateKey.PublicKey

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(data))
	if err != nil {
		return nil, err
	}
	log.Info("Encrypted:", string(encrypted))

	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encrypted)
	if err != nil {
		return nil, err
	}
	log.Info("Decrypted:", string(decrypted))

	return map[string]string{
		"publicKey":  string(publicKeyPEM),
		"privateKey": string(privateKeyPEM),
	}, nil
}
