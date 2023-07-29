package bc

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/cristianrb/fidelityblockchain/utils"
	"strings"
	"time"
)

const (
	MINING_DIFFICULTY           = 3
	MINING_SENDER               = "THE BLOCKCHAIN"
	FIDELITY_BLOCKCHAIN_ADDRESS = "THE FIDELITY BLOCKCHAIN ADDRESS"
	MINING_REWARD               = 1.0
	MINING_PRODUCT              = "MINING_PRODUCT"
	MINING_TIMER_SEC            = 20
	FC_CURRENCY                 = "FC"
)

type Blockchain struct {
	TransactionPool   []*Transaction `json:"transactionPool"`
	Chain             []*Block       `json:"chain"`
	BlockChainAddress string         `json:"blockChainAddress"`
}

func NewBlockChain(blockChainAddress string) *Blockchain {
	genesis := &Block{}
	bc := new(Blockchain)
	bc.BlockChainAddress = blockChainAddress
	bc.CreateBlock(0, genesis.Hash())
	return bc
}

func (bc *Blockchain) Start() {
	bc.StartMining()
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.TransactionPool)
	bc.Chain = append(bc.Chain, b)
	bc.TransactionPool = []*Transaction{}
	return b
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.Chain[len(bc.Chain)-1]
}

func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.TransactionPool {
		transactions = append(transactions, NewTransaction(t.Sender, t.Recipient, t.Product, t.Currency, t.Value))
	}

	return transactions
}

func (bc *Blockchain) VerifySignature(senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	t2 := struct {
		Product  string  `json:"product"`
		Currency string  `json:"currency"`
		Value    float32 `json:"value"`
	}{
		Product:  t.Product,
		Currency: t.Currency,
		Value:    t.Value,
	}
	m, err := json.Marshal(t2)
	if err != nil {
		println("marshal err")
		return false
	}
	println(fmt.Sprintf("%064x%064x", senderPublicKey.X.Bytes(),
		senderPublicKey.Y.Bytes()))

	println(fmt.Sprintf("%x%x", s.R, s.S))
	h := sha256.Sum256(m)
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}

func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transactions []*Transaction, difficulty int) bool {
	zeros := strings.Repeat("0", difficulty)
	guessBlock := Block{0, nonce, previousHash, transactions}
	guessHashStr := fmt.Sprintf("%x", guessBlock.Hash())
	return guessHashStr[:difficulty] == zeros
}

func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactionPool()
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULTY) {
		nonce += 1
	}

	return nonce
}

func (bc *Blockchain) Mining() bool {
	bc.AddTransaction(MINING_SENDER, &bc.BlockChainAddress, MINING_PRODUCT, FC_CURRENCY, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, previousHash)
	println("Mining")
	return true
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
}

func (bc *Blockchain) AddTransaction(sender string, recipient *string, product, currency string, value float32, senderPublicKey *ecdsa.PublicKey, signature *utils.Signature) bool {
	if sender == MINING_SENDER {
		t := NewTransaction(sender, *recipient, product, currency, value)
		bc.TransactionPool = append(bc.TransactionPool, t)
		return true
	}

	var t *Transaction
	if currency != FC_CURRENCY {
		t = NewTransaction(FIDELITY_BLOCKCHAIN_ADDRESS, sender, product, currency, value/10.0)
	} else {
		t = NewTransaction(sender, FIDELITY_BLOCKCHAIN_ADDRESS, product, currency, value)
	}

	if bc.VerifySignature(senderPublicKey, signature, t) {
		bc.TransactionPool = append(bc.TransactionPool, t)
		return true
	}

	println("not verified")

	return false
}

func (bc *Blockchain) CalculateTotalAmount(blockChainAddress string) float32 {
	var totalAmount float32 = 0.0
	for _, b := range bc.Chain {
		for _, t := range b.Transactions {
			if blockChainAddress == t.Recipient {
				totalAmount += t.Value
			}

			if blockChainAddress == t.Sender {
				totalAmount -= t.Value
			}
		}
	}
	return totalAmount
}
