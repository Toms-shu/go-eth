package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	//连接sepolia测试网
	infuraURL := "https://sepolia.infura.io/v3/d364d4a09e0c4edbbaa0a83919b91786"

	// 创建连接
	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Fatal("sepolia连接节点失败", err)
	}
	fmt.Println("√ sepolia测试网已连接")

	//获取最新区块号
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal("获取区块头失败：", err)
	}
	fmt.Printf("%+v\n", header)
	latestBlockNum := header.Number
	fmt.Println("最新区块号：", latestBlockNum.String())

	//查询指定区块
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal("获取区块失败：", err)
	}

	//输出区块信息
	fmt.Println("==============区块信息=============")
	fmt.Printf("区块哈希：%s\n", block.Hash().Hex())
	fmt.Printf("区块高度:%d\n", block.Number().Uint64())

	//时间戳转换
	timestamp := time.Unix(int64(block.Time()), 0)
	fmt.Printf("时间戳：%d(%s)\n)", block.Time(), timestamp.Format("2006-01-02 15:04:05"))

	fmt.Printf("交易数量：%d\n", len(block.Transactions()))
	fmt.Printf("挖矿难度：%s\n", block.Difficulty().String())

	txCount, err := client.TransactionCount(context.Background(), block.Hash())
	if err != nil {
		log.Fatal("获取交易数量失败：", err)
	}

	fmt.Printf("交易数量（确认）：%d\n", txCount)
}
