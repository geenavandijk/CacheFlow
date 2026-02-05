package password

import (
	"errors"
	"unicode"

	"code.cacheflow.internal/util/secrets"
	"code.cacheflow.internal/util/ptr"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the password using bcrypt and peppers the result.
func HashPassword(password *string) (*string, *int64, error) {
	if password == nil || *password == "" {
		return nil, nil, errors.New("password cannot be empty")
	}

	// Validate the password
	valid := ValidatePassword(password)
	if !*valid {
		return nil, nil, errors.New("password does not meet the required criteria")
	}

	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	encryptionVersion := secrets.SymmetricKeyValueLatestVersion

	// Pepper the hashed password (encrypt it)
	pepperedPassword, err := EncryptWithPepper(hashedPassword, ptr.Int64(int64(encryptionVersion)))
	if err != nil {
		return nil, nil, err
	}

	return &pepperedPassword, ptr.Int64(int64(encryptionVersion)), nil
}

// VerifyPassword verifies a password against a peppered bcrypt hash.
func VerifyPassword(password *string, pepperedHash *string, version *int64) (bool, error) {
	if password == nil || pepperedHash == nil || *password == "" || *pepperedHash == "" {
		return false, errors.New("password and hashed password cannot be empty")
	}

	// Decrypt the peppered hash
	decryptedHash, err := DecryptWithPepper(*pepperedHash, version)
	if err != nil {
		return false, err
	}

	// Compare the password with the decrypted hash
	err = bcrypt.CompareHashAndPassword(decryptedHash, []byte(*password))
	if err != nil {
		// Passwords do not match
		return false, nil
	}

	// Passwords match
	return true, nil
}

// MARK: RotatePassword
func RotatePassword(hash *string, currentversion *int64, version *int64) (*string, error) {
	if hash == nil || *hash == "" {
		return nil, errors.New("password, salt, and hash cannot be empty")
	}

	// Decrypt the hash
	decryptedHash, err := DecryptWithPepper(*hash, currentversion)
	if err != nil {
		return nil, err
	}

	// Re-enrypt the hash with the new version
	pepperedPassword, err := EncryptWithPepper(decryptedHash, version)
	if err != nil {
		return nil, err
	}

	return &pepperedPassword, nil
}

// MARK: ValidatePassword
func ValidatePassword(password *string) *bool {
	// Define the validation criteria
	const minLen = 8
	const maxLen = 32

	// If the password is nil, return false
	if password == nil {
		result := false
		return &result
	}

	// Check the length of the password
	length := len(*password)
	if length < minLen || length > maxLen {
		result := false
		return &result
	}

	// Flags for required characters
	var hasUpper bool
	var hasDigit bool
	var hasSpecial bool

	// Iterate over each character in the password
	for _, char := range *password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Check if all conditions are met
	result := hasUpper && hasDigit && hasSpecial
	return &result
}