package network

import (
	"fmt"
	"log"
	"net/rpc"
	"project2/local_bc"
	"project2/signature"
	"time"
)

// 节点的信息
type NodeArgs struct {
	IP string
	Port string
	PublicKey signature.JSONPublicKey
	KeyValid bool
}

// 广播交易的参数
type TransactionUpdateArgs struct {
	NewTransaction local_bc.Transaction
}

// 广播区块的参数
type BlockchainUpdateArgs struct {
	NewBlock local_bc.Block
}

var connection_pool []*rpc.Client = make([]*rpc.Client, 0) // 管理了所有的连接。广播时只要遍历它就行了
var self_ip string // 自己的 ip
var retry_limit int = 30 // 最多重试多少次

// 矿工向除了自己之外其他的所有矿工连接
// 由于程序启动顺序，先启动的节点会连接失败。此处重复若干次来保证能连到后来的节点
func InitConnection(node_list []NodeArgs, my_ip string) {
	self_ip = my_ip
	for _, dest := range node_list {
		if dest.IP != self_ip {
			retry_count := 0
			for {
				client, err := rpc.Dial("tcp", dest.IP + ":" + dest.Port)
				if err != nil {
					fmt.Printf("Miner %v: Error connecting %v: %v\n", self_ip, dest.IP, err)
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
}

// 广播一条交易
func BroadcastTransaction(tx local_bc.Transaction) {
	for _, connection := range connection_pool {
		args := TransactionUpdateArgs { tx }
		err := connection.Call("BlockchainService.NewTransaction", args, nil)
		if err != nil {
			fmt.Printf("%v call failed while broadcasting transaction: %v\n", self_ip, err)
			log.Fatal()
		}
	}
}

// 广播一个区块
func BroadcastBlock(block local_bc.Block) {
	for _, connection := range connection_pool {
		args := BlockchainUpdateArgs { 
			NewBlock: block, 
		}
		err := connection.Call("BlockchainService.NewBlock", args, nil)
		if err != nil {
			fmt.Printf("%v call failed while broadcasting block: %v\n", self_ip, err)
			log.Fatal()
		}
	}
}

// 广播自己的公钥
func BroadcastKey(public_key signature.JSONPublicKey, myself NodeArgs) {
	for _, connection := range connection_pool {
		args := NodeArgs { 
			IP: myself.IP,
			Port: myself.Port,
			PublicKey: public_key,
		}
		err := connection.Call("BlockchainService.NewNodeKey", args, nil)
		if err != nil {
			fmt.Printf("%v call failed while broadcasting key: %v\n", self_ip, err)
			log.Fatal()
		}
	}
}

// 别忘了关闭连接
func CloseConnection() {
	for _, connection := range connection_pool {
		connection.Close()
	}
}