package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

type TransactionToVerify struct {
	Product  string  `json:"product"`
	Currency string  `json:"currency"`
	Value    float32 `json:"value"`
}

func (s *Signature) String() string {
	return fmt.Sprintf("%064x%064x", s.R, s.S)
}

func String2BigIntTuple(s string) (big.Int, big.Int) {
	bx, _ := hex.DecodeString(s[:64])
	by, _ := hex.DecodeString(s[64:])

	var bix big.Int
	var biy big.Int

	_ = bix.SetBytes(bx)
	_ = biy.SetBytes(by)

	return bix, biy
}

func SignatureFromString(s string) *Signature {
	x, y := String2BigIntTuple(s)
	return &Signature{&x, &y}
}

func PublicKeyFromString(s string) *ecdsa.PublicKey {
	x, y := String2BigIntTuple(s)
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: &x, Y: &y}
}

func PrivateKeyFromString(s string, publicKey *ecdsa.PublicKey) *ecdsa.PrivateKey {
	b, _ := hex.DecodeString(s[:])
	var bi big.Int
	_ = bi.SetBytes(b)
	return &ecdsa.PrivateKey{PublicKey: *publicKey, D: &bi}
}

func VerifySignature(senderPublicKey *ecdsa.PublicKey, s *Signature, t *TransactionToVerify) bool {
	m, err := json.Marshal(t)
	if err != nil {
		return false
	}
	h := sha256.Sum256(m)
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}
