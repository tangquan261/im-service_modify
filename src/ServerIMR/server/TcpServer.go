package server

import (
	"ServerIMR/config"
	"ServerIMR/model"
	"fmt"
	"net"
	"time"

	log "github.com/golang/glog"
)

//作为转发服务器，等待IM连接，处理数据
func ListenClient() {
	Listen(config.Config.Listen)
}

func Listen(listen_addr string) {
	listen, err := net.Listen("tcp", listen_addr)
	if err != nil {
		fmt.Println("初始化失败", err.Error())
		return
	}
	tcp_listener, ok := listen.(*net.TCPListener)
	if !ok {
		fmt.Println("listen error")
		return
	}

	for {
		client, err := tcp_listener.AcceptTCP()
		if err != nil {
			return
		}
		//获得一个网络连接
		handle_client(client)
	}
}

func handle_client(conn *net.TCPConn) {
	conn.SetKeepAlive(true)

	//设置网络超时时间10分钟
	conn.SetKeepAlivePeriod(time.Duration(10 * 60 * time.Second))
	client := model.NewClient(conn)
	log.Info("new client:", conn.RemoteAddr())
	client.Run()
}
