package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	infuraAPIKey := "d364d4a09e0c4edbbaa0a83919b91786"
	rpcURL := "https://sepolia.infura.io/v3/" + infuraAPIKey
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal(err)
	}

	account := "0x0D628088f8f00285a6295bD66A4DC5569a327CE6"
	address := common.HexToAddress(account)

	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatal(err)
	}

	ethBalance := new(big.Float).Quo(
		new(big.Float).SetInt(balance),
		new(big.Float).SetFloat64(math.Pow10(18)),
	)

	fmt.Printf("地址: %s\n", address.Hex())
	fmt.Printf("余额: %s ETH\n", ethBalance.Text('f', 6))
	fmt.Printf("余额(wei): %s\n", balance.String())
}
