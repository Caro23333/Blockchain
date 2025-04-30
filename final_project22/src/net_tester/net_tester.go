package main

import (
	"flag"
	"fmt"
	"net"
)

func isAvailable(test_port string) bool {
	listener, error_listen := net.Listen("tcp", ":"+test_port)
	if error_listen == nil {
		listener.Close()
		return true
	} else {
		return false
	}
}

func main() {
	port_ptr := flag.String("port", "8056", "port of connection between miners")
	port_client_ptr := flag.String("port-client", "8057", "port of connection between miners and client")
	flag.Parse()
	port := *port_ptr
	port_client := *port_client_ptr

	if isAvailable(port) && isAvailable(port_client) {
		fmt.Printf("available")
	} else {
		fmt.Printf("unavailable")
	}
}
