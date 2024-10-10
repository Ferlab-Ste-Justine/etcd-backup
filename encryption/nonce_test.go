package encryption

import (
	"bytes"
    "encoding/binary"
	"testing"

    chacha "golang.org/x/crypto/chacha20poly1305"
)

func TestNewNonce(t *testing.T) {
    nonce1, nonce1Err := NewNonce()
	if nonce1Err != nil {
		t.Errorf("Error occured creating first nonce: %s", nonce1Err.Error())
		return
	}

	if len(nonce1) != chacha.NonceSizeX {
		t.Errorf("Expected first nonce to be %d bytes and it was %d bytes", chacha.NonceSizeX, len(nonce1))
		return
	}

    nonce2, nonce2Err := NewNonce()
	if nonce2Err != nil {
		t.Errorf("Error occured creating second nonce: %s", nonce2Err.Error())
		return
	}

	if len(nonce2) != chacha.NonceSizeX {
		t.Errorf("Expected second nonce to be %d bytes and it was %d bytes", chacha.NonceSizeX, len(nonce2))
		return
	}

	if bytes.Equal(nonce1, nonce2) {
		t.Errorf("Expected generated keys to be different and they were the same")
		return
	}
}

func TestNewNonceIncNew(t *testing.T) {
    nonce1, nonce1Err := NewNonceInc()
	if nonce1Err != nil {
		t.Errorf("Error occured creating first nonce: %s", nonce1Err.Error())
		return
	}

	if len(nonce1.Base) != (chacha.NonceSizeX - 8) {
		t.Errorf("Expected first nonce base to be %d bytes and it was %d bytes", (chacha.NonceSizeX - 8), len(nonce1.Base))
		return
	}

    nonce2, nonce2Err := NewNonceInc()
	if nonce2Err != nil {
		t.Errorf("Error occured creating second nonce: %s", nonce2Err.Error())
		return
	}

	if len(nonce2.Base) != (chacha.NonceSizeX - 8) {
		t.Errorf("Expected second nonce base to be %d bytes and it was %d bytes", (chacha.NonceSizeX - 8), len(nonce2.Base))
		return
	}

	if bytes.Equal(nonce1.Base[:8], nonce2.Base[:8]) {
		t.Errorf("Expected epoch part of nonces to be different. They weren't")
		return
	}

	if bytes.Equal(nonce1.Base[8:], nonce2.Base[8:]) {
		t.Errorf("Expected random part of nonces to be different. They weren't")
		return
	}
}

func TestNewNonceIncNext(t *testing.T) {
    nonce, nonceErr := NewNonceInc()
	if nonceErr != nil {
		t.Errorf("Error occured creating nonce: %s", nonceErr.Error())
		return
	}

    nonceNext1 := nonce.Next()
	if len(nonceNext1) != chacha.NonceSizeX {
		t.Errorf("Expected incremented nonce to %d bytes and it was %d bytes", chacha.NonceSizeX, len(nonceNext1))
		return
	}

	if !bytes.Equal(nonce.Base[:8], nonceNext1[:8]) {
		t.Errorf("Expected first 8 bytes of incremented nonce to be the epoch. It wasn't")
		return
	}

	if !bytes.Equal(nonce.Base[8:], nonceNext1[8:(chacha.NonceSizeX - 8)]) {
		t.Errorf("Expected next part of nonce to be the same random part. It wasn't")
		return
	}

    increment := binary.BigEndian.Uint64(nonceNext1[(chacha.NonceSizeX - 8):])
    if increment != 0 {
		t.Errorf("Expected the increment part to be 0. It was %d", increment)
		return
    }

    nonceNext2 := nonce.Next()
	if len(nonceNext2) != chacha.NonceSizeX {
		t.Errorf("Expected twice incremented nonce to %d bytes and it was %d bytes", chacha.NonceSizeX, len(nonceNext2))
		return
	}

	if !bytes.Equal(nonce.Base[:8], nonceNext2[:8]) {
		t.Errorf("Expected first 8 bytes of twice incremented nonce to be the epoch. It wasn't")
		return
	}

	if !bytes.Equal(nonce.Base[8:], nonceNext2[8:(chacha.NonceSizeX - 8)]) {
		t.Errorf("Expected next part of twice incremented nonce to be the same random part. It wasn't")
		return
	}

    increment = binary.BigEndian.Uint64(nonceNext2[(chacha.NonceSizeX - 8):])
    if increment != 1 {
		t.Errorf("Expected the increment after second increment part to be 1. It was %d", increment)
		return
    }
}