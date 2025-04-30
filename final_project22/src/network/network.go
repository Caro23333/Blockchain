package network

import (
	"fmt"
	"log"
	"net/rpc"
	"project2/local_bc"
	"project2/signature"
	"sync"
)

// 节点的信息
type NodeArgs struct {
	IP string
	Port string
	PublicKey signature.JSONPublicKey
}

// 广播交易的参数
type TransactionUpdateArgs struct {
	NewTransaction local_bc.Transaction
}

// 广播区块的参数
type BlockchainUpdateArgs struct {
	NewBlock local_bc.Block
}

// 广播测试消息的参数
type TestMessageArgs struct {
	M []byte	// 被签名的消息
	R []byte	 
	S []byte	// 数字签名
	OriginIP string
}

var connection_pool []*rpc.Client = make([]*rpc.Client, 0) // 管理了所有的连接。广播时只要遍历它就行了
var self_ip string // 自己的 ip

// 1.5修改：增加一个新的连接
func NewConnection(new_conn *rpc.Client) {
	connection_pool = append(connection_pool, new_conn)
}

// 广播一条交易
var NewTransactionDigestPool map[string]int = make(map[string]int)
var NewTransactionDigestMutex sync.Mutex
func NewTransactionDigestQuery(digest []byte) bool {
	NewTransactionDigestMutex.Lock()
	_, ok := NewTransactionDigestPool[string(digest)]
	NewTransactionDigestMutex.Unlock()
	return ok
}
func BroadcastTransaction(tx local_bc.Transaction) {
	NewTransactionDigestMutex.Lock()
	NewTransactionDigestPool[string(local_bc.Hash(tx))] = 1
	NewTransactionDigestMutex.Unlock()
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
var NewBlockDigestPool map[string]int = make(map[string]int)
var NewBlockDigestMutex sync.Mutex
func NewBlockDigestQuery(digest []byte) bool {
	NewBlockDigestMutex.Lock()
	_, ok := NewBlockDigestPool[string(digest)]
	NewBlockDigestMutex.Unlock()
	return ok
}
func BroadcastBlock(block local_bc.Block) {
	NewBlockDigestMutex.Lock()
	NewBlockDigestPool[string(local_bc.Hash(block))] = 1
	NewBlockDigestMutex.Unlock()
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

// 1.5修改：广播一个新节点
var NewNodeDigestPool map[string]int = make(map[string]int)
var NewNodeDigestMutex sync.Mutex
func NewNodeDigestQuery(digest []byte) bool {
	NewNodeDigestMutex.Lock()
	_, ok := NewNodeDigestPool[string(digest)]
	NewNodeDigestMutex.Unlock()
	return ok
}
func BroadcastNewNode(new_node_list []NodeArgs, new_digest []byte) {
	NewNodeDigestMutex.Lock()
	NewNodeDigestPool[string(new_digest)] = 1
	NewNodeDigestMutex.Unlock()
	for _, connection := range connection_pool {
		err := connection.Call("NetworkService.NewNodeList", new_node_list, nil)
		if err != nil {
			fmt.Printf("%v call failed while broadcasting new node: %v\n", self_ip, err)
			log.Fatal()
		}
	}
}

// 1.5修改：广播测试消息
var TestMessageDigestPool map[string]int = make(map[string]int)
var TestMessageDigestMutex sync.Mutex
func TestMessageDigestQuery(digest []byte) bool {
	_, ok := TestMessageDigestPool[string(digest)]
	return ok
}
func BroadcastTestMessage(test_message TestMessageArgs, new_digest []byte) {
	TestMessageDigestPool[string(new_digest)] = 1
	for _, connection := range connection_pool {
		err := connection.Call("NetworkService.TestMessage", test_message, nil)
		if err != nil {
			fmt.Printf("%v call failed while broadcasting test message: %v\n", self_ip, err)
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