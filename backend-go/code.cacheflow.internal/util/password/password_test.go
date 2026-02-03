package password

import (
	"testing"

	"code.cacheflow.internal/util/secrets"
	"code.cacheflow.internal/util/ptr"
	"github.com/stretchr/testify/assert"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		Name 			string
		Input 			string
		ExpectSuccess 	bool
	}{
		{
			Name: "Valid password",
			Input: "iloveMustafa123$",
			ExpectSuccess: true,
		},
		{
			Name: "Invalid password",
			Input: "f",
			ExpectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			valid := ValidatePassword(&tt.Input)

			if tt.ExpectSuccess {
				assert.True(t, *valid)
			} else {
				assert.False(t, *valid)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {

	secrets.SymmetricKey()

	tests := []struct{
		Name			string
		Password		string
		ExpectSuccess	bool
	}{
		{
			Name: "Valid password",
			Password: "iloveMustafa123$",
			ExpectSuccess: true,
		},
		{
			Name: "Invalid password",
			Password: "f",
			ExpectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			hash, _, err := HashPassword(&tt.Password)

			if tt.ExpectSuccess {

				assert.Nil(t, err)
				assert.NotEmpty(t, hash)
			} else {
				assert.NotNil(t, err)
				assert.Empty(t, hash)
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {

	secrets.SymmetricKey()

	tests := []struct{
		Name			string
		Password		string
		ExpectSuccess	bool
	}{
		{
			Name: "Valid password",
			Password: "iloveMustafa123$",
			ExpectSuccess: true,
		},
		{
			Name: "Invalid password",
			Password: "iloveSharks123$",
			ExpectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			hash, _, _ := HashPassword(ptr.String("iloveMustafa123$"))

			verify, err := VerifyPassword(&tt.Password, hash, ptr.Int64(2))

			if tt.ExpectSuccess {
				assert.Nil(t, err)
				assert.True(t, verify)
			} else {
				assert.False(t, verify)
			}
		})
	}
}

func TestRotatePassword(t *testing.T) {

	secrets.SymmetricKey()

	tests := []struct{
		Name			string
		Password		string
		Salt			string
		Hash			string
		CurrentVersion	int
		Version			int
		ExpectSuccess	bool
	}{
		{
			Name: "Valid password",
			Password: "iloveMustafa123$",
			CurrentVersion: 1,
			Version: 2,
			ExpectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			hash, _, _ := HashPassword(ptr.String("iloveMustafa123$"))

			newHash, err := RotatePassword(hash, ptr.Int64(int64(tt.CurrentVersion)), ptr.Int64(int64(tt.Version)))

			if tt.ExpectSuccess {
				assert.Nil(t, err)
				assert.NotEmpty(t, newHash)
			}
		})
	}
}