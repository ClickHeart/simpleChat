package main

import (
	"flag"
	"fmt"
)

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器IP地址(默认是8888)")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>> 连接服务器失败...")
		return
	}

	go client.DealResponse()

	fmt.Println(">>>>>> 链接服务器成功...")

	client.Run()
}
