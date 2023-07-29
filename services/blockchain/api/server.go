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
		//log.Printf("private key %v\n", minersWallet.PrivateKeyStr())
		//log.Printf("public key %v\n", minersWallet.PublicKeyStr())
		//log.Printf("blockchain address %v\n", minersWallet.BlockchainAddress())
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
	blockchain := s.GetBlockchain()
	ctx.JSON(http.StatusOK, blockchain)
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
	isCreated := blockchain.AddTransaction(req.SenderBlockchainAddress, nil, req.Product, req.Currency, req.Value, pubKey, signature)

	if isCreated {
		ctx.Status(http.StatusCreated)
		return
	}

	println(isCreated)
	ctx.Status(http.StatusBadRequest)
}

func (s *Server) getAmount(ctx *gin.Context) {
	blockchain := s.GetBlockchain()
	bcAddress := ctx.Query("blockchain_address")
	amount := blockchain.CalculateTotalAmount(bcAddress)

	ctx.JSON(http.StatusOK, AmountResponse{
		Recipient: bcAddress,
		Amount:    amount,
	})
}
