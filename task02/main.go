package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"sepolia-homework/task02/contract"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	// ================1.配置参数===================
	infuraAPIKey := "d364d4a09e0c4edbbaa0a83919b91786"
	rpcURL := "https://sepolia.infura.io/v3/" + infuraAPIKey
	fmt.Println("测试网地址：", rpcURL)

	//==================2.连接网络==================
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal("sepolia 测试网络连接错误：", err)
	}

	// =================3.加载私钥======================
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	privateKeyHex := os.Getenv("PRIVATE_KEY_Hex")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal("私钥解析失败：", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		//log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		log.Fatal("获取公钥失败")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal("获取nonce失败：", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal("获取gas价格失败：", err)
	}

	chainID, err := client.ChainID(context.Background())
	//创建交易授权对象
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal("创建授权对象失败：", err)
	}

	//设置交易序列号，防止重放攻击，保证交易按顺序处理
	auth.Nonce = big.NewInt(int64(nonce))
	//设置转账金额
	auth.Value = big.NewInt(0)
	// 设置gas上限
	auth.GasLimit = uint64(300000)
	// 设置gas 价格，即每单位gas支付的价格，单位为 Wei,从链上获取
	auth.GasPrice = gasPrice

	//初次部署合约
	//address, tx, instance, err := contract.DeployCounter(auth, client)
	//if err != nil {
	//	log.Fatal("部署失败：", err)
	//}
	//fmt.Printf("✓ 合约部署成功！\n")
	//fmt.Printf("合约地址: %s\n", address.Hex())
	//fmt.Printf("部署交易哈希: %s\n", tx.Hash().Hex())
	//fmt.Printf("查看交易: https://sepolia.etherscan.io/tx/%s\n", tx.Hash().Hex())
	//fmt.Printf("查看合约: https://sepolia.etherscan.io/address/%s\n", address.Hex())

	// 合约已被部署，走这里
	contractAddress := "0x348308d50fF3E3ED93933ab4b550ABff213EA861"
	contractAddressHex := common.HexToAddress(contractAddress)
	instance, err := contract.NewCounter(contractAddressHex, client)
	if err != nil {
		log.Fatal("合约实例化失败：", err)
	}

	//// ===============等待15s 让部署交易被确认=================
	//fmt.Println("等待部署交易被确认（15S）...")
	//time.Sleep(15 * time.Second)

	// =============读取初始计时（调用 view 函数）==============
	count, err := instance.GetCount(&bind.CallOpts{Context: context.Background()})
	if err != nil {
		log.Fatal("读取计数失败：", err)
	}
	fmt.Println("当前计数：", count)

	// ===============调用increment增加计数==================
	fmt.Println("正在调用 increment() 方法...")

	tx, err := instance.Increment(auth)
	if err != nil {
		log.Fatal("调用 increment 失败：", err)
	}
	fmt.Printf("✓ increment 交易已发送\n")
	fmt.Printf("交易哈希: %s\n", tx.Hash().Hex())
	fmt.Printf("查看: https://sepolia.etherscan.io/tx/%s\n", tx.Hash().Hex())

	// 可选：等待几秒后再读取新值
	fmt.Println("\n等待交易确认后，计数应该变为 1")
	// 实际使用中需要监听交易确认事件
	count, err = instance.GetCount(&bind.CallOpts{Context: context.Background()})
	if err != nil {
		log.Fatal("读取计数失败：", err)
	}
	fmt.Println("最新交易计数", count)

	/*
		第一次部署运行
			测试网地址： https://sepolia.infura.io/v3/d364d4a09e0c4edbbaa0a83919b91786
			等待部署交易被确认（15S）...
			当前计数： 0
			正在调用 increment() 方法...
			✓ increment 交易已发送
			交易哈希: 0x7c7d8ac650c2dc8d41110d0cb21efe6f043309c0a7cfe1572f2d49f86e11db22
			查看: https://sepolia.etherscan.io/tx/0x7c7d8ac650c2dc8d41110d0cb21efe6f043309c0a7cfe1572f2d49f86e11db22

			等待交易确认后，计数应该变为 1

		第二次运行
			测试网地址： https://sepolia.infura.io/v3/d364d4a09e0c4edbbaa0a83919b91786
			当前计数： 1
			正在调用 increment() 方法...
			✓ increment 交易已发送
			交易哈希: 0x7937d4f4b9b629f8657e86ab2c0e3033af3f71147a7bd9d18ef0d2d043e7347e
			查看: https://sepolia.etherscan.io/tx/0x7937d4f4b9b629f8657e86ab2c0e3033af3f71147a7bd9d18ef0d2d043e7347e

			等待交易确认后，计数应该变为 1
			最新交易计数 1
	*/

}
