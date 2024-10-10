package encryption

import(
	"crypto/rand"
	"encoding/binary"
	"errors"
	"math"
	"time"
	chacha "golang.org/x/crypto/chacha20poly1305"
)

type NonceInc struct {
	Base      []byte
	Increment int64
}

func NewNonceInc() (NonceInc, error) {
	epochAsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(epochAsBytes, uint64(time.Now().UnixNano()))

	randomBytes := make([]byte, chacha.NonceSizeX - 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return NonceInc{}, err
	}

	return NonceInc{Base: append(epochAsBytes, randomBytes...), Increment: 0}, nil
}

func (nonce *NonceInc) Next() []byte {
	if nonce.Increment == math.MaxInt64 {
		//Should not be reached in our practical use cases with a decent chunk size
		//But putting this awkward guard for now 
		panic(errors.New("Maximum nonce increment reached."))
	}

	incBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(incBytes, uint64(nonce.Increment))
	nonce.Increment += 1
	return append(nonce.Base, incBytes...)
}

func NewNonce() ([]byte, error) {
	epochAsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(epochAsBytes, uint64(time.Now().UnixNano()))

	randomBytes := make([]byte, chacha.NonceSizeX - 8)
	_, err := rand.Read(randomBytes)
	
	return append(epochAsBytes, randomBytes...), err
}