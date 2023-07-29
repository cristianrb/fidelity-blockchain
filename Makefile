blockchain:
	cd services/blockchain/cmd && go run main.go -address="0.0.0.0:5000"

wallet_server:
	cd services/wallet/cmd && go run main.go -address="0.0.0.0:8080"