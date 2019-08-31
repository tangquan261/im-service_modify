package model

import (
	CommonModel "ServerCommon/model"
	"ServerIM/config"
	"sync/atomic"
	"time"

	"github.com/valyala/gorpc"

	log "github.com/golang/glog"
)

var current_deliver_index uint64

var Group_message_delivers []*GroupMessageDeliver

//route server
var Route_channels []*Channel

//super group route server
var Group_route_channels []*Channel

func InitConnection() {
	Route_channels = make([]*Channel, 0)
	for _, addr := range config.Config.Route_addrs {
		channel := NewChannel(addr, DispatchAppMessage, DispatchGroupMessage,
			DispatchRoomMessage)
		channel.Start()
		Route_channels = append(Route_channels, channel)
	}

	if len(config.Config.Group_route_addrs) > 0 {
		Group_route_channels = make([]*Channel, 0)
		for _, addr := range config.Config.Group_route_addrs {
			channel := NewChannel(addr, DispatchAppMessage, DispatchGroupMessage,
				DispatchRoomMessage)
			channel.Start()
			Group_route_channels = append(Group_route_channels, channel)
		}
	} else {
		Group_route_channels = Route_channels
	}

}

func SendAppGroupMessage(appid int64, group_id int64, msg *CommonModel.Message) {
	now := time.Now().UnixNano()
	amsg := &CommonModel.AppMessage{Appid: appid, Receiver: group_id, Msgid: 0,
		Timestamp: now, Msg: msg}
	channel := GetGroupChannel(group_id)
	channel.PublishGroup(amsg)
	DispatchGroupMessage(amsg)
}

func PublishGroupMessage(appid int64, group_id int64, msg *CommonModel.Message) {
	now := time.Now().UnixNano()

	amsg := &CommonModel.AppMessage{Appid: appid, Receiver: group_id,
		Msgid: 0, Timestamp: now, Msg: msg}
	channel := GetGroupChannel(group_id)
	channel.PublishGroup(amsg)
}

func GetGroupChannel(group_id int64) *Channel {
	if group_id < 0 {
		group_id = -group_id
	}
	index := group_id % int64(len(Group_route_channels))
	return Group_route_channels[index]
}

func DispatchGroupMessage(amsg *CommonModel.AppMessage) {
	now := time.Now().UnixNano()
	d := now - amsg.Timestamp
	log.Infof("dispatch group message:%s %d %d",
		CommonModel.Command(amsg.Msg.Cmd), amsg.Msg.Flag, d)
	if d > int64(time.Second) {
		log.Warning("dispatch group message slow...")
	}

	group := CommonModel.Group_manager.FindGroup(amsg.Receiver)
	if group == nil {
		log.Warningf("can't dispatch group message, appid:%d group id:%d",
			amsg.Appid, amsg.Receiver)
		return
	}

	route := App_route.FindRoute(amsg.Appid).(*Route)
	if route == nil {
		log.Warningf("can't dispatch app message, appid:%d uid:%d cmd:%s",
			amsg.Appid, amsg.Receiver, CommonModel.Command(amsg.Msg.Cmd))
		return
	}

	members := group.Members()
	for member := range members {
		clients := route.FindClientSet(member)
		if len(clients) == 0 {
			continue
		}

		for c, _ := range clients {
			c.EnqueueNonBlockMessage(amsg.Msg)
		}
	}
}

//超级群，离线消息推送
func PushGroupMessage(appid int64, group_id int64, m *CommonModel.Message) {
	now := time.Now().UnixNano()
	amsg := &CommonModel.AppMessage{Appid: appid, Receiver: group_id,
		Msgid: 0, Timestamp: now, Msg: m}

	group := CommonModel.Group_manager.FindGroup(amsg.Receiver)
	if group == nil {
		log.Warningf("can't dispatch group message, appid:%d group id:%d",
			amsg.Appid, amsg.Receiver)
		return
	}

	channels := make(map[*Channel]struct{})
	members := group.Members()
	for member := range members {
		channel := GetChannel(member)
		if _, ok := channels[channel]; !ok {
			channels[channel] = struct{}{}
		}
	}

	for channel, _ := range channels {
		channel.Publish(amsg)
	}
}

func GetChannel(uid int64) *Channel {
	if uid < 0 {
		uid = -uid
	}
	index := uid % int64(len(Route_channels))
	return Route_channels[index]
}

func GetRoomChannel(room_id int64) *Channel {
	if room_id < 0 {
		room_id = -room_id
	}
	index := room_id % int64(len(Route_channels))
	return Route_channels[index]
}

func DispatchAppMessage(amsg *CommonModel.AppMessage) {
	now := time.Now().UnixNano()
	d := now - amsg.Timestamp
	log.Infof("dispatch app message:%s %d %d", CommonModel.Command(amsg.Msg.Cmd),
		amsg.Msg.Flag, d)
	if d > int64(time.Second) {
		log.Warning("dispatch app message slow...")
	}

	route := App_route.FindRoute(amsg.Appid).(*Route)
	if route == nil {
		log.Warningf("can't dispatch app message, appid:%d uid:%d cmd:%s",
			amsg.Appid, amsg.Receiver, CommonModel.Command(amsg.Msg.Cmd))
		return
	}
	clients := route.FindClientSet(amsg.Receiver)
	if len(clients) == 0 {
		log.Infof("can't dispatch app message, appid:%d uid:%d cmd:%s",
			amsg.Appid, amsg.Receiver, CommonModel.Command(amsg.Msg.Cmd))
		return
	}
	for c, _ := range clients {
		c.EnqueueNonBlockMessage(amsg.Msg)
	}
}

func DispatchRoomMessage(amsg *CommonModel.AppMessage) {
	log.Info("dispatch room message", CommonModel.Command(amsg.Msg.Cmd))
	room_id := amsg.Receiver
	route := App_route.FindOrAddRoute(amsg.Appid).(*Route)
	clients := route.FindRoomClientSet(room_id)

	if len(clients) == 0 {
		log.Infof("can't dispatch room message, appid:%d room id:%d cmd:%s",
			amsg.Appid, amsg.Receiver, CommonModel.Command(amsg.Msg.Cmd))
		return
	}
	for c, _ := range clients {
		c.EnqueueNonBlockMessage(amsg.Msg)
	}
}

func GetGroupMessageDeliver(group_id int64) *GroupMessageDeliver {
	if group_id < 0 {
		group_id = -group_id
	}

	deliver_index := atomic.AddUint64(&current_deliver_index, 1)
	index := deliver_index % uint64(len(Group_message_delivers))
	return Group_message_delivers[index]
}

func SaveGroupMessage(appid int64, gid int64, device_id int64, msg *CommonModel.Message) (int64, error) {
	dc := GetGroupStorageRPCClient(gid)

	gm := &GroupMessage{
		AppID:    appid,
		GroupID:  gid,
		DeviceID: device_id,
		Cmd:      int32(msg.Cmd),
		Raw:      msg.ToData(),
	}
	resp, err := dc.Call("SaveGroupMessage", gm)
	if err != nil {
		log.Warning("save group message err:", err)
		return 0, err
	}
	msgid := resp.(int64)
	log.Infof("save group message:%d %d %d\n", appid, gid, msgid)
	return msgid, nil
}

func SaveMessage(appid int64, uid int64, device_id int64, m *CommonModel.Message) (int64, error) {
	dc := GetStorageRPCClient(uid)

	pm := &PeerMessage{
		AppID:    appid,
		Uid:      uid,
		DeviceID: device_id,
		Cmd:      int32(m.Cmd),
		Raw:      m.ToData(),
	}

	resp, err := dc.Call("SavePeerMessage", pm)
	if err != nil {
		log.Error("save peer message err:", err)
		return 0, err
	}

	msgid := resp.(int64)
	log.Infof("save peer message:%d %d %d %d\n", appid, uid, device_id, msgid)
	return msgid, nil
}

//离线消息推送
func PushMessage(appid int64, uid int64, m *CommonModel.Message) {
	PublishMessage(appid, uid, m)
}

func PublishMessage(appid int64, uid int64, m *CommonModel.Message) {
	now := time.Now().UnixNano()
	amsg := &CommonModel.AppMessage{Appid: appid, Receiver: uid, Msgid: 0,
		Timestamp: now, Msg: m}
	channel := GetChannel(uid)
	channel.Publish(amsg)
}

func SendAppMessage(appid int64, uid int64, msg *CommonModel.Message) {
	now := time.Now().UnixNano()
	amsg := &CommonModel.AppMessage{Appid: appid, Receiver: uid, Msgid: 0,
		Timestamp: now, Msg: msg}
	channel := GetChannel(uid)
	channel.Publish(amsg)
	DispatchAppMessage(amsg)
}

//超级群消息
func GetGroupStorageRPCClient(group_id int64) *gorpc.DispatcherClient {
	if group_id < 0 {
		group_id = -group_id
	}
	index := group_id % int64(len(Group_rpc_clients))
	return Group_rpc_clients[index]
}

//个人消息／普通群消息／客服消息
func GetStorageRPCClient(uid int64) *gorpc.DispatcherClient {
	if uid < 0 {
		uid = -uid
	}
	index := uid % int64(len(Rpc_clients))
	return Rpc_clients[index]
}
