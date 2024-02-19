package main

import (
	"fmt"
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

// 创建User
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	// 启动监听当前User channel消息的goroutine
	go user.ListenMessage()

	return user
}

// 监听当前User channel的方法，一旦有消息直接发送给客户端
func (user *User) ListenMessage() {
	for msg := range user.C {
		_, err := user.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println(msg)
			panic(err)
		}
	}
	err := user.conn.Close()
	if err != nil {
		panic(err)
	}
}

// 发送消息
func (user *User) SendMsg(msg string) {
	user.C <- msg
}
