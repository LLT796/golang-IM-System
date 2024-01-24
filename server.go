package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	// 消息广播的 channel
	Message chan string
}

// 构造 Server 对象
func newServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 启动服务器的接口
func (this *Server) Start() {
	// 1.socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net listen err", err)
		return
	}
	// 4.close listener
	defer listener.Close()

	// 启动监听 Message 的 goroutine
	go this.ListenMessage()

	for {
		// 2.accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err")
			continue
		}
		// 3.do handler
		go this.Handler(conn)
	}
}

// 广播消息的方法
func (this *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

// 监听 Message 广播消息 channel 的 goroutine, 一旦有消息就发送给全部在线的 user
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		// 将 msg 发送给全部的在线的 user
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) Handler(conn net.Conn) {
	// ..处理业务逻辑
	// fmt.Println("连接成功! ")

	user := NewUser(conn)
	// 用户上线，将用户加入到 OnlineMap 中
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	// 广播当前用户上线消息
	this.Broadcast(user, "已上线")

	// 接收客户端发送的数据
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				this.Broadcast(user, "下线")
				return
			}
			if err != nil {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取用户的消息(去除'\n')
			msg := string(buf[:n-1])

			// 将得到的消息进行广播
			this.Broadcast(user, msg)
		}
	}()

	// 当前 handler 阻塞
	select {}
}
