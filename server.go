package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 监听Message广播消息，一旦有消息就广播
func (server *Server) ListenMessager() {
	for {
		msg := <-server.Message
		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

// 广播消息方法
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := fmt.Sprintf(`[%s]%s %s`, user.Addr, user.Name, msg)
	server.Message <- sendMsg
}

func (server *Server) onlineUser(user *User) {
	// 添加用户到onlineMap中
	server.mapLock.Lock()
	server.OnlineMap[user.Name] = user
	server.mapLock.Unlock()

	// 广播当前用户上线消息
	server.BroadCast(user, "已上线")
}

func (server *Server) offlineUser(user *User) {
	// 从onlineMap删除用户
	server.mapLock.Lock()
	delete(server.OnlineMap, user.Name)
	server.mapLock.Unlock()

	// 广播当前用户下线消息
	server.BroadCast(user, "下线")
}

func (server *Server) getOnlineUsers() string {
	server.mapLock.RLock()
	var onlineMsg string
	for _, user := range server.OnlineMap {
		onlineMsg += fmt.Sprintf(`[%s]%s 在线...`, user.Addr, user.Name) + "\n"
	}
	server.mapLock.RUnlock()
	return onlineMsg
}

func (server *Server) Handler(conn net.Conn) {

	user := NewUser(conn)

	server.onlineUser(user)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := user.conn.Read(buf)
			if n == 0 {
				server.offlineUser(user)
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取用户的消息(去掉'\n')
			msg := string(buf[:n-1])
			if msg == "who" {
				msg = server.getOnlineUsers()
				user.SendMsg(msg)
			} else {
				// 将获得消息进行广播
				server.BroadCast(user, msg)
			}

		}
	}()

}

// 启动server
func (server *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	// close listen socket
	defer listener.Close()

	// 启动监听Message
	go server.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		// do handler
		go server.Handler(conn)

	}

}
