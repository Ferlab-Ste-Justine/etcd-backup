package encryption

import (
	"bytes"
	"testing"
)

func TestGenerateRandomKey(t *testing.T) {
	key1, key1Err := GenerateRandomKey()
	if key1Err != nil {
		t.Errorf("Error occured creating first key: %s", key1Err.Error())
		return
	}

	if len(key1) != 32 {
		t.Errorf("Expected first key to be 32 bytes and it was %d bytes", len(key1))
		return
	}

	key2, key2Err := GenerateRandomKey()
	if key2Err != nil {
		t.Errorf("Error occured creating second key: %s", key2Err.Error())
		return
	}

	if len(key2) != 32 {
		t.Errorf("Expected first key to be 32 bytes and it was %d bytes", len(key1))
		return
	}

	if bytes.Equal(key1, key2) {
		t.Errorf("Expected generated keys to be different and they were the same")
		return
	}
}