package model

import (
	CommonModel "ServerCommon/model"
	//"ServerIMS/model"
	"net"
	"sync"
	"time"
)

//
type SyncClient struct {
	conn *net.TCPConn
	ewt  chan *CommonModel.Message
}

type Master struct {
	Ewt chan *EMessage

	Mutex   sync.Mutex
	Clients map[*SyncClient]struct{}
}

func (master *Master) AddClient(client *SyncClient) {
	master.Mutex.Lock()
	defer master.Mutex.Unlock()
	master.Clients[client] = struct{}{}
}

func (master *Master) RemoveClient(client *SyncClient) {
	master.Mutex.Lock()
	defer master.Mutex.Unlock()
	delete(master.Clients, client)
}

func (master *Master) CloneClientSet() map[*SyncClient]struct{} {
	master.Mutex.Lock()
	defer master.Mutex.Unlock()
	clone := make(map[*SyncClient]struct{})
	for k, v := range master.Clients {
		clone[k] = v
	}
	return clone
}

func (master *Master) SendBatch(cache []*EMessage) {
	if len(cache) == 0 {
		return
	}

	batch := &MessageBatch{Msgs: make([]*CommonModel.Message, 0, 1000)}
	batch.First_id = cache[0].Msgid
	for _, em := range cache {
		batch.Last_id = em.Msgid
		batch.Msgs = append(batch.Msgs, em.Msg)
	}
	m := &CommonModel.Message{Cmd: MSG_STORAGE_SYNC_MESSAGE_BATCH, Body: batch}
	clients := master.CloneClientSet()
	for c := range clients {
		c.ewt <- m
	}
}

func (master *Master) Run() {
	cache := make([]*EMessage, 0, 1000)
	var first_ts time.Time
	for {
		t := 60 * time.Second
		if len(cache) > 0 {
			ts := first_ts.Add(time.Second * 1)
			now := time.Now()

			if ts.After(now) {
				t = ts.Sub(now)
			} else {
				master.SendBatch(cache)
				cache = cache[0:0]
			}
		}
		select {
		case emsg := <-master.Ewt:
			cache = append(cache, emsg)
			if len(cache) == 1 {
				first_ts = time.Now()
			}
			if len(cache) >= 1000 {
				master.SendBatch(cache)
				cache = cache[0:0]
			}
		case <-time.After(t):
			if len(cache) > 0 {
				master.SendBatch(cache)
				cache = cache[0:0]
			}
		}
	}
}

func (master *Master) Start() {
	go master.Run()
}
