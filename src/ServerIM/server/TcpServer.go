package server

import (
	//CommonModel "ServerCommon/model"
	"ServerIM/config"
	"ServerIM/model"
	"crypto/tls"
	"fmt"
	"net"

	log "github.com/golang/glog"
)

func ListenClient() {

	go StartHttpServer(config.Config.Http_listen_address)

	go StartSocketIO(config.Config.Socket_io_address, config.Config.Tls_address,
		config.Config.Cert_file, config.Config.Key_file)
	go StartWSServer(config.Config.Ws_address, config.Config.Wss_address,
		config.Config.Cert_file, config.Config.Key_file)

	if config.Config.Ssl_port > 0 && len(config.Config.Cert_file) > 0 &&
		len(config.Config.Key_file) > 0 {
		go ListenSSL(config.Config.Ssl_port, config.Config.Cert_file,
			config.Config.Key_file)
	}

	Listen(config.Config.Port)
}

func Listen(port int) {
	listen_addr := fmt.Sprintf("0.0.0.0:%d", port)
	listen, err := net.Listen("tcp", listen_addr)
	if err != nil {
		log.Errorf("listen err:%s", err)
		return
	}
	tcp_listener, ok := listen.(*net.TCPListener)
	if !ok {
		log.Error("listen err")
		return
	}

	for {
		client, err := tcp_listener.AcceptTCP()
		if err != nil {
			log.Errorf("accept err:%s", err)
			return
		}
		handle_client(client)
	}
}

func handle_client(conn net.Conn) {
	log.Infoln("handle new connection")
	client := model.NewClient(conn)
	client.Run()
}

func ListenSSL(port int, cert_file, key_file string) {
	cert, err := tls.LoadX509KeyPair(cert_file, key_file)
	if err != nil {
		log.Fatal("load cert err:", err)
		return
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	addr := fmt.Sprintf(":%d", port)
	listen, err := tls.Listen("tcp", addr, config)
	if err != nil {
		log.Fatal("ssl listen err:", err)
	}

	log.Infof("ssl listen...")
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal("ssl accept err:", err)
		}
		handle_ssl_client(conn)
	}
}

func handle_ssl_client(conn net.Conn) {
	log.Infoln("handle new ssl connection")
	client := model.NewClient(conn)
	client.Run()
}
