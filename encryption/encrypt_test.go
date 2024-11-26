package encryption

import (
	"bytes"
	"io"
	"math"
	"strings"
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

func TestEncryptStream(t *testing.T) {
	testInput := "This is some test input"

	masterKey, masterKeyErr := GenerateRandomKey()
	if masterKeyErr != nil {
		t.Errorf("Error generating master key: %s", masterKeyErr.Error())
		return
	}

	encryptStr, encryptStrErr := NewEncryptStream(
		masterKey,
		strings.NewReader(testInput),
		5,
	)
	if encryptStrErr != nil {
		t.Errorf("Error generating encryption stream: %s", encryptStrErr.Error())
		return
	}

	var cypherText bytes.Buffer
	n, cpyErr := io.Copy(&cypherText, encryptStr)
	if cpyErr != nil {
		t.Errorf("Error reading encryption stream: %s", cpyErr.Error())
		return
	}

	expectedLen := len(testInput) + int(math.Ceil(float64(len(testInput))/5.0)*(chacha.NonceSizeX+chacha.Overhead))

	if n != int64(expectedLen) {
		t.Errorf("Expected to read %d bytes and instead read %d", expectedLen, n)
		return
	}

	cypherKeyCypher, cypherKeyCypherErr := encryptStr.GetEncryptedCipherKey()
	if cypherKeyCypherErr != nil {
		t.Errorf("Error reading encrypted encryption key: %s", cypherKeyCypherErr.Error())
		return
	}

	cypherKey, cypherKeyErr := DecryptBytes(cypherKeyCypher, masterKey)
	if cypherKeyErr != nil {
		t.Errorf("Error decrypting encryption key: %s", cypherKeyErr.Error())
		return
	}

	part1, part1Err := DecryptBytes(cypherText.Bytes()[:45], cypherKey)
	if part1Err != nil {
		t.Errorf("Error decrypting part 1 of stream: %s", part1Err.Error())
		return
	}

	if !bytes.Equal(part1, []byte("This ")) {
		t.Errorf("Decrypted part 1 was not expected output: '%s'", part1)
		return
	}

	part2, part2Err := DecryptBytes(cypherText.Bytes()[45:90], cypherKey)
	if part2Err != nil {
		t.Errorf("Error decrypting part 2 of stream: %s", part2Err.Error())
		return
	}

	if !bytes.Equal(part2, []byte("is so")) {
		t.Errorf("Decrypted part 2 was not expected output: '%s'", part2)
		return
	}

	part3, part3Err := DecryptBytes(cypherText.Bytes()[90:135], cypherKey)
	if part3Err != nil {
		t.Errorf("Error decrypting part 3 of stream: %s", part3Err.Error())
		return
	}

	if !bytes.Equal(part3, []byte("me te")) {
		t.Errorf("Decrypted part 3 was not expected output: '%s'", part3)
		return
	}

	part4, part4Err := DecryptBytes(cypherText.Bytes()[135:180], cypherKey)
	if part4Err != nil {
		t.Errorf("Error decrypting part 4 of stream: %s", part4Err.Error())
		return
	}

	if !bytes.Equal(part4, []byte("st in")) {
		t.Errorf("Decrypted part 4 was not expected output: '%s'", part4)
		return
	}

	part5, part5Err := DecryptBytes(cypherText.Bytes()[180:223], cypherKey)
	if part5Err != nil {
		t.Errorf("Error decrypting part 5 of stream: %s", part5Err.Error())
		return
	}

	if !bytes.Equal(part5, []byte("put")) {
		t.Errorf("Decrypted part 5 was not expected output: '%s'", part5)
		return
	}

	nonce := NonceInc{
		Base:      cypherText.Bytes()[0:16],
		Increment: 0,
	}

	if !bytes.Equal(nonce.Next(), cypherText.Bytes()[0:24]) {
		t.Errorf("Part 1 nonce did not match expected value")
		return
	}

	if !bytes.Equal(nonce.Next(), cypherText.Bytes()[45:69]) {
		t.Errorf("Part 2 nonce did not match expected value")
		return
	}

	if !bytes.Equal(nonce.Next(), cypherText.Bytes()[90:114]) {
		t.Errorf("Part 3 nonce did not match expected value")
		return
	}

	if !bytes.Equal(nonce.Next(), cypherText.Bytes()[135:159]) {
		t.Errorf("Part 4 nonce did not match expected value")
		return
	}

	if !bytes.Equal(nonce.Next(), cypherText.Bytes()[180:204]) {
		t.Errorf("Part 5 nonce did not match expected value")
		return
	}
}

func TestDecryptStream(t *testing.T) {
	testInput := "This is some test input"

	masterKey, masterKeyErr := GenerateRandomKey()
	if masterKeyErr != nil {
		t.Errorf("Error generating master key: %s", masterKeyErr.Error())
		return
	}

	encryptStr, encryptStrErr := NewEncryptStream(
		masterKey,
		strings.NewReader(testInput),
		5,
	)
	if encryptStrErr != nil {
		t.Errorf("Error generating encryption stream: %s", encryptStrErr.Error())
		return
	}

	cypherKeyCypher, cypherKeyCypherErr := encryptStr.GetEncryptedCipherKey()
	if cypherKeyCypherErr != nil {
		t.Errorf("Error reading encrypted encryption key: %s", cypherKeyCypherErr.Error())
		return
	}

	decryptStr, decryptStrErr := NewDecryptStream(
		masterKey,
		cypherKeyCypher,
		encryptStr,
		5,
	)
	if decryptStrErr != nil {
		t.Errorf("Error generating decryption stream: %s", decryptStrErr.Error())
		return
	}

	var plainText bytes.Buffer
	n, cpyErr := io.Copy(&plainText, decryptStr)
	if cpyErr != nil {
		t.Errorf("Error reading decryption stream: %s", cpyErr.Error())
		return
	}

	if n != int64(len(testInput)) {
		t.Errorf("Expected to read %d bytes and instead read %d", len(testInput), n)
		return
	}

	if !bytes.Equal([]byte(testInput), plainText.Bytes()) {
		t.Errorf("Value from decrypted stream did not match original plaintext value")
		return
	}
}
