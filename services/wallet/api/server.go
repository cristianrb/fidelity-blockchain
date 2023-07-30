package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cristianrb/fidelityblockchainwallet/utils"
	"github.com/cristianrb/fidelityblockchainwallet/wallet"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	router  *gin.Engine
	address string
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

func (s *Server) setupRouter() {
	router := gin.Default()

	router.POST("/", s.createAccount)
	router.POST("/transactions", s.createTransaction)
	router.GET("/amount", s.getAmount)

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
	if err != nil || resp.StatusCode < 200 || resp.StatusCode > 299 {
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusCreated)
}

func (s *Server) getAmount(ctx *gin.Context) {
	blockchainAddress := ctx.Query("blockchain_address")
	url := fmt.Sprintf("http://localhost:5000/amount?blockchain_address=%s", blockchainAddress)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		ctx.JSON(resp.StatusCode, resp.Body)
		return
	}

	var getAmountResponse GetAmountResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&getAmountResponse); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.JSON(resp.StatusCode, getAmountResponse)
}
