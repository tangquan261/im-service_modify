package model

import (
	CommonModel "ServerCommon/model"
	"net"

	log "github.com/golang/glog"
)

//只负责获取数据的转发

func (client *SyncClient) Run() {
	go client.RunLoop()
}

func NewSyncClient(conn *net.TCPConn) *SyncClient {
	c := new(SyncClient)
	c.conn = conn
	c.ewt = make(chan *CommonModel.Message, 10)
	return c
}

func (client *SyncClient) RunLoop() {
	seq := 0
	msg := ReceiveMessage(client.conn)
	if msg == nil {
		return
	}
	if msg.Cmd != MSG_STORAGE_SYNC_BEGIN {
		return
	}

	cursor := msg.Body.(*SyncCursor)
	log.Info("cursor msgid:", cursor.Msgid)
	c := StorageObj.LoadSyncMessagesInBackground(cursor.Msgid)

	for batch := range c {
		msg := &CommonModel.Message{Cmd: MSG_STORAGE_SYNC_MESSAGE_BATCH, Body: batch}
		seq++
		msg.Seq = seq
		CommonModel.SendMessage(client.conn, msg)
	}

	MasterObj.AddClient(client)
	defer MasterObj.RemoveClient(client)

	for {
		//这里的ewt通过master控制获取数据，
		msg := <-client.ewt
		if msg == nil {
			log.Warning("chan closed")
			break
		}

		seq++
		msg.Seq = seq
		err := CommonModel.SendMessage(client.conn, msg)
		if err != nil {
			break
		}
	}
}
