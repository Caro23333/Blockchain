package main

import (
	"fmt"
	"log"
	"net/rpc"
	"project2/network"
	"project2/signature"
	"time"
)

// 客户端这些方法都和服务端的意思差不多
// 目前只支持了这五台主机，可以改成所有的机器

/* var node_list []network.NodeArgs = []network.NodeArgs{
	{IP: "10.1.0.102", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.119", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.104", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.105", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.107", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
} */
var node_list []network.NodeArgs = []network.NodeArgs{
	{IP: "10.1.0.95", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.96", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.94", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.119", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.104", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.105", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.107", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.109", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.110", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.111", Port: "8057", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
}

var connection_pool []*rpc.Client = make([]*rpc.Client, 0)
var retry_limit int = 30

func InitConnection(node_list []network.NodeArgs) {
	for _, dest := range node_list {
		retry_count := 0
		for {
			client, err := rpc.Dial("tcp", dest.IP+":"+dest.Port)
			if err != nil {
				fmt.Printf("Client: Error connecting %v: %v\n", dest.IP, err)
				time.Sleep(2 * time.Second)
				retry_count += 1
				if retry_count >= retry_limit {
					fmt.Println("Connection retry timeout.")
					break
				}
			} else {
				connection_pool = append(connection_pool, client)
				break
			}
		}
	}
}

func CloseConnection() {
	for _, connection := range connection_pool {
		connection.Close()
	}
}

type ClientTransactionArgs struct {
	DestinationID int
	Value         float32
}

type ClientQueryArgs struct {
	Account int
}

type ClientQueryOneMinerArgs struct {
	Account int
}

type DifficultyArgs struct {
	NewLogDifficulty int
}

func main() {
	// 初始化 rpc
	InitConnection(node_list)
	defer CloseConnection()
	if len(connection_pool) != len(node_list) {
		log.Fatal("Connection not successful")
	}

	// 基于命令行界面的交互
	// 两种格式：
	//		query - 输出所有矿工的余额
	//		tx id1 id2 v - 要求编号为 id1 的矿工给编号为 id2 的矿工转帐 v 的价值
	for {
		var operation string
		var source, dest int
		var value float32
		_, input_err := fmt.Scanf("%s", &operation)
		if input_err != nil {
			fmt.Println("Invalid input format. Please retry.")
			continue
		}
		if operation == "query" {
			for index, connection := range connection_pool {
				var query_res float32 = 0.0
				connection.Call(
					"ClientService.ClientQuery",
					ClientQueryArgs{source},
					&query_res)
				fmt.Printf("Balance of node %d = %f\n", index, query_res)
			}
		} else if operation == "query_one"{
			_, input_err := fmt.Scanf("%d", &source)
			if input_err != nil {
				fmt.Println("Invalid input format. Please retry.")
				continue
			}
			if source < len(connection_pool) {
				for index, connection := range connection_pool {
					var query_res float32 = 0.0
					connection.Call(
						"ClientService.ClientQueryOneMiner",
						ClientQueryOneMinerArgs{source},
						&query_res)
					fmt.Printf("From node %d: Balance of node %d = %f\n", index, source, query_res)
				}
			} 
		} else if operation == "tx" {
			_, input_err := fmt.Scanf("%d %d %f", &source, &dest, &value)
			if input_err != nil {
				fmt.Println("Invalid input format. Please retry.")
				continue
			}
			if source < len(connection_pool) && dest < len(connection_pool) {
				var tx_flag float32 = 0
				connection_pool[source].Call(
					"ClientService.ClientTransaction",
					ClientTransactionArgs{
						DestinationID: dest,
						Value:         value,
					},
					&tx_flag)
				if tx_flag >= 0 {
					fmt.Printf("Transaction successfully submitted, spend confirmed balance of %v\n", tx_flag)
				} else if tx_flag == -1 {
					fmt.Println("No enough confirmed balance to make the transaction")
				} else if tx_flag == -2 {
					fmt.Println("Transaction destination is invalid")
				}
			}
		} else if operation == "adjust" {
			_, input_err := fmt.Scanf("%d", &dest)
			if input_err != nil {
				fmt.Println("Invalid input format. Please retry.")
				continue
			}
			for _, connection := range connection_pool {
				connection.Call(
					"ClientService.AdjustDifficulty",
					DifficultyArgs{dest},
					nil)
			}
		} else if operation == "exit" {
			break
		} else {
			fmt.Println("Invalid operation. Please retry.")
		}
	}
}

var ip_mapping = map[string]int{
	"10.1.0.91": 0,
	"10.1.0.92": 1,
	"10.1.0.94": 2,
	"10.1.0.95": 3,
	"10.1.0.96": 4,
	"10.1.0.98": 5,
	"10.1.0.99": 6,
	"10.1.0.102": 7,
	"10.1.0.119": 8,
	"10.1.0.104": 9,
	"10.1.0.105": 10,
	"10.1.0.107": 11,
	"10.1.0.109": 12,
	"10.1.0.110": 13,
	"10.1.0.111": 14,
	"10.1.0.112": 15,
	"10.1.0.113": 16,
	"10.1.0.115": 17,
	"10.1.0.116": 18,
	"10.1.0.117": 19,
}