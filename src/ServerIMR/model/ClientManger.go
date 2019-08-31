package model

import (
	"ServerCommon/model"
	"sync"
)

//保存所有对应IM的服务器
var clients ClientSet
var mutex sync.Mutex

func init() {
	clients = NewClientSet()
}

func AddClient(client *Client) {
	mutex.Lock()
	defer mutex.Unlock()

	clients.Add(client)
}

func RemoveClient(client *Client) {
	mutex.Lock()
	defer mutex.Unlock()

	clients.Remove(client)
}

//clone clients
func GetClientSet() ClientSet {
	mutex.Lock()
	defer mutex.Unlock()

	s := NewClientSet()

	for c := range clients {
		s.Add(c)
	}
	return s
}

func FindClientSet(id *model.AppUserID) ClientSet {
	mutex.Lock()
	defer mutex.Unlock()

	s := NewClientSet()

	for c := range clients {
		if c.ContainAppUserID(id) {
			s.Add(c)
		}
	}
	return s
}

func FindRoomClientSet(id *model.AppRoomID) ClientSet {
	mutex.Lock()
	defer mutex.Unlock()

	s := NewClientSet()

	for c := range clients {
		if c.ContainAppRoomID(id) {
			s.Add(c)
		}
	}
	return s
}

func IsUserOnline(appid, uid int64) bool {
	mutex.Lock()
	defer mutex.Unlock()

	id := &model.AppUserID{Appid: appid, Uid: uid}

	for c := range clients {
		if c.IsAppUserOnline(id) {
			return true
		}
	}
	return false
}
