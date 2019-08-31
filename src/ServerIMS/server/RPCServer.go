package server

import (
	log "github.com/golang/glog"
	"github.com/valyala/gorpc"
)

func ListenRPCClient(Rpc_listen string) {
	dispatcher := gorpc.NewDispatcher()
	dispatcher.AddFunc("SyncMessage", SyncMessage)           //同步历史消息
	dispatcher.AddFunc("SyncGroupMessage", SyncGroupMessage) //同步历史群聊消息
	dispatcher.AddFunc("SavePeerMessage", SavePeerMessage)   //保存私聊消息
	dispatcher.AddFunc("SaveGroupMessage", SaveGroupMessage) //保存群聊消息
	dispatcher.AddFunc("GetNewCount", GetNewCount)           //获取最新消息数
	dispatcher.AddFunc("GetLatestMessage", GetLatestMessage) //获取最近最新消息

	s := &gorpc.Server{
		Addr:    Rpc_listen,
		Handler: dispatcher.NewHandlerFunc(),
	}

	if err := s.Serve(); err != nil {
		log.Fatalf("Cannot start rpc server: %s", err)
	}
}
