package api

import (
	"fmt"
	"github.com/cristianrb/fidelityblockchain/bc"
	"github.com/cristianrb/fidelityblockchain/models"
	"github.com/cristianrb/fidelityblockchain/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

var cache = make(map[string]*bc.Blockchain)

type Server struct {
	router  *gin.Engine
	address string
	port    uint16
}

type GetAmountResponse struct {
	Recipient string  `json:"recipient"`
	Amount    float32 `json:"amount"`
}

func NewServer(address string, port uint16) *Server {
	server := &Server{
		address: address,
		port:    port,
	}
	server.setupRouter()

	return server
}

func (s *Server) GetBlockchain() *bc.Blockchain {
	blockchain, ok := cache["blockchain"]
	if !ok {
		blockchain = bc.NewBlockChain("TBD", s.port)
		cache["blockchain"] = blockchain
	}

	return blockchain
}

func (s *Server) setupRouter() {
	router := gin.Default()

	router.GET("/chain", s.getChain)
	router.GET("/amount", s.getAmount)
	router.POST("/transactions", s.createTransaction)
	router.PUT("/transactions", s.updateTransactions)
	router.DELETE("/transactions", s.deleteTransactions)
	router.PUT("/consensus", s.consensus)

	s.GetBlockchain().Start()

	s.router = router
}

// Start runs the HTTP server on a specific address.
func (s *Server) Start() error {
	return s.router.Run(fmt.Sprintf("%s:%d", s.address, s.port))
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

	var req models.TransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	pubKey := utils.PublicKeyFromString(req.SenderPublicKey)
	signature := utils.SignatureFromString(req.Signature)
	t := bc.NewTransaction(req.SenderBlockchainAddress, bc.FIDELITY_BLOCKCHAIN_ADDRESS, req.Product, req.Currency, req.Value)
	err := blockchain.CreateTransaction(t, pubKey, signature)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}

	ctx.Status(http.StatusCreated)
}

func (s *Server) updateTransactions(ctx *gin.Context) {
	blockchain := s.GetBlockchain()

	var req models.TransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	pubKey := utils.PublicKeyFromString(req.SenderPublicKey)
	signature := utils.SignatureFromString(req.Signature)
	t := bc.NewTransaction(req.SenderBlockchainAddress, bc.FIDELITY_BLOCKCHAIN_ADDRESS, req.Product, req.Currency, req.Value)
	err := blockchain.AddTransaction(t, pubKey, signature)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}

	ctx.Status(http.StatusCreated)
}

func (s *Server) deleteTransactions(ctx *gin.Context) {
	s.GetBlockchain().ClearTransactionPool()
	ctx.Status(http.StatusAccepted)
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

func (s *Server) consensus(ctx *gin.Context) {
	s.GetBlockchain().ResolveConflicts()
	ctx.Status(http.StatusAccepted)
}
