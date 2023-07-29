package api

import (
	"bytes"
	"encoding/json"
	"github.com/cristianrb/fidelityblockchainwallet/utils"
	"github.com/cristianrb/fidelityblockchainwallet/wallet"
	"github.com/gin-gonic/gin"
	"net/http"
)

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

func (s *Server) setupRouter() {
	router := gin.Default()

	router.POST("/", s.createAccount)
	router.POST("/transactions", s.createTransaction)

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

func (s *Server) createAccount(ctx *gin.Context) {
	wt := wallet.NewWallet()
	ctx.JSON(http.StatusCreated, ToWalletResponse(wt))
}

func (s *Server) createTransaction(ctx *gin.Context) {
	var req TransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	publicKey := utils.PublicKeyFromString(req.SenderPublicKey)
	privateKey := utils.PrivateKeyFromString(req.SenderPrivateKey, publicKey)
	t := Transaction{
		Product:  req.Product,
		Currency: req.Currency,
		Value:    req.Value,
	}
	signature := t.GenerateSignature(privateKey)

	bt := &TransactionSignedRequest{
		SenderBlockchainAddress: req.SenderBlockchainAddress,
		SenderPublicKey:         req.SenderPublicKey,
		Signature:               signature.String(),
		Product:                 req.Product,
		Currency:                req.Currency,
		Value:                   req.Value,
	}
	m, _ := json.Marshal(bt)
	buf := bytes.NewBuffer(m)

	resp, err := http.Post("http://localhost:5000/transactions", "application/json", buf)
	defer resp.Body.Close()
	if err != nil {
		ctx.JSON(resp.StatusCode, resp.Body)
		return
	}

	ctx.JSON(resp.StatusCode, resp.Body)
}
