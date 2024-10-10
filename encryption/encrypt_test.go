package encryption

import (
	"bytes"
	"testing"

	chacha "golang.org/x/crypto/chacha20poly1305"
)

func TestEncryptDecryptBytes(t *testing.T) {
	cypherKey, cypherKeyErr := GenerateRandomKey()
	if cypherKeyErr != nil {
		t.Errorf("Error generating cypher key: %s", cypherKeyErr.Error())
		return
	}

	plaintext := []byte("I am home. tralala tralala tralala! So happy!")

	encr, encrErr := EncryptBytes(plaintext, cypherKey)
	if encrErr != nil {
		t.Errorf("Error encrypting plaintext: %s", encrErr.Error())
		return
	}

	aead, aeadErr := chacha.NewX(cypherKey)
	if aeadErr != nil {
		t.Errorf("Error constructing encryption structyre to get encryption overhead size: %s", aeadErr.Error())
		return
	}

	expectLen := len(plaintext) + chacha.NonceSizeX + aead.Overhead()
	if len(encr) != expectLen {
		t.Errorf("Expcted cyphertext to be of length %d. It was of length %d", expectLen, len(encr))
		return
	}

	if bytes.Contains(encr, []byte("I am home")) || bytes.Contains(encr, []byte("tralala")) || bytes.Contains(encr, []byte("So happy!")) {
		t.Errorf("Cyphertext contains appreciable parts of plaintext")
		return
	}

	decr, decErr := DecryptBytes(encr, cypherKey)
	if decErr != nil {
		t.Errorf("Error decrypting cyphertext: %s", encrErr.Error())
		return
	}

	if !bytes.Equal(decr, plaintext) {
		t.Errorf("Expected original plaintext to be equal to decrypted value. It wasn't")
		return
	}
}

