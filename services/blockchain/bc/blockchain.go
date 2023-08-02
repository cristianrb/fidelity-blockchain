package bc

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cristianrb/fidelityblockchain/models"
	"github.com/cristianrb/fidelityblockchain/network"
	"github.com/cristianrb/fidelityblockchain/utils"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	MINING_DIFFICULTY                 = 3
	MINING_SENDER                     = "THE BLOCKCHAIN"
	FIDELITY_BLOCKCHAIN_ADDRESS       = "THE FIDELITY BLOCKCHAIN ADDRESS"
	MINING_REWARD                     = 1.0
	MINING_PRODUCT                    = "MINING_PRODUCT"
	MINING_TIMER_SEC                  = 20
	FC_CURRENCY                       = "FC"
	BLOCKCHAIN_PORT_RANGE_START       = 5000
	BLOCKCHAIN_PORT_RANGE_END         = 5003
	NEIGHBOR_IP_RANGE_START           = 0
	NEIGHBOR_IP_RANGE_END             = 0
	BLOCKCHAIN_NEIGHBOR_SYNC_TIME_SEC = 20
)

type Blockchain struct {
	TransactionPool   []*Transaction `json:"transactionPool"`
	Chain             []*Block       `json:"chain"`
	BlockChainAddress string         `json:"blockChainAddress"`
	port              uint16
	mux               sync.Mutex

	neighbors    []string
	muxNeighbors sync.Mutex
}

func NewBlockChain(blockChainAddress string, port uint16) *Blockchain {
	genesis := &Block{}
	bc := new(Blockchain)
	bc.BlockChainAddress = blockChainAddress
	bc.port = port
	bc.CreateBlock(0, genesis.Hash())
	return bc
}

func (bc *Blockchain) Start() {
	bc.StartSyncNeighbors()
	bc.ResolveConflicts()
	bc.StartMining()
}

func (bc *Blockchain) SetNeighbors() {
	bc.neighbors = network.FindNeighbors(
		network.GetHost(), bc.port,
		NEIGHBOR_IP_RANGE_START, NEIGHBOR_IP_RANGE_END,
		BLOCKCHAIN_PORT_RANGE_START, BLOCKCHAIN_PORT_RANGE_END)
	log.Printf("%v", bc.neighbors)
}

func (bc *Blockchain) SyncNeighbors() {
	bc.muxNeighbors.Lock()
	defer bc.muxNeighbors.Unlock()
	bc.SetNeighbors()
}

func (bc *Blockchain) StartSyncNeighbors() {
	bc.SyncNeighbors()
	_ = time.AfterFunc(time.Second*BLOCKCHAIN_NEIGHBOR_SYNC_TIME_SEC, bc.StartSyncNeighbors)
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.TransactionPool)
	bc.Chain = append(bc.Chain, b)
	bc.TransactionPool = []*Transaction{}

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/transactions", n)
		client := &http.Client{}
		req, _ := http.NewRequest("DELETE", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}

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

func (bc *Blockchain) ClearTransactionPool() {
	bc.TransactionPool = bc.TransactionPool[:0]
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
	bc.mux.Lock()
	defer bc.mux.Unlock()

	t := NewTransaction(MINING_SENDER, bc.BlockChainAddress, MINING_PRODUCT, FC_CURRENCY, MINING_REWARD)
	bc.AddTransaction(t, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, previousHash)

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/consensus", n)
		client := &http.Client{}
		req, _ := http.NewRequest("PUT", endpoint, nil)
		client.Do(req)
	}

	return true
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
}

func (bc *Blockchain) CreateTransaction(t *Transaction, senderPublicKey *ecdsa.PublicKey, signature *utils.Signature) error {
	if t.Currency != FC_CURRENCY {
		t = NewTransaction(FIDELITY_BLOCKCHAIN_ADDRESS, t.Sender, t.Product, t.Currency, t.Value/10.0)
	}

	err := bc.AddTransaction(t, senderPublicKey, signature)
	if err == nil {
		for _, n := range bc.neighbors {
			publicKeyStr := fmt.Sprintf("%064x%064x", senderPublicKey.X.Bytes(),
				senderPublicKey.Y.Bytes())
			signatureStr := signature.String()
			bt := &models.TransactionRequest{
				t.Sender, publicKeyStr, signatureStr, t.Product, t.Currency, t.Value}
			m, _ := json.Marshal(bt)
			buf := bytes.NewBuffer(m)
			endpoint := fmt.Sprintf("http://%s/transactions", n)
			client := &http.Client{}
			req, _ := http.NewRequest("PUT", endpoint, buf)
			client.Do(req)
		}
	}

	return err
}

func (bc *Blockchain) AddTransaction(t *Transaction, senderPublicKey *ecdsa.PublicKey, signature *utils.Signature) error {
	if t.Sender == MINING_SENDER {
		bc.TransactionPool = append(bc.TransactionPool, t)
		return nil
	}

	if t.Sender != FIDELITY_BLOCKCHAIN_ADDRESS && bc.HasPendingTransaction(t.Sender) {
		return errors.New(fmt.Sprintf("pending transaction for %s", t.Sender))
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
				return errors.New(fmt.Sprintf("%s has not enough coins", t.Sender))
			}
		}
		bc.TransactionPool = append(bc.TransactionPool, t)
		return nil
	}

	return errors.New("unknown error")
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

func (bc *Blockchain) ResolveConflicts() bool {
	var longestChain []*Block = nil
	maxLength := len(bc.Chain)

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/chain", n)
		resp, _ := http.Get(endpoint)
		if resp.StatusCode == 200 {
			var bcResp Blockchain
			decoder := json.NewDecoder(resp.Body)
			_ = decoder.Decode(&bcResp)

			chain := bcResp.Chain
			validChain := bc.ValidChain(chain)
			if len(chain) > maxLength && validChain {
				maxLength = len(chain)
				longestChain = chain
			}
		}
	}

	if longestChain != nil {
		bc.Chain = longestChain
		return true
	}

	return false
}

func (bc *Blockchain) ValidChain(chain []*Block) bool {
	preBlock := chain[0]
	currentIndex := 1
	for currentIndex < len(chain) {
		b := chain[currentIndex]
		if b.PreviousHash != preBlock.Hash() {
			return false
		}

		if !bc.ValidProof(b.Nonce, b.PreviousHash, b.Transactions, MINING_DIFFICULTY) {
			return false
		}

		preBlock = b
		currentIndex += 1
	}

	return true
}
