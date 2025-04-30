package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"project2/network"
	"project2/signature"
	"strings"
	"time"
)

// 客户端这些方法都和服务端的意思差不多
// 目前只支持了这两台主机，可以改成所有的机器

const command_file_path string = "temp/command.txt"
const result_file_path string = "temp/result.txt"
const file_async_error_tolerating_timeout uint = 50
const ending_symbol string = "#-end-#"
const empty_symbol string = "#-empty-#"
const flush_result_command string = "flush-result"
const flush_result_token string = "@-please-flush-result-@"

var random_generator *rand.Rand
var latest_timestamp uint64 = 0

var node_list []network.NodeArgs = []network.NodeArgs{
	{IP: "10.1.0.91", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.92", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.94", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.95", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.96", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.98", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.99", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.102", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.119", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.104", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.105", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.107", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.109", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.110", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.111", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.112", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.113", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.115", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.116", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
	{IP: "10.1.0.117", Port: "8057", PublicKey: *new(signature.JSONPublicKey)},
}

var connection_pool []*rpc.Client = make([]*rpc.Client, 20)
var my_ip string = "10.1.0.99"
var port string = "8056"
var port_client string = "8057"

// =========================================================================================================================
// ======================================================== 网络连接 ========================================================
// =========================================================================================================================

// 第二个参数的含义: 0 == 成功, 1 == 失败
func trySingleConnection(index int) (string, int) {
	if index < 0 || index >= 20 {
		return fmt.Sprintf("index = %d out of range. It should be in [0,19].", index), 1
	}
	if node_list[index].IP == my_ip {
		return fmt.Sprintf("index = %d has the same ip to my_ip = %s.", index, my_ip), 1
	}
	if connection_pool[index] != nil {
		return fmt.Sprintf("Connection from client to miner of %d already exists.", index), 1
	}
	miner, error_dial := rpc.Dial("tcp", node_list[index].IP+":"+node_list[index].Port)
	if error_dial != nil {
		return fmt.Sprintf("Error connecting to miner of %d at %v: %v.", index, node_list[index].IP, error_dial), 1
	}
	var greet_value int = 1
	var greet_res string
	miner.Call("ClientService.Greet", greet_value, &greet_res)
	if greet_res == "GREET" {
		connection_pool[index] = miner
		return fmt.Sprintf("Successfully connected to miner %d.", index), 0
	} else {
		connection_pool[index] = nil
		return fmt.Sprintf("Error greeting to miner %d.", index), 1
	}
}

// func initConnection(node_list []network.NodeArgs) {
// 	for index := range node_list {
// 		description, state := trySingleConnection(index)
// 		if state != 0 {
// 			fmt.Printf("%s\n", description)
// 		}
// 	}
// }

func closeConnection() {
	for _, connection := range connection_pool {
		connection.Close()
	}
}

// =========================================================================================================================
// ===================================================== 客户端提供的功能 ====================================================
// =========================================================================================================================

type ClientTransactionArgs struct {
	DestinationIP string
	Value         float32
	Fee           float32
}

type ClientQueryArgs struct {
	Account int
}

type DifficultyArgs struct {
	NewLogDifficulty int
}

type NewConnectionArgs struct {
	NewNeighborIP string
	NewIPPort     string
}

func query() string {
	var ans string = ""
	for index, connection := range connection_pool {
		if connection != nil {
			var query_res float32 = 0.0
			connection.Call(
				"ClientService.ClientQuery",
				ClientQueryArgs{0},
				&query_res)
			if query_res >= 0 {
				ans = ans + fmt.Sprintf("%d,%f ", index, query_res)
			}
		}
	}
	if ans == "" {
		return ""
	} else {
		return ans[:len(ans)-1]
	}
}

func tx(source int, dest int, value float32, fee float32) string {
	if source < 0 || source >= len(connection_pool) {
		return fmt.Sprintf("source = %d is out of range. It should be in [0,%d].", source, len(connection_pool)-1)
	}
	if dest < 0 || dest >= len(connection_pool) {
		return fmt.Sprintf("dest = %d is out of range. It should be in [0,%d].", dest, len(connection_pool)-1)
	}
	if connection_pool[source] == nil {
		return fmt.Sprintf("There are no connection from client to source miner %d.", source)
	}
	var tx_flag float32 = 0
	connection_pool[source].Call(
		"ClientService.ClientTransaction",
		ClientTransactionArgs{
			DestinationIP: node_list[dest].IP,
			Value:         value,
			Fee:           fee,
		},
		&tx_flag)
	if tx_flag >= 0 {
		return fmt.Sprintf("Transaction successfully submitted, spend confirmed balance of %v.", tx_flag)
	} else if tx_flag == -1 {
		return "No enough confirmed balance to make the transaction."
	} else if tx_flag == -2 {
		return "Transaction destination is invalid."
	} else if tx_flag == -3 {
		return "New node is not prepared yet."
	} else {
		return fmt.Sprintf("Unexpected Error! ClientService.ClientTransaction at miner returns tx_flag = %f.", tx_flag)
	}
}

func adjust(log_difficulty int) string {
	for _, connection := range connection_pool {
		if connection != nil {
			connection.Call(
				"ClientService.AdjustDifficulty",
				DifficultyArgs{log_difficulty},
				nil)
		}
	}
	return fmt.Sprintf("Successfully adjusted log_difficulty to %d.", log_difficulty)
}

func connect(source int, dest int) string {
	if source < 0 || source >= len(connection_pool) {
		return fmt.Sprintf("source = %d is out of range. It should be in [0,%d].", source, len(connection_pool)-1)
	}
	if dest < 0 || dest >= len(connection_pool) {
		return fmt.Sprintf("dest = %d is out of range. It should be in [0,%d].", dest, len(connection_pool)-1)
	}
	if connection_pool[source] == nil {
		return fmt.Sprintf("There are no connection from client to source miner %d.", source)
	}
	var return_flag int = 0
	connection_pool[source].Call(
		"ClientService.ConnectTo",
		NewConnectionArgs{
			NewNeighborIP: node_list[dest].IP,
			NewIPPort:     port,
		},
		&return_flag)
	if return_flag == 0 {
		return "Successfully added connection."
	} else if return_flag == 1 {
		return "Connection already exists."
	} else if return_flag == 2 {
		return "RPC dial failed."
	} else if return_flag == 3 {
		return "RPC call failed."
	} else {
		return fmt.Sprintf("Unexpected Error! ClientService.ConnectTo at miner returns return_flag = %d.", return_flag)
	}
}

func newnode(index int) string {
	description, _ := trySingleConnection(index)
	return description
}

// =========================================================================================================================
// ================================================ 文件通讯系统辅助函数 =====================================================
// =========================================================================================================================

func checkEnding(s string) bool {
	l := strings.Split(s, " ")
	return l[len(l)-1] == ending_symbol
}

func getLatestTimestamp(file_path string) uint64 {
	var latest_timestamp_ans uint64 = 0
	content_bytes, error_read_content := os.ReadFile(file_path)
	if error_read_content != nil {
		fmt.Printf("Error reading file %s for timestamp!", file_path)
		log.Fatal(error_read_content)
	}
	content_string := string(content_bytes)
	if content_string == empty_symbol { // 由deploy.py初始化时生成的
		latest_timestamp_ans = 0
	} else {
		_, error_parse_latest_timestamp := fmt.Sscanf(content_string, "%d", &latest_timestamp_ans)
		if error_parse_latest_timestamp != nil {
			fmt.Printf("Error parsing latest timestamp in file %s!\n", file_path)
			log.Fatal(error_parse_latest_timestamp)
		}
	}
	return latest_timestamp_ans
}

// =========================================================================================================================
// ===================================================== 文件通讯系统 =======================================================
// =========================================================================================================================

// 假定client不可能在一个result写到一半时被关闭, 这样这里就避免了匹配尾部等工作/问题
func initLatestTimestamp() {
	latest_command_timestamp := getLatestTimestamp(command_file_path)
	latest_result_timestamp := getLatestTimestamp(result_file_path)
	if latest_command_timestamp != latest_result_timestamp {
		fmt.Println("Command and result file have different timestamp!")
		log.Fatal("Command and result file have different timestamp!")
	}
	latest_timestamp = latest_command_timestamp
}

// 第二个返回值的含义:
// 0 == 没有新命令 (也有可能是file async)
// 1 == 成功处理新命令
// 2 == 还在file async中
// 3 == 遇到错误, 此时必须像没错误一样, 把错误传递回去, 并且也要照常更新latest_timestamp
// client最重要的是不能死, 而且要尽量处理所有请求, 因为helper在等待
// client几乎一直开着
func readCommandAndExecute() (string, int) {
	command_bytes, error_read_command := os.ReadFile(command_file_path)
	if error_read_command != nil {
		fmt.Println("Error reading command file! (may due to file async)")
		return "", 2
	}
	command_string := string(command_bytes)
	if command_string == empty_symbol { // 由deploy.py初始化时生成的
		return "", 0
	}
	if !checkEnding(command_string) {
		fmt.Println("Ending does not match! (may due to file async)")
		return "", 2
	}

	var command_timestamp uint64
	var command_type string
	_, error_parse_command_type := fmt.Sscanf(command_string, "%d %s", &command_timestamp, &command_type)
	if error_parse_command_type != nil {
		latest_timestamp = command_timestamp
		return "Error parsing command type in command file!", 3
	}
	if command_timestamp == latest_timestamp {
		return "", 0
	}
	if command_timestamp >= latest_timestamp+2 || command_timestamp < latest_timestamp {
		latest_timestamp = command_timestamp
		return "Command timestamp jump by more than 1 once!", 3
	}

	if command_type == flush_result_command {
		latest_timestamp = command_timestamp
		return flush_result_token, 1
	} else if command_type == "query" {
		ans_str := query()
		latest_timestamp = command_timestamp
		return ans_str, 1
	} else if command_type == "tx" {
		var source, dest int
		var value, fee float32
		_, error_parse_tx_arguments := fmt.Sscanf(command_string, "%d %s %d %d %f %f",
			&command_timestamp, &command_type, &source, &dest, &value, &fee)
		if error_parse_tx_arguments != nil {
			latest_timestamp = command_timestamp
			return "Error parsing tx arguments in command file!", 3
		}
		ans_str := tx(source, dest, value, fee)
		latest_timestamp = command_timestamp
		return ans_str, 1
	} else if command_type == "adjust" {
		var log_difficulty int
		_, error_parse_adjust_arguments := fmt.Sscanf(command_string, "%d %s %d",
			&command_timestamp, &command_type, &log_difficulty)
		if error_parse_adjust_arguments != nil {
			latest_timestamp = command_timestamp
			return "Error parsing adjust arguments in command file!", 3
		}
		ans_str := adjust(log_difficulty)
		latest_timestamp = command_timestamp
		return ans_str, 1
	} else if command_type == "connect" {
		var source, dest int
		_, error_parse_connect_arguments := fmt.Sscanf(command_string, "%d %s %d %d",
			&command_timestamp, &command_type, &source, &dest)
		if error_parse_connect_arguments != nil {
			latest_timestamp = command_timestamp
			return "Error parsing connect arguments in command file!", 3
		}
		ans_str := connect(source, dest)
		latest_timestamp = command_timestamp
		return ans_str, 1
	} else if command_type == "newnode" {
		var index int
		_, error_parse_newnode_arguments := fmt.Sscanf(command_string, "%d %s %d",
			&command_timestamp, &command_type, &index)
		if error_parse_newnode_arguments != nil {
			latest_timestamp = command_timestamp
			return "Error parsing newnode arguments in command file!", 3
		}
		ans_str := newnode(index)
		latest_timestamp = command_timestamp
		return ans_str, 1
	} else {
		latest_timestamp = command_timestamp
		return "Unknown command type in command file!", 3
	}
}

// 先清空再写, 避免ending_symbol错误地生效
func writeResult(my_result string) {
	if my_result == flush_result_token {
		error_flush_result := os.WriteFile(result_file_path, []byte(empty_symbol), 0666)
		if error_flush_result != nil {
			fmt.Println("Error flushing result!")
			log.Fatal(error_flush_result)
		}
	} else {
		error_clear_result := os.WriteFile(result_file_path, []byte(""), 0666)
		if error_clear_result != nil {
			fmt.Println("Error clearing result!")
			log.Fatal(error_clear_result)
		}
		error_write_result := os.WriteFile(result_file_path, []byte(
			fmt.Sprintf("%d %s %s", latest_timestamp, my_result, ending_symbol)), 0666)
		if error_write_result != nil {
			fmt.Println("Error writing result!")
			log.Fatal(error_write_result)
		}
	}
}

// =========================================================================================================================
// ====================================================== 主函数 ============================================================
// =========================================================================================================================

func main() {
	my_ip_ptr := flag.String("my-ip", "10.1.0.99", "The IP address that client holds")
	port_ptr := flag.String("port", "8056", "port of connection between miners")
	port_client_ptr := flag.String("port-client", "8057", "port of connection between miners and client")
	flag.Parse()
	my_ip = *my_ip_ptr
	port = *port_ptr
	port_client = *port_client_ptr

	for index := range node_list {
		node_list[index].Port = port_client
	}

	// 初始化 rpc
	// initConnection(node_list)
	defer closeConnection()
	if len(connection_pool) != len(node_list) {
		log.Fatal("Connection not successful")
	}

	random_generator = rand.New(rand.NewSource(time.Now().UnixNano()))
	initLatestTimestamp()

	// 基于命令行界面的交互
	// 几种格式：
	//		query - 输出所有矿工的余额
	//		tx id1 id2 v f - 要求编号为 id1 的矿工给编号为 id2 的矿工转帐 v 的价值，交易费为 f
	//		adjust d - 将所有矿工的挖矿难度（对数）改为 d
	// 		connect id1 id2 - 创建一个将 id1 矿工到 id2 矿工的连接，此处如果有新节点加入，必须是 id1 为新节点！！
	//		newnode id - 将客户端连接到 id 矿工（需要远端已经启动）
	// 注：矿工的 id = 机器端口号 - 8051
	var file_async_error_tolerating_times uint = 0
	for {
		my_result, state := readCommandAndExecute()
		if state == 1 {
			file_async_error_tolerating_times = 0
			writeResult(my_result)
		} else if state == 2 {
			file_async_error_tolerating_times += 1
			if file_async_error_tolerating_times >= file_async_error_tolerating_timeout {
				fmt.Println("File async error tolerating timeouts. Maybe command file has fatal error")
				log.Fatal("File async error tolerating timeouts.")
			}
		} else if state == 3 {
			file_async_error_tolerating_times = 0
			writeResult(my_result)
			fmt.Println(my_result)
		}
		time.Sleep(time.Duration(100+random_generator.Intn(100)) * time.Millisecond)
	}
}
