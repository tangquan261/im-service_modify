package server

//提供给IM链接，只用户来IM同步消息使用（消息来源自GRPC）
import (
	"ServerIMS/model"
	"fmt"
	"net"
	"time"
)

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
		handle_sync_client(client)
	}
}

//该长连接主要用来
func handle_sync_client(conn *net.TCPConn) {
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Duration(10 * 60 * time.Second))
	client := model.NewSyncClient(conn)
	client.Run()
}
