package server

import (
	"ServerIM/config"
	"ServerIM/model"

	"github.com/valyala/gorpc"
)

func ListenRPCClient() {
	model.Rpc_clients = make([]*gorpc.DispatcherClient, 0)

	for _, addr := range config.Config.Storage_rpc_addrs {
		c := &gorpc.Client{
			Conns: 4,
			Addr:  addr,
		}
		c.Start()

		dispatcher := gorpc.NewDispatcher()
		dispatcher.AddFunc("SyncMessage", SyncMessageInterface)
		dispatcher.AddFunc("SyncGroupMessage", SyncGroupMessageInterface)
		dispatcher.AddFunc("SavePeerMessage", SavePeerMessageInterface)
		dispatcher.AddFunc("SaveGroupMessage", SaveGroupMessageInterface)
		dispatcher.AddFunc("GetLatestMessage", GetLatestMessageInterface)

		dc := dispatcher.NewFuncClient(c)

		model.Rpc_clients = append(model.Rpc_clients, dc)
	}

	if len(config.Config.Group_storage_rpc_addrs) > 0 {
		model.Group_rpc_clients = make([]*gorpc.DispatcherClient, 0)
		for _, addr := range config.Config.Group_storage_rpc_addrs {
			c := &gorpc.Client{
				Conns: 4,
				Addr:  addr,
			}
			c.Start()

			dispatcher := gorpc.NewDispatcher()
			dispatcher.AddFunc("SyncMessage", SyncMessageInterface)
			dispatcher.AddFunc("SyncGroupMessage", SyncGroupMessageInterface)
			dispatcher.AddFunc("SavePeerMessage", SavePeerMessageInterface)
			dispatcher.AddFunc("SaveGroupMessage", SaveGroupMessageInterface)

			dc := dispatcher.NewFuncClient(c)

			model.Group_rpc_clients = append(model.Group_rpc_clients, dc)
		}
	} else {
		model.Group_rpc_clients = model.Rpc_clients
	}
}
