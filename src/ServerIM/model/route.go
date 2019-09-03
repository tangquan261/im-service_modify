package model

import (
	"sync"

	CommonModel "ServerCommon/model"

	log "github.com/golang/glog"
)

var App_route *CommonModel.AppRoute

type Route struct {
	appid        int64
	mutex        sync.Mutex
	clients      map[int64]ClientSet
	room_clients map[int64]ClientSet
}

func NewRoute(appid int64) CommonModel.RouteBase {
	route := new(Route)
	route.appid = appid
	route.clients = make(map[int64]ClientSet)
	route.room_clients = make(map[int64]ClientSet)
	return route
}

func (route *Route) AddRoomID(room_id int64) {

}
func (route *Route) AddUserID(uid int64, online bool) {

}
func (route *Route) ContainRoomID(room_id int64) bool {
	return false
}

func (route *Route) ContainUserID(room_id int64) bool {
	return false
}
func (route *Route) GetAppID() int64 {
	return 0
}
func (route *Route) RemoveUserID(uid int64) {

}
func (route *Route) RemoveRoomID(room_id int64) {

}
func (route *Route) IsUserOnline(uid int64) bool {
	return false
}

func (route *Route) AddRoomClient(room_id int64, client *Client) {
	route.mutex.Lock()
	defer route.mutex.Unlock()
	set, ok := route.room_clients[room_id]
	if !ok {
		set = NewClientSet()
		route.room_clients[room_id] = set
	}
	set.Add(client)
}

//todo optimise client set clone
func (route *Route) FindRoomClientSet(room_id int64) ClientSet {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	set, ok := route.room_clients[room_id]
	if ok {
		return set.Clone()
	} else {
		return nil
	}
}

func (route *Route) RemoveRoomClient(room_id int64, client *Client) bool {
	route.mutex.Lock()
	defer route.mutex.Unlock()
	if set, ok := route.room_clients[room_id]; ok {
		set.Remove(client)
		if set.Count() == 0 {
			delete(route.room_clients, room_id)
		}
		return true
	}
	log.Info("room client non exists")
	return false
}

func (route *Route) AddClient(client *Client) {
	route.mutex.Lock()
	defer route.mutex.Unlock()
	set, ok := route.clients[client.uid]
	if !ok {
		set = NewClientSet()
		route.clients[client.uid] = set
	}
	set.Add(client)
}

func (route *Route) RemoveClient(client *Client) bool {
	route.mutex.Lock()
	defer route.mutex.Unlock()
	if set, ok := route.clients[client.uid]; ok {
		set.Remove(client)
		if set.Count() == 0 {
			delete(route.clients, client.uid)
		}
		return true
	}
	log.Info("client non exists")
	return false
}

func (route *Route) FindClientSet(uid int64) ClientSet {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	set, ok := route.clients[uid]
	if ok {
		return set.Clone()
	} else {
		return nil
	}
}

func (route *Route) IsOnline(uid int64) bool {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	set, ok := route.clients[uid]
	if ok {
		return len(set) > 0
	}
	return false
}

func (route *Route) GetUserIDs() CommonModel.IntSet {
	return CommonModel.NewIntSet()
}
