package bc

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type Block struct {
	Timestamp    int64          `json:"timestamp"`
	Nonce        int            `json:"nonce"`
	PreviousHash [32]byte       `json:"previousHash"`
	Transactions []*Transaction `json:"transactions"`
}

func NewBlock(nonce int, previousHash [32]byte, transactions []*Transaction) *Block {
	return &Block{
		Timestamp:    time.Now().UnixNano(),
		Nonce:        nonce,
		PreviousHash: previousHash,
		Transactions: transactions,
	}
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256(m)
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"nonce"`
		PreviousHash string         `json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.Timestamp,
		Nonce:        b.Nonce,
		PreviousHash: fmt.Sprintf("%x", b.PreviousHash),
		Transactions: b.Transactions,
	})
}
