package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

const command_file_path string = "temp/command.txt"
const result_file_path string = "temp/result.txt"
const file_async_error_tolerating_timeout uint = 50
const ending_symbol string = "#-end-#"
const empty_symbol string = "#-empty-#"
const flush_result_command string = "flush-result"

var random_generator *rand.Rand
var my_command string = ""
var is_flush_temp bool = false
var latest_command_string string = ""
var latest_result_string string = ""
var latest_timestamp uint64 = 0

func parseCommandline() {
	my_command_ptr := flag.String("command", "",
		"the command you ask the client to do; should be wrapped by quotes if containing spaces")
	is_flush_temp_ptr := flag.Bool("flush-temp", false,
		"flush the temp and ignore current --command argument")
	flag.Parse()
	my_command = *my_command_ptr
	is_flush_temp = *is_flush_temp_ptr

	if is_flush_temp {
		my_command = flush_result_command
	}
}

func checkEnding(s string) bool {
	l := strings.Split(s, " ")
	return l[len(l)-1] == ending_symbol
}

// 错误处理: 输出并报错. 因此不适合在等待循环中使用
func readFile(file_path string) string {
	content_bytes, error_read_content := os.ReadFile(file_path)
	if error_read_content != nil {
		fmt.Printf("Error reading file %s!", file_path)
		log.Fatal(error_read_content)
	}
	return string(content_bytes)
}

// 错误处理: 输出并报错. 因此不适合在等待循环中使用
func writeFile(file_path string, content_string string) {
	error_write_content := os.WriteFile(file_path, []byte(content_string), 0666)
	if error_write_content != nil {
		fmt.Printf("Error writing file %s!", file_path)
		log.Fatal(error_write_content)
	}
}

func readTimestamp(content_string string) uint64 {
	var timestamp_ans uint64 = 0
	if content_string == empty_symbol { // 由deploy.py初始化时生成的
		timestamp_ans = 0
	} else {
		_, error_parse_latest_timestamp := fmt.Sscanf(content_string, "%d", &timestamp_ans)
		if error_parse_latest_timestamp != nil {
			fmt.Printf("Error parsing latest timestamp in string %s!", content_string)
			log.Fatal(error_parse_latest_timestamp)
		}
	}
	return timestamp_ans
}

func initLatestContents() {
	latest_command_string = readFile(command_file_path)
	latest_result_string = readFile(result_file_path)
	latest_command_timestamp := readTimestamp(latest_command_string)
	latest_result_timestamp := readTimestamp(latest_result_string)
	if latest_command_timestamp != latest_result_timestamp {
		fmt.Printf("command and result file have different timestamp!")
		log.Fatal("command and result file have different timestamp!")
	}
	latest_timestamp = latest_command_timestamp
}

// 先清空再写, 避免ending_symbol错误地生效
func writeCommand() {
	writeFile(command_file_path, "")
	writeFile(command_file_path, fmt.Sprintf("%d %s %s", latest_timestamp+1, my_command, ending_symbol))
}

func rollbackCommand() {
	writeFile(command_file_path, latest_command_string)
}

// 第二个返回值的含义:
// 0 == 没有新命令 (也有可能是file async)
// 1 == 成功处理新命令
// 2 == 还在file async中
// 3 == 遇到错误
// 不能用前面的readFile和writeFile, 因为要传递错误
// helper最重要的是发生什么都要输出, 但报错需要统一传递到最后, 因为所有输出都会被GUI接住
// helper需要用Printf而不是Println输出, 因为最后不能有换行符, 否则怕影响GUI解析
// helper几乎一直关着
func readResult() (string, int) {
	result_bytes, error_read_result := os.ReadFile(result_file_path)
	if error_read_result != nil {
		return "", 2
	}
	result_string := string(result_bytes)
	if result_string == empty_symbol { // 由deploy.py初始化时生成的
		if is_flush_temp {
			return "", 1
		} else {
			return "", 0
		}
	}
	if !checkEnding(result_string) {
		return "", 2
	}

	var result_timestamp uint64
	_, error_parse_result_timestamp := fmt.Sscanf(result_string, "%d", &result_timestamp)
	if error_parse_result_timestamp != nil {
		return "Error parsing timestamp in result file!", 3
	}
	if result_timestamp == latest_timestamp {
		return "", 0
	}
	if result_timestamp >= latest_timestamp+2 || result_timestamp < latest_timestamp {
		return "Result timestamp jump by more than 1 once!", 3
	}

	result_list := strings.Split(result_string, " ")
	pure_result_list := result_list[1 : len(result_list)-1]
	pure_result_string := strings.Join(pure_result_list, " ")
	return pure_result_string, 1
}

func flushTemp(file_path string) {
	error_flush := os.WriteFile(file_path, []byte(empty_symbol), 0666)
	if error_flush != nil {
		fmt.Printf("Error flushing file %s!", file_path)
		log.Fatal(error_flush)
	}
}

func postProcess(client_result string, state int) {
	if is_flush_temp {
		if state == 1 {
			fmt.Printf("Client flushes %s. ", result_file_path)
			flushTemp(command_file_path)
			fmt.Printf("Helper flushes %s. Successfully flushed temp.", command_file_path)
		} else if state == 2 {
			fmt.Printf("Client timeout. Maybe it is off. ")
			flushTemp(command_file_path)
			flushTemp(result_file_path)
			fmt.Printf("Helper flushes %s and %s. Successfully flushed temp.", command_file_path, result_file_path)
		} else if state == 3 {
			fmt.Printf("Client: %s", client_result)
			log.Fatal("Client:", client_result)
		} else {
			fmt.Printf("After reading result from client, reach unknown state: %d", state)
			log.Fatal("After reading result from client, reach unknown state:", state)
		}
	} else {
		if state == 1 {
			fmt.Printf("%s", client_result)
		} else if state == 2 {
			fmt.Printf("Client timeout. Maybe it is off. ")
			rollbackCommand()
			fmt.Printf("The command is rolled back.")
		} else if state == 3 {
			fmt.Printf("Client: %s", client_result)
			log.Fatal("Client:", client_result)
		} else {
			fmt.Printf("After reading result from client, reach unknown state: %d", state)
			log.Fatal("After reading result from client, reach unknown state:", state)
		}
	}
}

func main() {
	parseCommandline()
	initLatestContents()
	writeCommand()
	random_generator = rand.New(rand.NewSource(time.Now().UnixNano()))

	var file_async_error_tolerating_times uint = 0
	var client_result string = ""
	var state int = 0
	for {
		client_result, state = readResult()
		if state == 1 || state == 3 {
			break
		} else if state == 2 {
			file_async_error_tolerating_times += 1
			if file_async_error_tolerating_times >= file_async_error_tolerating_timeout {
				break
			}
		}
		time.Sleep(time.Duration(100+random_generator.Intn(100)) * time.Millisecond)
	}
	// 为避免: client写到一半挂了, 此时state应为3但实为2
	if state == 2 && readFile(result_file_path) != latest_result_string {
		state = 3
	}
	postProcess(client_result, state)
}
