package encryption

import (
	"bytes"
	chacha "golang.org/x/crypto/chacha20poly1305"
	"io"
)

func encryptWithNonce(plaintext []byte, encrKey []byte, nonce []byte) ([]byte, error) {
	aead, err := chacha.NewX(encrKey)
	if err != nil {
		return nil, err
	}

	ciphertext := []byte{}
	ciphertext = aead.Seal(ciphertext, nonce, plaintext, nil)

	return append(nonce, ciphertext...), nil
}

func EncryptBytes(plaintext []byte, encrKey []byte) ([]byte, error) {
	nonce, nonceErr := NewNonce()
	if nonceErr != nil {
		return nil, nonceErr
	}

	return encryptWithNonce(plaintext, encrKey, nonce)
}

func DecryptBytes(encryptedMsg []byte, decKey []byte) ([]byte, error) {
	aead, err := chacha.NewX(decKey)
	if err != nil {
		return nil, err
	}

	nonce, ciphertext := encryptedMsg[:aead.NonceSize()], encryptedMsg[aead.NonceSize():]

	return aead.Open(nil, nonce, ciphertext, nil)
}

type EncryptStream struct {
	ChunkSize    int64
	MasterKey    []byte
	CipherKey    []byte
	Nonce        NonceInc
	Source       io.Reader
	SourceErr    error
	SourceBuffer *bytes.Buffer
}

func NewEncryptStream(masterKey []byte, source io.Reader, chunkSize int64) (*EncryptStream, error) {
	nonce, nonceErr := NewNonceInc()
	if nonceErr != nil {
		return nil, nonceErr
	}

	cipherKey, cipherKeyErr := GenerateRandomKey()
	if cipherKeyErr != nil {
		return nil, cipherKeyErr
	}

	return &EncryptStream{
		MasterKey:    masterKey,
		ChunkSize:    chunkSize,
		Source:       source,
		Nonce:        nonce,
		CipherKey:    cipherKey,
		SourceErr:    nil,
		SourceBuffer: bytes.NewBuffer(make([]byte, 0)),
	}, nil
}

func (stream *EncryptStream) AddCiphertext() {
	if stream.SourceErr != nil {
		return
	}

	srcInput := make([]byte, stream.ChunkSize)
	n, nErr := stream.Source.Read(srcInput)
	if nErr != nil {
		stream.SourceErr = nErr
	}
	if n == 0 {
		return
	}

	ciphertext, encrErr := encryptWithNonce(srcInput[:n], stream.CipherKey, stream.Nonce.Next())
	if encrErr != nil {
		stream.SourceErr = encrErr
		return
	}

	_, wrErr := stream.SourceBuffer.Write(ciphertext)
	if wrErr != nil {
		stream.SourceErr = wrErr
	}
}

func (stream *EncryptStream) Read(p []byte) (n int, err error) {
	if stream.SourceBuffer.Len() >= len(p) {
		n, nErr := stream.SourceBuffer.Read(p)
		if nErr != nil && nErr != io.EOF {
			return n, nErr
		}

		return n, nil
	}

	for stream.SourceBuffer.Len() < len(p) && stream.SourceErr == nil {
		stream.AddCiphertext()
	}

	n, nErr := stream.SourceBuffer.Read(p)
	if nErr != nil && nErr != io.EOF {
		return n, nErr
	}

	return n, stream.SourceErr
}

func (stream *EncryptStream) GetEncryptedCipherKey() ([]byte, error) {
	return EncryptBytes(stream.CipherKey, stream.MasterKey)
}

type DecryptStream struct {
	ChunkSize    int64
	MasterKey    []byte
	CipherKey    []byte
	Source       io.Reader
	SourceErr    error
	SourceBuffer *bytes.Buffer
}

func NewDecryptStream(masterKey []byte, encrCipherKey []byte, source io.Reader, chunkSize int64) (*DecryptStream, error) {
	cipherKey, cipherKeyErr := DecryptBytes(encrCipherKey, masterKey)
	if cipherKeyErr != nil {
		return nil, cipherKeyErr
	}

	return &DecryptStream{
		MasterKey:    masterKey,
		ChunkSize:    chunkSize,
		Source:       source,
		CipherKey:    cipherKey,
		SourceErr:    nil,
		SourceBuffer: bytes.NewBuffer(make([]byte, 0)),
	}, nil
}

func (stream *DecryptStream) AddPlaintext() {
	if stream.SourceErr != nil {
		return
	}

	aead, err := chacha.NewX(stream.CipherKey)
	if err != nil {
		stream.SourceErr = err
		return
	}

	srcInput := make([]byte, int(stream.ChunkSize)+aead.NonceSize()+aead.Overhead())
	n, nErr := stream.Source.Read(srcInput)
	if nErr != nil {
		stream.SourceErr = nErr
	}
	if n == 0 {
		return
	}

	plaintext, plaintextErr := DecryptBytes(srcInput[:n], stream.CipherKey)
	if plaintextErr != nil {
		stream.SourceErr = plaintextErr
		return
	}

	_, wrErr := stream.SourceBuffer.Write(plaintext)
	if wrErr != nil {
		stream.SourceErr = wrErr
	}
}

func (stream *DecryptStream) Read(p []byte) (n int, err error) {
	if stream.SourceBuffer.Len() >= len(p) {
		n, nErr := stream.SourceBuffer.Read(p)
		if nErr != nil && nErr != io.EOF {
			return n, nErr
		}

		return n, nil
	}

	for stream.SourceBuffer.Len() < len(p) && stream.SourceErr == nil {
		stream.AddPlaintext()
	}

	n, nErr := stream.SourceBuffer.Read(p)
	if nErr != nil && nErr != io.EOF {
		return n, nErr
	}

	return n, stream.SourceErr
}
