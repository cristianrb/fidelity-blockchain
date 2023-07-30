package bc

import (
	"crypto/ecdsa"
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
	t := NewTransaction(MINING_SENDER, bc.BlockChainAddress, MINING_PRODUCT, FC_CURRENCY, MINING_REWARD)
	bc.AddTransaction(t, nil, nil)
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

func (bc *Blockchain) AddTransaction(t *Transaction, senderPublicKey *ecdsa.PublicKey, signature *utils.Signature) bool {
	if t.Sender == MINING_SENDER {
		bc.TransactionPool = append(bc.TransactionPool, t)
		return true
	}

	println(t.Sender)
	if t.Sender != FIDELITY_BLOCKCHAIN_ADDRESS && bc.HasPendingTransaction(t.Sender) {
		println("pending transaction")
		return false
	}

	oldValue := t.Value
	if t.Currency != FC_CURRENCY {
		oldValue = t.Value * 10
	}
	ttv := &utils.TransactionToVerify{
		Product:  t.Product,
		Currency: t.Currency,
		Value:    oldValue,
	}
	if utils.VerifySignature(senderPublicKey, signature, ttv) {
		if t.Currency == FC_CURRENCY {
			if bc.CalculateTotalAmount(t.Sender) < t.Value {
				return false
			}
		}
		bc.TransactionPool = append(bc.TransactionPool, t)
		return true
	}

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

func (bc *Blockchain) HasPendingTransaction(blockchainAddress string) bool {
	for _, t := range bc.TransactionPool {
		if t.Sender == blockchainAddress {
			return true
		}
	}

	return false
}
