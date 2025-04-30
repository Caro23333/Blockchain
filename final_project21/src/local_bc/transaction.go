package local_bc

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
    "bytes"
	"fmt"
	"log"
	"project2/signature"
	"time"
)

// OutputKey 在所有已经 confirm 的交易中唯一地确定一个 output
type OutputKey struct {
	TransactionID []byte
	Index int
}

// 交易的单个 input
type TransactionInput struct {
	M []byte	// 被签名的消息
	R []byte	 
	S []byte	// 数字签名
	TransactionID []byte // 从哪个交易的
	OutputID int 	// 哪个 output。这两个字段实际上组成了 OutputKey
}

// 交易的单个 output
type TransactionOutput struct {
	PublicKey signature.JSONPublicKey // 这笔钱送给谁（用公钥表示身份）
	Value float32 // 送了多少
    M []byte // 用于签名的消息
    R []byte
    S []byte // 数字签名
}

// 交易
type Transaction struct {
	Inputs []TransactionInput
	Outputs []TransactionOutput
}

// 可以对任意类型的变量取 SHA-256 哈希值。注意该哈希值是长度为 32 的 []byte, 但当它作为键值使用时我们需要将其等价转换为 string
func Hash(item interface{}) []byte {
	json_data, err := json.Marshal(item)
	if err != nil {
		log.Fatal(err)
	}
	hash := sha256.New()
	hash.Write(json_data)
	hash_bytes := hash.Sum(nil)
	return hash_bytes
}

// 验证一条交易是否合法。为此，我们需要知道：
// 		- 当前有哪些 output 没被用过 (UTXO_set)
//		- 当前有哪些 message 被用过 (stale_message_set)。只有 message 没被用过时它的签名才有效力。
//				* 这是因为密码学安全的数字签名机制只能保证攻击者无法伪造“新消息”的数字签名。
//				* 毕竟如果某个人签名了消息 m, 这个签名将会公布。攻击者只要记录下来这个签名就可以伪造针对消息 m 的签名。
//		- 对应区块创建者的公钥 (sender_public_key)。我们只容许至多一条来自区块创建者的 coin creation 交易（即挖矿奖励）。
//				* 是否有更多条则由 has_creation_before 来控制。
// 				* coin creation 的值必须恰好符合指定的 reward 大小。
func (tx *Transaction) Validate(used_UTXO_set map[string]bool, UTXO_set map[string]TransactionOutput, stale_message_set map[string]string, 
	sender_public_key signature.JSONPublicKey, creation_reward float32, has_creation_before *bool) bool {
	var total_input_value float32 = 0
	var total_output_value float32 = 0
	var current_hash string = string(Hash(*tx))
	res_flag := true // 记录结果
    var tx_source_public_key signature.JSONPublicKey
	for _, output := range tx.Outputs {
        if output.Value < 0 {
            fmt.Println("Negative value")
            return false
        }
		total_output_value += output.Value
	}
	for _, input := range tx.Inputs {
		// 首先验证消息是否被用过
		stale_hash, ok := stale_message_set[string(input.M)]
		if ok && stale_hash != current_hash {
			fmt.Println("Stale message")
			res_flag = false
		}
		// stale_message_set[string(input.M)] = current_hash 只在验证成功时更新stale状态
		if input.OutputID == -1 { // coin creation. 验证是否为区块创建者的挖矿奖励
			fmt.Println("Validating coin creation")
			if *has_creation_before {
				return false
			}
			*has_creation_before = true
			res_flag = res_flag && total_output_value == creation_reward && 
				signature.VerifySignature(sender_public_key, input.M, input.R, input.S)
			if res_flag {
				stale_message_set[string(input.M)] = current_hash
			}
			return res_flag
		} else { // 普通的交易
			fmt.Println("Validating normal transaction")
			// 验证是否有对应的 UTXO；如果有，再验证数字签名是否合法（是否得到授权）
			input_UTXO, ok := UTXO_set[string(Hash(OutputKey{input.TransactionID, input.OutputID}))]
			not_ok := used_UTXO_set[string(Hash(OutputKey{input.TransactionID, input.OutputID}))]
			if !(ok && signature.VerifySignature(
				input_UTXO.PublicKey, input.M, input.R, input.S)) || not_ok {
				fmt.Println("UTXO verification fail")
				return false
			}
			used_UTXO_set[string(Hash(OutputKey{input.TransactionID, input.OutputID}))] = true
			total_input_value += input_UTXO.Value
            tx_source_public_key = input_UTXO.PublicKey // 用于验证 outputs 的签名
		}
	} 
    // 对普通交易，验证outputs的签名是否合法
    for _, output := range tx.Outputs {
        sign_time := output.M
		stale_hash, ok := stale_message_set[string(sign_time)]
		if ok && stale_hash != current_hash {
			fmt.Println("Stale message")
			res_flag = false
		}
        sign_value := output.Value
        sign_public_key := output.PublicKey
        var sign_buf bytes.Buffer
        sign_buf.Write(Hash(sign_public_key))
        binary.Write(&sign_buf, binary.LittleEndian, sign_value)
        sign_buf.Write(sign_time)
        if !signature.VerifySignature(tx_source_public_key, sign_buf.Bytes(), output.R, output.S) {
            fmt.Println("Output signature verification fail")
            res_flag = false
        }
    }
	// 验证总输出量是否不多于总输入量。普通的交易不能无中生有
	if (total_input_value < total_output_value) {
		res_flag = false
	}
    if res_flag {
        for _, input := range tx.Inputs {
            stale_message_set[string(input.M)] = current_hash
        }
		for _, output := range tx.Outputs {
			stale_message_set[string(output.M)] = current_hash
		}
    }
	return res_flag
}

// 获取 UNIX 时间戳（微秒为单位）并转换成 []byte 类型
// 诚实节点将时间戳作为每次签名的信息，这样就不会重复
func GetCurrentTime() []byte {
	time_stamp := time.Now().UnixNano()
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(time_stamp))
	return buf
}

// 基于区块创建者的信息，返回一条区块创建奖励。
func BlockCreationReward(my_private_key *ecdsa.PrivateKey, my_public_key signature.JSONPublicKey, creation_reward float32) Transaction {
	message := GetCurrentTime()
	r, s, _ := signature.SignMessage(my_private_key, message)
	new_tx := Transaction {
		Inputs: []TransactionInput{
			{
				M: message,
				R: r,
				S: s,
				TransactionID: Hash(0), // 仅作占位
				OutputID: -1, // OutputID 为 -1 是“无中生有”的标志
			},
		},
		Outputs: []TransactionOutput{
			{
				PublicKey: my_public_key,
				Value: creation_reward,
                M: message,
                R: r,
                S: s,
			},
		},
	}
	return new_tx
}