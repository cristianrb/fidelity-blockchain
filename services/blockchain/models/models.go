package models

type TransactionRequest struct {
	SenderBlockchainAddress string  `json:"sender_blockchain_address"`
	SenderPublicKey         string  `json:"sender_public_key"`
	Signature               string  `json:"signature"`
	Product                 string  `json:"product"`
	Currency                string  `json:"currency"`
	Value                   float32 `json:"value"`
}
