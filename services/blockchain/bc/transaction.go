package bc

type Transaction struct {
	Sender    string  `json:"sender"`
	Recipient string  `json:"recipient"`
	Product   string  `json:"product"`
	Currency  string  `json:"currency"`
	Value     float32 `json:"value"`
}

func NewTransaction(sender, recipient, product, currency string, value float32) *Transaction {
	return &Transaction{
		Sender:    sender,
		Recipient: recipient,
		Product:   product,
		Currency:  currency,
		Value:     value,
	}
}
