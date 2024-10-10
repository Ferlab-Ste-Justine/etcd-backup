package main

import (
	//"crypto/rand"
	//"encoding/binary"
	//"fmt"
	//"os"
	//"time"
	"github.com/Ferlab-Ste-Justine/etcd-backup/cmd"
)

/*const CHUNK_SIZE = 10 * 1024 * 1024

func readKey() ([]byte, error) {
    return os.ReadFile("../terraform/key")
}

type Nonce struct {
	Base      []byte
	Increment int64
}

func NewNonce() (Nonce, error) {
	epochAsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(epochAsBytes, uint64(time.Now().UnixNano()))

	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return Nonce{}, err
	}

	return Nonce{Base: append(epochAsBytes, randomBytes...), Increment: 0}, nil
}

func (nonce *Nonce) Next() []byte {
	incBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(incBytes, uint64(nonce.Increment))
	nonce.Increment += 1
	return append(nonce.Base, incBytes...)
}

type Nonce struct {
	Base      []byte
	Increment int64
}

type File struct {
	Path      string
	ChunkSize int64
}

func main() {
	_, keyErr := readKey()
	if keyErr != nil {
		fmt.Println(keyErr.Error())
		os.Exit(1);
	}

	nonce, nonceErr := NewNonce()
	if nonceErr != nil {
		fmt.Println(nonceErr.Error())
		os.Exit(1);
	}

	fmt.Println(nonce.Base)
	fmt.Println(nonce.Next())
	fmt.Println(nonce.Next())
	fmt.Println(nonce.Next())
}*/

func main() {
	cmd.Execute()
}
