package local_bc

import (
	// "crypto/sha256"
	// "encoding/hex"
	// "encoding/json"
	"fmt"
	"project2/signature"
	"time"
)

// 区块头
// 工作量证明只对区块头计算哈希
type BlockHeader struct {
	FatherBlockHash []byte // 父区块的哈希
	Nonce int64 // 工作量证明时生成的随机数
	TimeStamp int64 // 时间戳。功能约等于随机盐
}


// 1.8 修改
type BlockHashData struct {
	Head BlockHeader
	Transactions []Transaction
}

// 区块
// 注：难度 = 工作量证明成功一次期望所需的随机次数
// 假设 SHA-256 在 {0, 1}^256 上均匀随机，则 log(难度) = 我们要求 SHA-256 值拥有的前导零数量
type Block struct {
	Head BlockHeader 
	FatherBlockIndex int // 父区块的编号。每个节点存了一个区块的数组（“区块池”），可以通过编号访问每个已有的区块
	Transactions []Transaction // 区块的所有交易
	TotalDifficulty int64 // 从创世区块到当前区块为止一共积累了多少“难度”。blockchain fork时我们总是选取总难度更大的链，即 longest chain rule
	Difficulty int // 单个区块的难度
	LogDifficulty int // 要求的前导零数量
	CreatorPublicKey signature.JSONPublicKey // 区块创建者的公钥，用来验证区块奖励
}

// 创建一个创世区块，用于初始化区块链
func Genesis() Block {
	res := Block {
		Head: BlockHeader {
			FatherBlockHash: make([]byte, 32),
			Nonce: 0,
			TimeStamp: 0,
		},
		FatherBlockIndex: -1, // 这是创世区块的标志
		Transactions: make([]Transaction, 0),
		TotalDifficulty: 0,
	}
	return res
}

// 根据上一个区块，准备出下一个区块的“雏形”供矿工挖矿
// 挖到这个区块的矿工会把 Nonce 和 Transactions 填满之后广播这个区块
func Prepare(last_block *Block, difficulty int, log_difficulty int, my_public_key signature.JSONPublicKey) Block {
	hash_bytes := Hash(BlockHashData{last_block.Head, last_block.Transactions})
	new_block := Block {
		Head: BlockHeader {
			FatherBlockHash: hash_bytes,
			Nonce: 0,
			TimeStamp: time.Now().UnixNano(),
		},
		FatherBlockIndex: -1, // 占位。当它被广播到每个节点时，再根据区块头中的哈希值，从对应节点的区块池中找到父区块的下标
		Transactions: make([]Transaction, 0),
		TotalDifficulty: last_block.TotalDifficulty + int64(difficulty),
		Difficulty: difficulty,
		LogDifficulty: log_difficulty,
		CreatorPublicKey: my_public_key,
	}
	return new_block
}

// 假设现在的 UTXO 集合是 UTXO_set，在把 block 这个区块里面的所有交易都执行之后，更新 UTXO_set 的状态
// UTXO_key_set 存储的是 UTXO 的 OutputKey，而 UTXO_set 存的是 Output 本身
func ApplyBlock(block *Block, UTXO_set map[string]TransactionOutput, UTXO_key_set map[string]OutputKey) {
	for _, tx := range block.Transactions {
		tx_id := Hash(tx)
		for _, input := range tx.Inputs { // 执行之后所有输入的 UTXO 都被用掉了
			key := string(Hash(OutputKey{input.TransactionID, input.OutputID}))
			delete(UTXO_set, key)
			delete(UTXO_key_set, key)
		}
		for index, output := range tx.Outputs { // 输出的变成了新的 UTXO
			key := string(Hash(OutputKey{tx_id, index}))
			UTXO_set[key] = output
			UTXO_key_set[key] = OutputKey{tx_id, index}
		}
	}
}

// 假设现在的 UTXO 集合是 UTXO_set 并且 block 这个区块已经都被执行了，在把所有交易都复原之后，更新 UTXO_set 的状态
func RevertBlock(block *Block, UTXO_set map[string]TransactionOutput, UTXO_key_set map[string]OutputKey, block_pool []Block) {
	for _, tx := range block.Transactions {
		tx_id := Hash(tx)
		for _, input := range tx.Inputs { // 原来输入用掉的 UTXO 回来了
			// 因为我们没办法根据 input 里面的信息直接定位到需要恢复的 output（在恢复之前，它不在任何一个 map 里面），所以要遍历区块链直到找到
			key := string(Hash(OutputKey{input.TransactionID, input.OutputID}))
			tmp_index := block.FatherBlockIndex
			found_flag := false
			for tmp_index != -1 {
				current_block := block_pool[tmp_index]
				for _, cr_tx := range current_block.Transactions {
					cr_tx_id := Hash(cr_tx)
					if string(cr_tx_id) == string(input.TransactionID) && input.OutputID < len(cr_tx.Outputs) {
						UTXO_set[key] = cr_tx.Outputs[input.OutputID]
						UTXO_key_set[key] = OutputKey{input.TransactionID, input.OutputID}
						found_flag = true
						break
					}
				}
				if found_flag {
					break
				}
				tmp_index = current_block.FatherBlockIndex
			}
		}
		for index := range tx.Outputs { // 原来输出的 UTXO 没了
			key := string(Hash(OutputKey{tx_id, index}))
			delete(UTXO_set, key)
			delete(UTXO_key_set, key)
		}
	}
}

var masks [8]uint8 = [8]uint8{0x80, 0xc0, 0xe0, 0xf0, 0xf8, 0xfc, 0xfe, 0xff}

// 根据要求的前导零数量，验证工作量证明是否正确完成
func ValidatePoW(log_difficulty int, hash_bytes []byte) bool {
	for i := 0; i < 32; i++ {
		if i * 8 >= log_difficulty {
			break
		} else if (i + 1) * 8 < log_difficulty {
			if hash_bytes[i] != 0 {
				return false
			}
		} else {
			if hash_bytes[i] & masks[log_difficulty - i * 8 - 1] != 0 {
				return false
			}
		}
	}
	return true
}

// 12.6 添加
// 验证每一条交易，并给出不合法的交易
func (block *Block) ValidateTransaction(UTXO_set map[string]TransactionOutput, stale_message_set map[string]string, creation_reward float32) (bool, []Transaction) {
    var invalid_txs []Transaction
    res_flag := true
	has_creation_before := false
	var used_UTXO_set map[string]bool = make(map[string]bool)
	for index, tx := range block.Transactions {
		if !tx.Validate(used_UTXO_set, UTXO_set, stale_message_set, block.CreatorPublicKey, creation_reward, &has_creation_before) {
			fmt.Println("Block invalid on transaction", index)
			res_flag = false
            invalid_txs = append(invalid_txs, tx)
		}
	}
    return res_flag, invalid_txs
}

// 验证区块是否合法
// 如你所见，首先验证工作量证明，再依次验证每一条交易
func (block *Block) Validate(UTXO_set map[string]TransactionOutput, stale_message_set map[string]string, creation_reward float32) bool {
	hash_bytes := Hash(block.Head)
	if !ValidatePoW(block.LogDifficulty, hash_bytes) {
		return false
	}
	/*res_flag := true
	has_creation_before := false
	for _, tx := range block.Transactions {
		if !tx.Validate(UTXO_set, stale_message_set, block.CreatorPublicKey, creation_reward, &has_creation_before) {
			fmt.Println("Block invalid")
			res_flag = false
		}
	} 12.6 修改*/
    res_flag, _ := block.ValidateTransaction(UTXO_set, stale_message_set, creation_reward) // 12.6 修改 增加，实际不使用本函数
	return res_flag
}
