package main

import "net"

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 启动监听当前 user channel 消息的 goroutine
	go user.ListenerMessage()

	return user
}

// 监听当前 User channel 的方法, 一旦有消息, 就直接发送给对端客户端
func (this *User) ListenerMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线功能
func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	// 广播当前用户上线消息
	this.server.Broadcast(this, "已上线")
}

// 用户下线功能
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	this.server.Broadcast(this, "下线")
}

// 给当前 user 对应的客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 处理广播信息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + this.Addr + "]" + this.Name + ":" + "在线...\n"
			user.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else {
		this.server.Broadcast(this, msg)
	}
}
