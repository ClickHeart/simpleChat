package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
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

func (server *Server) doMessage(msg string, user *User) {
	if msg == "who" {
		// 查询在线用户
		msg = server.getOnlineUsers()
		user.SendMsg(msg)
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 改名
		newName := msg[7:]
		_, ok := server.OnlineMap[newName]
		if ok {
			user.SendMsg("当前用户名已被使用\n")
		} else {
			server.mapLock.Lock()
			delete(server.OnlineMap, user.Name)
			server.OnlineMap[newName] = user
			server.mapLock.Unlock()

			user.Name = newName
			user.SendMsg("您已更新用户名：" + user.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式: to |张三|消息内容
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.SendMsg("消息格式不正确，请使用\"to|张三|你好呀\"")
			return
		}

		remoteUser, ok := server.OnlineMap[remoteName]
		if !ok {
			user.SendMsg("该用户不存在\n")
			return
		}

		content := strings.Split(msg, "|")[2]
		remoteUser.SendMsg(fmt.Sprintf("[私聊]%s: %s", user.Name, content))
	} else {
		// 将获得消息进行广播
		server.BroadCast(user, msg)
	}
}

func (server *Server) Handler(conn net.Conn) {

	user := NewUser(conn)

	server.onlineUser(user)

	isLive := make(chan bool)

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
			server.doMessage(msg, user)

			isLive <- true
		}
	}()

	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 60):
			user.SendMsg("你已超时离线了")
			close(user.C)
			return
		}
	}

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
