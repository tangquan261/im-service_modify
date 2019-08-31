package model

import (
	CommonModel "ServerCommon/model"

	//"ServerIMS/model"
	"net"
	"time"

	log "github.com/golang/glog"
)

//主从服务器，连接从服务
type Slaver struct {
	addr string
}

func NewSlaver(addr string) *Slaver {
	s := new(Slaver)
	s.addr = addr
	return s
}

func (slaver *Slaver) Run() {
	nsleep := 100
	for {
		conn, err := net.Dial("tcp", slaver.addr)
		if err != nil {
			log.Info("connect master server error:", err)
			nsleep *= 2
			if nsleep > 60*1000 {
				nsleep = 60 * 1000
			}
			log.Info("slaver sleep:", nsleep)
			time.Sleep(time.Duration(nsleep) * time.Millisecond)
			continue
		}
		tconn := conn.(*net.TCPConn)
		tconn.SetKeepAlive(true)
		tconn.SetKeepAlivePeriod(time.Duration(10 * 60 * time.Second))
		log.Info("slaver connected with master")
		nsleep = 100
		slaver.RunOnce(tconn)
	}
}

func (slaver *Slaver) RunOnce(conn *net.TCPConn) {
	defer conn.Close()

	seq := 0

	msgid := StorageObj.NextMessageID()
	cursor := &SyncCursor{msgid}
	log.Info("cursor msgid:", msgid)

	msg := &CommonModel.Message{Cmd: MSG_STORAGE_SYNC_BEGIN, Body: cursor}
	seq += 1
	msg.Seq = seq
	CommonModel.SendMessage(conn, msg)

	for {
		msg := CommonModel.ReceiveStorageSyncMessage(conn)
		if msg == nil {
			return
		}

		if msg.Cmd == MSG_STORAGE_SYNC_MESSAGE {
			emsg := msg.Body.(*EMessage)
			StorageObj.SaveSyncMessage(emsg)
		} else if msg.Cmd == MSG_STORAGE_SYNC_MESSAGE_BATCH {
			mb := msg.Body.(*MessageBatch)
			StorageObj.SaveSyncMessageBatch(mb)
		} else {
			log.Error("unknown message cmd:", CommonModel.Command(msg.Cmd))
		}
	}
}
