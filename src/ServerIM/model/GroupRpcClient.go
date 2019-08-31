package model

import (
	"github.com/valyala/gorpc"
)

//super group storage server
var Group_rpc_clients []*gorpc.DispatcherClient

//storage server,  peer, group, customer message
var Rpc_clients []*gorpc.DispatcherClient
