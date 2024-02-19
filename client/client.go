package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		Name:       fmt.Sprintf("%s:%d", serverIp, serverPort),
		flag:       999,
	}

	conn, err := net.Dial("tcp", client.Name)
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	return client
}

func (client *Client) menu() bool {
	var flag int

	fmt.Println("0.退出")
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>请输入合法范围内的数字<<<<<<")
		return false
	}
}

func (client *Client) UpdateName() bool {

	fmt.Println(">>>>>>请输入用户名:")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

func (client *Client) PrivateChat() {

	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println(">>>>>>请输入聊天对象[用户名],exit退出:")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>>请输入聊天内容,exit退出:")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>>>请输入聊天内容,exit退出.")
			fmt.Scanln(&chatMsg)

		}

		client.SelectUsers()
		fmt.Println(">>>>>>请输入聊天对象[用户名],exit退出:")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) PublicChat() {

	var chatMsg string

	fmt.Println(">>>>>>请输入聊天内容,exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {

		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>>>请输入聊天内容,exit退出.")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) Run() {
	fmt.Println(client.flag)
	for client.flag != 0 {
		for !client.menu() {
		}

		switch client.flag {
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			break
		}
	}
}
