package api

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"github.com/cristianrb/fidelityblockchainwallet/utils"
	"github.com/cristianrb/fidelityblockchainwallet/wallet"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_VerifySignature(t *testing.T) {
	w := wallet.NewWallet()
	transaction := &Transaction{
		Product:  "product",
		Currency: "EUR",
		Value:    10,
	}

	s := transaction.GenerateSignature(w.PrivateKey)
	isVerified := VerifySignature(w.PublicKey, s, transaction)

	require.Equal(t, isVerified, true)
}

func VerifySignature(senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	m, _ := json.Marshal(t)
	h := sha256.Sum256(m)
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}
