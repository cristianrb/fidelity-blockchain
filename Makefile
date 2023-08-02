blockchain_5000:
	cd services/blockchain/cmd && go run main.go -address="0.0.0.0" -port=5000

blockchain_5001:
	cd services/blockchain/cmd && go run main.go -address="0.0.0.0" -port=5001

blockchain_5002:
	cd services/blockchain/cmd && go run main.go -address="0.0.0.0" -port=5002

wallet_server_8080:
	cd services/wallet/cmd && go run main.go -address="0.0.0.0:8080" -gateway="0.0.0.0:5000"

wallet_server_8081:
	cd services/wallet/cmd && go run main.go -address="0.0.0.0:8081" -gateway="0.0.0.0:5001"