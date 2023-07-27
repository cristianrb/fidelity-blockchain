package api

type TransactionRequest struct {
	Sender   string  `json:"sender"`
	Product  string  `json:"product"`
	Currency string  `json:"currency"`
	Value    float32 `json:"value"`
}
