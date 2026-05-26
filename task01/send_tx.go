package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/joho/godotenv"
)

func main() {
	// ================1.配置参数===================
	infuraAPIKey := "d364d4a09e0c4edbbaa0a83919b91786"
	rpcURL := "https://sepolia.infura.io/v3/" + infuraAPIKey
	fmt.Println("测试网地址：", rpcURL)

	// 私钥
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	privateKeyHex := os.Getenv("PRIVATE_KEY_Hex")
	// 去掉0x，一会儿统一处理
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	// 接收方地址
	toAddressMyself := "0xCfdfadF4f58aEED0d6Be5c7Da98f9d58f3AB683c"

	//转账金额,单位wei, 1ETH = 10^18 wei
	amountWei := big.NewInt(1000000000000000)

	// gas 上限,普通转账上限是 21000
	gasLimit := uint64(210000)

	// ================2.连接网络=================
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal("Error connecting to Ethereum client", err.Error())
	}
	fmt.Println("√ 已连接到 sepolia 测试网")

	// ===============3.获取链ID================
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatal("Error getting chain ID from Ethereum client", err.Error())
	}

	//================4.解析私钥================
	// 解码私钥，生成bytes类型私钥信息
	privateKeyHexBytes, err := hexutil.Decode("0x" + privateKeyHex)
	if err != nil {
		log.Fatal("Error decoding private key：", err.Error())
	}
	// 获取私钥对象，可以生成公钥
	privateKeyObject, err := crypto.ToECDSA(privateKeyHexBytes)
	if err != nil {
		log.Fatal("Error Generating public/private key pair fail", err.Error())
	}

	//==============5.获取发送方地址和Nonce=========
	// 公钥获取
	publicKey := privateKeyObject.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Println("发送方地址：", fromAddress)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal("Error getting nonce", err.Error())
	}
	fmt.Println("nonce", nonce)

	//==============6.获取 gas 价格参数================
	//baseFee被网络接收的最低费用
	baseFee, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal("Error getting gas price", err.Error())
	}
	//小费，给旷工的，激励费用
	priorityFee, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		log.Fatal("Error getting priority fee", err.Error())
	}
	// 是为了给交易足够的缓冲，确保能被打包
	//increment := new(big.Int).Mul(big.NewInt(2), big.NewInt(params.GWei))
	increment := getPriorityFee(baseFee)
	gasFeeCap := new(big.Int).Add(baseFee, increment)
	gasFeeCap.Add(gasFeeCap, priorityFee)

	fmt.Println("√ baseFee：", baseFee)
	fmt.Println("√ priorityFee：", priorityFee)
	fmt.Printf("√ Max Fee Per Gas: %s Gwei\n", weiToGwei(gasFeeCap, 2))

	//==================7.构造 EIP-1599 类型交易==================
	toAddr := common.HexToAddress(toAddressMyself)
	txData := types.DynamicFeeTx{
		ChainID:   chainId,
		Nonce:     nonce,
		GasTipCap: priorityFee,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &toAddr,
		Value:     amountWei,
		Data:      []byte{}, //普通转账，无附加数据
	}
	tx := types.NewTx(&txData)

	//====================8.签名交易=================================
	// 必须使用NewEIP155Signer,因为 sepolia 网络要求EIP-155 签名规范
	signer := types.NewLondonSigner(chainId)
	signedTx, err := types.SignTx(tx, signer, privateKeyObject)
	if err != nil {
		log.Fatal("Error signing tx ", err.Error())
	}
	fmt.Println("√ 交易签名完成")

	//====================9.发送交易===================================
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal("Error sending tx ", err.Error())
	}

	//====================10.输出交易哈希========================================
	fmt.Println("==================交易已发送==================")
	fmt.Printf("交易哈希：%s\n", signedTx.Hash().Hex())
	fmt.Printf("查看交易：https://sepolia.etherscan.io/tx/%s\n", signedTx.Hash().Hex())

	//转账金额 ETH 显示
	fmt.Printf("转账金额：%s ETH\n", weiToEth(amountWei, 4))
	fmt.Printf("接收方地址：%s\n", toAddr)

	/*
		测试网地址： https://sepolia.infura.io/v3/d364d4a09e0c4edbbaa0a83919b91786
		√ 已连接到 sepolia 测试网
		发送方地址： 0x0D628088f8f00285a6295bD66A4DC5569a327CE6
		nonce 0
		√ baseFee： 1113814400
		√ priorityFee： 1440000
		√ Max Fee Per Gas: 1.25 Gwei
		√ 交易签名完成
		==================交易已发送==================
		交易哈希：0x13b26549d41ce8fcc6fc1c7958dca382f6c2fe97ac92f82376c7adfde217a3f6
		查看交易：https://sepolia.ether.io/tx/0x13b26549d41ce8fcc6fc1c7958dca382f6c2fe97ac92f82376c7adfde217a3f6
		转账金额：0.0010 ETH
		接收方地址：0xCfdfadF4f58aEED0d6Be5c7Da98f9d58f3AB683c
	*/

}

func getPriorityFee(baseFee *big.Int) *big.Int {
	// 最小增量 2 Gwei
	minTip := big.NewInt(params.GWei * 2)
	// 百分比增量 12.5%
	percentTip := new(big.Int).Div(baseFee, big.NewInt(8))
	if percentTip.Cmp(minTip) < 0 {
		return percentTip
	}
	return baseFee
}

func weiToGwei(wei *big.Int, precision int) string {
	gwei := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetFloat64(params.GWei))
	return gwei.Text('f', precision)
}

func weiToEth(wei *big.Int, precision int) string {
	eth := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetFloat64(params.Ether))
	return eth.Text('f', precision)
}
