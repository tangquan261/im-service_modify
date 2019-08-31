package model

type ServerSummary struct {
	Nconnections      int64
	Nclients          int64
	In_message_count  int64
	Out_message_count int64
}

//统计需要
var Server_summary *ServerSummary

func NewServerSummary() {
	Server_summary = new(ServerSummary)
}
