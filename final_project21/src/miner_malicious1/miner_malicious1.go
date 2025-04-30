package main

import (
	"crypto/ecdsa"
	// "crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"bytes"
	"net"
	"net/rpc"
	"project2/local_bc"
	"project2/network"
	"project2/signature"
	"sync"
	"time"
)

// 每个矿工自己的密钥对
var my_public_key signature.JSONPublicKey
var my_private_key *ecdsa.PrivateKey

var port string = "8058" // 矿工之间通信的端口
var my_ip string         // 自己的 ip
/* var node_list []network.NodeArgs = []network.NodeArgs{
	{IP: "10.1.0.102", Port: "8056", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.119", Port: "8056", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.104", Port: "8056", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.105", Port: "8056", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.107", Port: "8056", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
} */ // 所有矿工的列表。最开始不知道其他节点的公钥
var node_list []network.NodeArgs = []network.NodeArgs{
	{IP: "10.1.0.95", Port: "8058", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
	{IP: "10.1.0.96", Port: "8058", PublicKey: *new(signature.JSONPublicKey), KeyValid: false},
}

var queue_capacity int = 10
var task_queue_block chan local_bc.Block = make(chan local_bc.Block, queue_capacity)          // 所有收到的区块在此排队
var task_queue_tx chan local_bc.Transaction = make(chan local_bc.Transaction, queue_capacity) // 所有收到的交易在此排队
var chain_mutex sync.Mutex                                                                    // 当操作或者读取整个链相关的数据时需要 mutex
var tx_mutex sync.Mutex                                                                       // 当操作或者读取交易池相关的数据时需要 mutex
var diff_mutex sync.Mutex

// 矿工之间的 rpc 服务
type BlockchainService struct{}
type BlockchainUpdateArgs struct {
	NewBlock local_bc.Block
}

type TransactionUpdateArgs struct {
	NewTransaction local_bc.Transaction
}

// 处理收到的区块
func (t *BlockchainService) NewBlock(request_arg *BlockchainUpdateArgs, reply *int) error {
	task_queue_block <- request_arg.NewBlock
	return nil
}

// 处理收到的交易
func (t *BlockchainService) NewTransaction(request_arg *TransactionUpdateArgs, reply *int) error {
	task_queue_tx <- request_arg.NewTransaction
	return nil
}

// 处理收到的公钥
func (t *BlockchainService) NewNodeKey(request_arg *network.NodeArgs, reply *int) error {
	for index := range node_list {
		if node_list[index].IP == request_arg.IP {
			node_list[index].PublicKey = request_arg.PublicKey
			node_list[index].KeyValid = true
		}
	}
	return nil
}

// 这里开始有点混乱...... sorry for that 首先盘一下数据结构维护的逻辑：
// 		- 每个矿工会维护一个包含了所有区块的数据结构 - 区块树（这里存在 block_pool 中，这个池子包含了所有的区块）
// 		- 树上每个区块的父区块在收到时由其中的父区块哈希值确定（由 FatherBlockIndex 指定了父区块在池中的编号，这样池子就有了树形结构）
// 		- 每个矿工还会维护一个头区块（下标为 main_chain_head），对于矿工自己而言，从树根（创世区块）到头区块这条链被称为主链
// 		- 每个矿工都只认自己的主链。当它挖掘下一个区块时，下一个区块的父区块将是当前的头区块
// 		- 对于不在主链上的区块，矿工会视而不见。但其他矿工广播的区块以它们为父区块是允许的，矿工会如实对应地更新区块树
// 		- 当某一个新的区块并没有增长自己的主链，但是这个新的区块所积累的总难度已经超过当前的主链，则将头区块切换到新的区块（longest chain rule）

var main_UTXO_set map[string]local_bc.TransactionOutput = make(map[string]local_bc.TransactionOutput) // 假设只考虑当前主链的交易，有哪些未使用 Output (UTXO)
var main_UTXO_key_set map[string]local_bc.OutputKey = make(map[string]local_bc.OutputKey)             // 它们对应的 OutputKey
var tmp_used_UTXO_set map[string]int = make(map[string]int)                                           // 每一时刻都会有一些本地未提交的交易，它们消耗了在当前主链的 UTXO 中的某一部分，把这部分记下来。如果再来新的交易，这部分虽然在 main_UTXO_set 中，但不能再用了
var stale_message_set map[string]string = make(map[string]string)                                     // 本节点已经“见到过”的所有消息。对这些消息的签名将不再有效
var block_index map[string]int = make(map[string]int)                                                 // 将区块哈希值映射到区块池下标
var block_pool []local_bc.Block = make([]local_bc.Block, 0)                                           // 区块池
var is_in_main_chain []bool = make([]bool, 0)                                                         // 标记一个区块是否在主链中
var main_chain_head int                                                                               // 头区块的下标

// 其次是关于交易的逻辑：
// 		- 矿工会不断地收到一些交易，这些交易会被储存在一个交易池中
//		- 每当矿工自己完成工作量证明的时候，矿工会将这些交易打包到一个区块并从交易池移除它们
//		- 每当收到其他矿工广播的区块，也从交易池中移除区块中已经包含的交易
//		- 由于调度顺序不确定，可能会出现其他矿工先广播某区块，随后本矿工才收到该区块中的一个交易
// 		- transaction_flag 记录了所有已经由于区块创建被删掉的或将要被删掉的交易。通过这样的记录我们可以避免后来的交易再进入交易池

var transaction_pool map[string]local_bc.Transaction = make(map[string]local_bc.Transaction) // 交易池
var transaction_flag map[string]int = make(map[string]int)

// 工作量证明和挖矿奖励的参数
var overall_log_difficulty int = 21
var overall_difficulty int = 1 << 21
var creation_reward float32 = 10.0

// 在区块树上找到 new_block_father 和当前区块头的最近公共祖先
// 这等价于找到 new_block_father 最近的包含在主链中的祖先
func findLeastCommonAncestor(new_block_father int) (int, []int) {
	index := new_block_father
	ancestor_list := make([]int, 0)
	for index != -1 && !is_in_main_chain[index] {
		index = block_pool[index].FatherBlockIndex // 通过 FatherBlockIndex 字段我们可以沿着树不断往上跳
		if index != -1 {
			ancestor_list = append(ancestor_list, index)
		}
	}
	return index, ancestor_list
}

// 处理新的区块
func onNewBlockAppearance(new_block local_bc.Block) {
	fmt.Println("New block appears")
	chain_mutex.Lock()
	time.Sleep(2 * time.Microsecond)
	defer chain_mutex.Unlock()

	// 父区块哈希是指定新区块在树中位置的唯一标准
	// 验证当前区块树是否有哈希值符合的区块
	new_block_father_hash := new_block.Head.FatherBlockHash
	father_index, ok := block_index[string(new_block_father_hash)]
	if !ok {
		fmt.Printf("Father block invalid, required hash = %v\n", hex.EncodeToString(new_block_father_hash))
		//log.Fatal()
		return // Father block invalid
	}
	new_block.FatherBlockIndex = father_index // 如果有，那就确定了新区块在区块树中的父亲

	// 区块交易的验证是基于 UTXO 集合的。但新区块不一定是接在主链后面的，也就是说验证新区块时所用的 UTXO 集合可能和实际的主链 UTXO 集合不同
	// 我们需要先计算出临时的 UTXO 集合用于验证
	LCA, ancestor_list := findLeastCommonAncestor(father_index) // LCA 即为新区块对应的链与主链分叉的地方
	// fmt.Println("---------- DEBUG: LCA = ", LCA)
	tmp_index := main_chain_head
	tmp_UTXO_set := main_UTXO_set
	tmp_UTXO_key_set := main_UTXO_key_set
	// 首先在临时集合上撤消掉所有从 LCA 之后的交易
	for tmp_index != LCA {
		current_block := block_pool[tmp_index]
		local_bc.RevertBlock(&current_block, tmp_UTXO_set, tmp_UTXO_key_set, block_pool)
		tmp_index = current_block.FatherBlockIndex
	}
	// 再从 LCA 开始依次应用一系列区块，直到新区块之父
	for i := len(ancestor_list) - 1; i >= 0; i-- {
		current_block := block_pool[ancestor_list[i]]
		local_bc.ApplyBlock(&current_block, tmp_UTXO_set, tmp_UTXO_key_set)
	}
	// 此时临时 UTXO 集合就是新区块交易所面临的 UTXO 集合，可以用它来验证新区块了

	// 12.6 修改
	// 首先验证新区块的交易是否合法
	hash_bytes := local_bc.Hash(new_block.Head)
	if !local_bc.ValidatePoW(new_block.LogDifficulty, hash_bytes) {
		return // PoW 无效
	}
	tx_mutex.Lock()
	if res_flag, _ := new_block.ValidateTransaction(tmp_UTXO_set, stale_message_set, creation_reward); !res_flag {
		/*for _, tx := range invalid_txs {
			key := string(local_bc.Hash(tx))
			// 删除区块池中本地未提交的交易时，对应“用掉的 UTXO 集合”也需要更新
			for _, input := range tx.Inputs {
				output_key := local_bc.Hash(local_bc.OutputKey{TransactionID: input.TransactionID, Index: input.OutputID})
				delete(tmp_used_UTXO_set, string(output_key))
			}
			transaction_flag[key] = 1
			delete(transaction_pool, string(local_bc.Hash(tx)))
		}
		
		tx_mutex.Unlock()
		return // 有交易无效*/
		res_flag = true
	}

	// 更新区块池及其删除标记
	tx_no_double_spend := true // 12.7修改
	for _, tx := range new_block.Transactions {
		var tx_used_UTXO_set map[string]int = make(map[string]int) // 12.7修改, 避免double spend的修改方式是将后一个交易所产生的影响全部取消，且不记录在block中
		key := string(local_bc.Hash(tx))
		// 删除区块池中本地未提交的交易时，对应“用掉的 UTXO 集合”也需要更新
		for _, input := range tx.Inputs {
			output_key := local_bc.Hash(local_bc.OutputKey{TransactionID: input.TransactionID, Index: input.OutputID})
			if _, ok := tmp_used_UTXO_set[string(output_key)]; !ok {
				tx_no_double_spend = false
				continue
			} // 12.7修改
			tx_used_UTXO_set[string(output_key)] = 1      // 12.7修改
			delete(tmp_used_UTXO_set, string(output_key)) // 这里似乎保证不了 input 一定在 set 中，还需要检验，但是这个操作是安全的 12.6
		}
		if !tx_no_double_spend {
			for key := range tx_used_UTXO_set {
				tmp_used_UTXO_set[key] = 1
			}
		} // 12.7修改
		transaction_flag[key] = 1
		delete(transaction_pool, string(local_bc.Hash(tx))) // 这里是不是可以缓一缓，等到确定新区块在主链上并Validate后再删除？ 12.6
	}
	tx_mutex.Unlock()

	
	if true { // new_block.Validate(tmp_UTXO_set, stale_message_set, creation_reward) { 12.6修改

		// 如果新区块通过验证，将它加入本地的区块树
		fmt.Println("New block accepted!")
		father_block := block_pool[father_index]
		new_block.TotalDifficulty = father_block.TotalDifficulty + int64(new_block.Difficulty)

		block_pool = append(block_pool, new_block)
		new_block_hash := local_bc.Hash(local_bc.BlockHashData{Head: new_block.Head, Transactions: new_block.Transactions})

		fmt.Println("-------------------- New Block Info ----------------------")
		j, _ := json.Marshal(new_block)
		fmt.Printf("\nJSON: %v\n", string(j))
		fmt.Printf("Hash: %x\n\n", new_block_hash)
		fmt.Println("----------------------------------------------------------")

		block_index[string(new_block_hash)] = len(block_pool) - 1
		is_in_main_chain = append(is_in_main_chain, false) // 首先假设新区块不在主链中

		// 然后我们来判断新区块到底应不应该在主链中。有两种情况：
		if father_index != main_chain_head { // 新区块的父区块不是主链头，也就是说它增长的不是主链，此时发生了 blockchian fork
			// 应用 longest chain rule。总难度严格更大的时候，我们切换到新的主链
			if new_block.TotalDifficulty > block_pool[main_chain_head].TotalDifficulty {
				main_UTXO_set = tmp_UTXO_set // 刚才计算的临时集合以后就是正式的了
				main_UTXO_key_set = tmp_UTXO_key_set
				// 因为切换了主链所以需要更新 is_in_main_chain. 也是到 LCA 一上一下
				tmp_index := main_chain_head
				for tmp_index != LCA {
					is_in_main_chain[tmp_index] = false
					current_block := block_pool[tmp_index]
					tmp_index = current_block.FatherBlockIndex
				}
				for _, index := range ancestor_list {
					is_in_main_chain[index] = true // modified at 12/05 11:34
				}
				local_bc.ApplyBlock(&new_block, main_UTXO_set, main_UTXO_key_set) // 因为新区块加入了主链，把它应用到主链的 UTXO 集合上
				main_chain_head = len(block_pool) - 1
				is_in_main_chain[len(block_pool)-1] = true
			}
		} else { // 另一种情况，新区块增长的是主链，那它自然就在主链中
			local_bc.ApplyBlock(&new_block, main_UTXO_set, main_UTXO_key_set)
			main_chain_head = len(block_pool) - 1
			is_in_main_chain[len(block_pool)-1] = true
		}
	}
}

// 处理新的交易
func onNewTransactionAppearance(new_tx local_bc.Transaction) {
	fmt.Println("New transaction appears")

	// 首先验证交易是否包含已经用过的签名消息。如果有的话，直接拒绝这样的交易
	// (我为啥要这样干来着？)
	valid_flag := true
	new_hash := string(local_bc.Hash(new_tx))
	for _, input := range new_tx.Inputs {
		if value, ok := stale_message_set[string(input.M)]; (ok && new_hash != value) {
			valid_flag = false
		}
	}
	if valid_flag {
		tx_mutex.Lock()
		// 如果不存在之前区块打上的“这个交易应该删除”的标记，加入交易池
		key := string(local_bc.Hash(new_tx))
		if flag, ok := transaction_flag[key]; !(ok && flag == 1) {
			transaction_pool[key] = new_tx
		}
		tx_mutex.Unlock()
	}
}

// 该后台线程处理从广播收到的区块
func newBlockProcessingThread() {
	for {
		new_block := <-task_queue_block
		onNewBlockAppearance(new_block)
	}
}

// 该后台线程处理从广播收到的交易
func newTransactionProcessingThread() {
	for {
		new_tx := <-task_queue_tx
		onNewTransactionAppearance(new_tx)
	}
}

// 该后台线程执行工作量证明，挖掘新的区块
func miningThread() {
	last_time_head := -1
	try_next_block := new(local_bc.Block) // 这是尝试挖掘的区块
	// hash := sha256.New()
	for {
		chain_mutex.Lock()
		if last_time_head != main_chain_head {
			fmt.Printf("Prepared: main_chain_head = %v\n", main_chain_head)
			*try_next_block = local_bc.Prepare(&block_pool[main_chain_head],
				overall_difficulty, overall_log_difficulty, my_public_key) // 每次根据当前的主链头准备一个新区块进行挖掘
			last_time_head = main_chain_head
		}
		chain_mutex.Unlock()
		try_next_block.Head.Nonce = rand.Int63() // 随机试数
		hash_bytes := local_bc.Hash(try_next_block.Head)
		diff_mutex.Lock()
		if local_bc.ValidatePoW(overall_log_difficulty, hash_bytes) {
			fmt.Println("!!!!!!! PoW succeed !!!!!!!")
			try_next_block.LogDifficulty = overall_log_difficulty
			try_next_block.Difficulty = overall_difficulty
			diff_mutex.Unlock()
			tx_mutex.Lock()
			// 遍历交易池
			for key, tx := range transaction_pool {
				// 如果没有删除标记，则加入所要打包的交易列表（我为啥要判断这个来着？）
				if flag, ok := transaction_flag[key]; !(ok && flag == 1) {
					tx.Outputs[0].PublicKey = my_public_key //接收者为自己
					try_next_block.Transactions = append(try_next_block.Transactions, tx)
				}
			}
			transaction_pool = make(map[string]local_bc.Transaction) // 清空交易池
			tx_mutex.Unlock()
			try_next_block.Transactions = append(try_next_block.Transactions,
				local_bc.BlockCreationReward(my_private_key, my_public_key, creation_reward)) // 加入挖矿奖励
			onNewBlockAppearance(*try_next_block) // 首先把新区块提交给自己

			network.BroadcastBlock(*try_next_block) // 然后再广播给其矿工
			//*try_next_block = local_bc.Prepare(&block_pool[main_chain_head],
				//overall_difficulty, overall_log_difficulty, my_public_key) // 每次根据当前的主链头准备一个新区块进行挖掘
		} else {
			diff_mutex.Unlock()
		}
	}
}

// 阻塞当前进程，直到所有矿工的公钥都已收到
func waitAllKey() {
	for {
		valid_flag := true
		for _, node := range node_list {
			valid_flag = node.KeyValid && valid_flag
		}
		if valid_flag {
			break
		}
	}
}

// 用来接受 rpc 请求
func receiveBroadcast(listener net.Listener) {
	for {
		connection, error_accept := listener.Accept()
		if error_accept != nil {
			fmt.Println("Broadcast connection failed:", error_accept)
			log.Fatal(error_accept)
		}
		go func() {
			rpc.ServeConn(connection)
		}()
	}
}

func main() {

	// 命令行参数。应该还有更多，比如通过命令行参数指定节点列表，等等
	my_ip_ptr := flag.String("my-ip", "10.1.0.115", "The IP address that miner holds")
	log_difficulty_ptr := flag.Int("log-difficulty", 24, "The number of prefix zeroes of SHA-256 required for PoW success")
	flag.Parse()
	my_ip = *my_ip_ptr
	overall_log_difficulty = *log_difficulty_ptr // 新建一个线程过一段时间增加一点难度？ 12.6
	overall_difficulty = 1 << overall_log_difficulty

	// 初始化 RPC
	gob.Register(network.NodeArgs{})
	gob.Register(BlockchainUpdateArgs{})
	gob.Register(TransactionUpdateArgs{})
	blockchain_service := new(BlockchainService)
	rpc.Register(blockchain_service)

	// 监听来自矿工的连接
	broadcast_listener, error_listen := net.Listen("tcp", ":"+port)
	if error_listen != nil {
		log.Fatal(error_listen)
	}
	defer broadcast_listener.Close()

	// 尝试连接所有其他矿工
	network.InitConnection(node_list, my_ip)
	defer network.CloseConnection()

	// 启动 rpc 服务
	go receiveBroadcast(broadcast_listener)

	// 获取自己的密钥对并且把公钥广播出去
	my_private_key = signature.KeyGen()
	raw_my_public_key := my_private_key.PublicKey
	my_public_key = signature.PublicKeyToJSON(raw_my_public_key)
	for index := range node_list {
		if node_list[index].IP == my_ip {
			node_list[index].PublicKey = my_public_key
			node_list[index].KeyValid = true
			network.BroadcastKey(my_public_key, node_list[index])
		}
	}
	waitAllKey() // 等待收到其他矿工的公钥

	// 创建创世区块，初始化区块链
	genesis_block := local_bc.Genesis()
	genesis_block_hash := local_bc.Hash(local_bc.BlockHashData{Head: genesis_block.Head, Transactions: genesis_block.Transactions})
	block_index[string(genesis_block_hash)] = 0
	block_pool = append(block_pool, local_bc.Genesis())
	is_in_main_chain = append(is_in_main_chain, true)
	main_chain_head = 0

	// 启动后台工作线程
	go newBlockProcessingThread()
	go newTransactionProcessingThread()
	go miningThread()

	// 开始处理来自客户端的 rpc 请求
	gob.Register(ClientTransactionArgs{})
	gob.Register(ClientQueryArgs{})
	client_service := new(ClientService)
	rpc.Register(client_service)
	client_listener, error_listen_client := net.Listen("tcp", ":"+port_client)
	if error_listen_client != nil {
		log.Fatal(error_listen_client)
	}
	defer client_listener.Close()
	for {
		connection, error_accept := client_listener.Accept()
		if error_accept != nil {
			fmt.Println("Connection from client failed:", error_accept)
			log.Fatal(error_accept)
		}
		go func() {
			rpc.ServeConn(connection)
		}()
	}

}

/*func difficultyCall(conn *rpc.Client, node_id int) {
	defer conn.Close()
	for {
		time.Sleep(10 * time.Second)

	}
}*/

/* ------------- service for client --------------- */

var port_client = "8057" // 客户端向矿工发送请求所用的端口

// 面向客户端的 rpc 服务
type ClientService struct{}
type ClientTransactionArgs struct {
	DestinationID int
	Value         float32
}

type ClientQueryArgs struct {
	Account int
}

type DifficultyArgs struct {
	NewLogDifficulty int
}

// 1.8修改：客户端调整 PoW 的 overall_log_difficulty
func (t *ClientService) AdjustDifficulty(request_args *DifficultyArgs, reply *int) error {
	diff_mutex.Lock()
	overall_log_difficulty = request_args.NewLogDifficulty
	overall_difficulty = 1 << overall_log_difficulty
	diff_mutex.Unlock()
	return nil
}


// 基于已确认的交易，客户端请求矿工给另一个指定的矿工转帐
// 这里客户端仅仅指定目标和数量。具体 input 用哪些，由矿工 Collect 方法决定
// !!!!!! 这一块逻辑如果和主链切换的过程叠加在一起可能会有点问题，目前还没想清楚，但是触发概率应该很小很小 !!!!!!
func (t *ClientService) ClientTransaction(request_args *ClientTransactionArgs, reply *float32) error {
	if request_args.DestinationID >= len(node_list) { // 指定的对象编号越界
		*reply = -2
		return nil
	}
	collected_value, output_list := Collect(request_args.Value) // 搜集当前自己有哪些没用的 UTXO
	if collected_value < request_args.Value {                   // 钱不够
		*reply = -1
		return nil
	}
	*reply = collected_value // 如果一切正常，返回值为消耗的 UTXO 的总价值

	// 准备交易的 input
	inputs := make([]local_bc.TransactionInput, 0)
	tx_mutex.Lock()
	for _, output := range output_list {
		message := local_bc.GetCurrentTime()
		r, s, _ := signature.SignMessage(my_private_key, message)
		inputs = append(inputs, local_bc.TransactionInput{
			M:             message,
			R:             r,
			S:             s,
			TransactionID: output.TransactionID,
			OutputID:      output.Index,
		})
		time.Sleep(2 * time.Microsecond) // 防止微秒级别的时间戳冲突，保证消息互异
		key := local_bc.Hash(output)
		tmp_used_UTXO_set[string(key)] = 1 // 标记一下当前有一个未打包提交的交易消耗了这条 UTXO
	}
	tx_mutex.Unlock()

		// 准备交易的 output
		dest_public_key := node_list[request_args.DestinationID].PublicKey
		outputs := make([]local_bc.TransactionOutput, 0)
		timestamp := local_bc.GetCurrentTime()
		out_value := request_args.Value
		var message bytes.Buffer
		message.Write(local_bc.Hash(dest_public_key))
		binary.Write(&message, binary.LittleEndian, out_value)
		message.Write(timestamp)
		r, s, _ := signature.SignMessage(my_private_key, message.Bytes())
		outputs = append(outputs, local_bc.TransactionOutput{
			PublicKey: dest_public_key,
			Value:     request_args.Value,
			M:         timestamp,
			R:         r,
			S:         s,
		})
		
		if collected_value > out_value { // 如果有 input 的量富裕的话，多出来的还要留下来给自己。但是这一部分留出来的在这条交易被打包提交之前不能用
			timestamp = local_bc.GetCurrentTime()
			recollect_value := collected_value - out_value
			var message_self bytes.Buffer
			message_self.Write(local_bc.Hash(my_public_key))
			binary.Write(&message_self, binary.LittleEndian, recollect_value)
			message_self.Write(timestamp)
			r_self, s_self, _ := signature.SignMessage(my_private_key, message_self.Bytes())
			outputs = append(outputs, local_bc.TransactionOutput{
				PublicKey: my_public_key,
				Value:     collected_value - request_args.Value,
				M:         timestamp,
				R:         r_self,
				S:         s_self,
			})
		}

	new_tx := local_bc.Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
	onNewTransactionAppearance(new_tx)   // 先放进自己的交易池
	network.BroadcastTransaction(new_tx) // 再广播给其他矿工
	return nil
}

// 基于已确认的交易，客户端查询当前矿工的余额
func (t *ClientService) ClientQuery(request_args *ClientQueryArgs, reply *float32) error {
	*reply = QueryBalance()
	return nil
}

// 查询余额
func QueryBalance() float32 {
	var sum float32 = 0
	// 仍然通过数字签名的方式来验证每笔 output 的归属
	tmp_message := make([]byte, 32)
	chain_mutex.Lock()
	tx_mutex.Lock()
	r, s, _ := signature.SignMessage(my_private_key, tmp_message)
	for _, output := range main_UTXO_set {
		if signature.VerifySignature(output.PublicKey, tmp_message, r, s) {
			sum += output.Value
		}
	}
	tx_mutex.Unlock()
	chain_mutex.Unlock()
	return sum
}

// 尽量收集当前主链 UTXO 集合中，未被其他未提交交易用掉的 output，直到总价值满足 target
func Collect(target float32) (float32, []local_bc.OutputKey) {
	// 仍然通过数字签名的方式来验证每笔 output 的归属
	tmp_message := make([]byte, 32)
	r, s, _ := signature.SignMessage(my_private_key, tmp_message)
	var sum float32 = 0
	var output_list []local_bc.OutputKey = make([]local_bc.OutputKey, 0)
	chain_mutex.Lock()
	tx_mutex.Lock()
	for key, output := range main_UTXO_set {
		if _, ok := tmp_used_UTXO_set[key]; ok { // 用过了不能再用了
			continue
		}
		if signature.VerifySignature(output.PublicKey, tmp_message, r, s) {
			sum += output.Value
			output_list = append(output_list, main_UTXO_key_set[key])
		}
		if sum >= target {
			break
		}
	}
	tx_mutex.Unlock()
	chain_mutex.Unlock()
	return sum, output_list
}
