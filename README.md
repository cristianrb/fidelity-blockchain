# Fidelity blockchain

Every time a user buys something from the store they will be awarded 10% of the amount as crypto currency (FC = Fidelity Coin). Then the user can use their FC to buy anything from the store. The relation is always 1EUR = 0.1FC.

Normally we should have P2P but for learning purposes, we will be using the same PC with different ports (5000, 5001, 5002...). The code is easy to modify to use P2P.

## APIs

### Blockchain Server

- GET /amount (to show how many FC the user has)
- POST /transaction (where we need product, currency[EUR|FC]. If EUR is used, then the blockchain should award the user 10% of the amount as FC)

### Wallet Server

- POST /wallet (to create a new wallet -> will return private key, public key and blockchain address)
- GET /amount (to show how many FC the user has)
- POST /transaction (where we need product, currency used (EUR or FC))
