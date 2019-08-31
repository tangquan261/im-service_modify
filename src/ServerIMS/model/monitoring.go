//监控统计请求
package model

var Server_summary *ServerSummary

func InitSummary() {
	Server_summary = NewServerSummary()
}

type ServerSummary struct {
	Nrequests           int64
	Peer_message_count  int64
	Group_message_count int64
}

func NewServerSummary() *ServerSummary {
	s := new(ServerSummary)
	return s
}
