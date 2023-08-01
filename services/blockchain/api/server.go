package api

import (
	"github.com/cristianrb/fidelityblockchain/bc"
	"github.com/cristianrb/fidelityblockchain/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

var cache = make(map[string]*bc.Blockchain)

type Server struct {
	router  *gin.Engine
	address string
}

type TransactionRequest struct {
	SenderBlockchainAddress string  `json:"sender_blockchain_address"`
	SenderPublicKey         string  `json:"sender_public_key"`
	Signature               string  `json:"signature"`
	Product                 string  `json:"product"`
	Currency                string  `json:"currency"`
	Value                   float32 `json:"value"`
}

type GetAmountResponse struct {
	Recipient string  `json:"recipient"`
	Amount    float32 `json:"amount"`
}

func NewServer(address string) *Server {
	server := &Server{
		address: address,
	}
	server.setupRouter()

	return server
}

func (s *Server) GetBlockchain() *bc.Blockchain {
	blockchain, ok := cache["blockchain"]
	if !ok {
		blockchain = bc.NewBlockChain("TBD")
		cache["blockchain"] = blockchain
	}

	return blockchain
}

func (s *Server) setupRouter() {
	router := gin.Default()

	router.GET("/chain", s.getChain)
	router.GET("/amount", s.getAmount)
	router.POST("/transactions", s.createTransaction)

	s.GetBlockchain().Start()

	s.router = router
}

// Start runs the HTTP server on a specific address.
func (s *Server) Start() error {
	return s.router.Run(s.address)
}

func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}

func (s *Server) getChain(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, s.GetBlockchain())
}

func (s *Server) createTransaction(ctx *gin.Context) {
	blockchain := s.GetBlockchain()

	var req TransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	pubKey := utils.PublicKeyFromString(req.SenderPublicKey)
	signature := utils.SignatureFromString(req.Signature)
	var t *bc.Transaction
	if req.Currency != bc.FC_CURRENCY {
		t = bc.NewTransaction(bc.FIDELITY_BLOCKCHAIN_ADDRESS, req.SenderBlockchainAddress, req.Product, req.Currency, req.Value/10.0)
	} else {
		t = bc.NewTransaction(req.SenderBlockchainAddress, bc.FIDELITY_BLOCKCHAIN_ADDRESS, req.Product, req.Currency, req.Value)
	}
	err := blockchain.AddTransaction(t, pubKey, signature)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}

	ctx.Status(http.StatusCreated)
}

func (s *Server) getAmount(ctx *gin.Context) {
	blockchain := s.GetBlockchain()
	bcAddress := ctx.Query("blockchain_address")
	amount := blockchain.CalculateTotalAmount(bcAddress)

	ctx.JSON(http.StatusOK, GetAmountResponse{
		Recipient: bcAddress,
		Amount:    amount,
	})
}
