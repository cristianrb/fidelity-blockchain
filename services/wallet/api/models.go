package api

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"github.com/cristianrb/fidelityblockchainwallet/utils"
	"github.com/cristianrb/fidelityblockchainwallet/wallet"
)

type WalletResponse struct {
	PrivateKey        string `json:"private_key"`
	PublicKey         string `json:"public_key"`
	BlockchainAddress string `json:"blockchain_address"`
}

func ToWalletResponse(w *wallet.Wallet) WalletResponse {
	return WalletResponse{
		PrivateKey:        w.PrivateKeyStr(),
		PublicKey:         w.PublicKeyStr(),
		BlockchainAddress: w.BlockchainAddress,
	}
}

type TransactionRequest struct {
	SenderPrivateKey        string  `json:"sender_private_key"`
	SenderPublicKey         string  `json:"sender_public_key"`
	SenderBlockchainAddress string  `json:"sender_blockchain_address"`
	Product                 string  `json:"product"`
	Currency                string  `json:"currency"`
	Value                   float32 `json:"value"`
}

type TransactionSignedRequest struct {
	SenderBlockchainAddress string  `json:"sender_blockchain_address"`
	SenderPublicKey         string  `json:"sender_public_key"`
	Signature               string  `json:"signature"`
	Product                 string  `json:"product"`
	Currency                string  `json:"currency"`
	Value                   float32 `json:"value"`
}

type Transaction struct {
	Product  string  `json:"product"`
	Currency string  `json:"currency"`
	Value    float32 `json:"value"`
}

func (t *Transaction) GenerateSignature(privateKey *ecdsa.PrivateKey) *utils.Signature {
	m, _ := json.Marshal(t)
	h := sha256.Sum256(m)
	r, s, _ := ecdsa.Sign(rand.Reader, privateKey, h[:])
	return &utils.Signature{R: r, S: s}
}
