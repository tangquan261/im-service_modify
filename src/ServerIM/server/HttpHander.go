package server

import (
	CommonModel "ServerCommon/model"
	"ServerIM/model"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/golang/glog"
	//"github.com/importcjj/sensitive"
	"ServerIM/config"

	"github.com/bitly/go-simplejson"
)

//http
func PostGroupNotification(w http.ResponseWriter, req *http.Request) {
	log.Info("post group notification")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		WriteHttpError(400, err.Error(), w)
		return
	}

	obj, err := simplejson.NewJson(body)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	appid, err := obj.Get("appid").Int64()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}
	group_id, err := obj.Get("group_id").Int64()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	notification, err := obj.Get("notification").String()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	members := CommonModel.NewIntSet()

	marray, err := obj.Get("members").Array()
	for _, m := range marray {
		if _, ok := m.(json.Number); ok {
			member, err := m.(json.Number).Int64()
			if err != nil {
				log.Info("error:", err)
				WriteHttpError(400, "invalid json format", w)
				return
			}
			members.Add(member)
		}
	}

	group := CommonModel.Group_manager.FindGroup(group_id)
	if group != nil {
		ms := group.Members()
		for m, _ := range ms {
			members.Add(m)
		}
	}

	if len(members) == 0 {
		WriteHttpError(400, "group no member", w)
		return
	}

	SendGroupNotification(appid, group_id, notification, members)

	log.Info("post group notification success:", members)
	w.WriteHeader(200)
}

func SendGroupNotification(appid int64, gid int64,
	notification string, members CommonModel.IntSet) {

	msg := &CommonModel.Message{Cmd: CommonModel.MSG_GROUP_NOTIFICATION,
		Body: &CommonModel.GroupNotification{notification}}

	for member := range members {
		msgid, err := model.SaveMessage(appid, member, 0, msg)
		if err != nil {
			break
		}

		//发送同步的通知消息
		notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
			Body: &CommonModel.SyncKey{msgid}}
		model.SendAppMessage(appid, member, notify)
	}
}

func PostIMMessage(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		WriteHttpError(400, err.Error(), w)
		return
	}

	m, _ := url.ParseQuery(req.URL.RawQuery)

	appid, err := strconv.ParseInt(m.Get("appid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	sender, err := strconv.ParseInt(m.Get("sender"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	var is_group bool
	msg_type := m.Get("class")
	if msg_type == "group" {
		is_group = true
	} else if msg_type == "peer" {
		is_group = false
	} else {
		log.Info("invalid message class")
		WriteHttpError(400, "invalid message class", w)
		return
	}

	obj, err := simplejson.NewJson(body)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	sender2, err := obj.Get("sender").Int64()
	if err == nil && sender == 0 {
		sender = sender2
	}

	receiver, err := obj.Get("receiver").Int64()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	content, err := obj.Get("content").String()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	im := &CommonModel.IMMessage{}
	im.Sender = sender
	im.Receiver = receiver
	im.Msgid = 0
	im.Timestamp = int32(time.Now().Unix())
	im.Content = content

	if is_group {
		SendGroupIMMessage(im, appid)
		log.Info("post group im message success")
	} else {
		SendIMMessage(im, appid)
		log.Info("post peer im message success")
	}
	w.WriteHeader(200)
}

func LoadLatestMessage(w http.ResponseWriter, req *http.Request) {
	log.Info("load latest message")
	m, _ := url.ParseQuery(req.URL.RawQuery)

	appid, err := strconv.ParseInt(m.Get("appid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	uid, err := strconv.ParseInt(m.Get("uid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	limit, err := strconv.ParseInt(m.Get("limit"), 10, 32)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}
	log.Infof("appid:%d uid:%d limit:%d", appid, uid, limit)

	rpc := model.GetStorageRPCClient(uid)

	s := &model.HistoryRequest{
		AppID: appid,
		Uid:   uid,
		Limit: int32(limit),
	}

	resp, err := rpc.Call("GetLatestMessage", s)
	if err != nil {
		log.Warning("get latest message err:", err)
		WriteHttpError(400, "internal error", w)
		return
	}

	hm := resp.([]*model.HistoryMessage)
	messages := make([]*model.EMessage, 0)
	for _, msg := range hm {
		m := &CommonModel.Message{Cmd: int(msg.Cmd), Version: CommonModel.DEFAULT_VERSION}
		m.FromData(msg.Raw)
		e := &model.EMessage{Msgid: msg.MsgID, Device_id: msg.DeviceID, Msg: m}
		messages = append(messages, e)
	}

	if len(messages) > 0 {
		//reverse
		size := len(messages)
		for i := 0; i < size/2; i++ {
			t := messages[i]
			messages[i] = messages[size-i-1]
			messages[size-i-1] = t
		}
	}

	msg_list := make([]map[string]interface{}, 0, len(messages))
	for _, emsg := range messages {
		if emsg.Msg.Cmd == CommonModel.MSG_IM ||
			emsg.Msg.Cmd == CommonModel.MSG_GROUP_IM {
			im := emsg.Msg.Body.(*CommonModel.IMMessage)

			obj := make(map[string]interface{})
			obj["content"] = im.Content
			obj["timestamp"] = im.Timestamp
			obj["sender"] = im.Sender
			obj["receiver"] = im.Receiver
			obj["command"] = emsg.Msg.Cmd
			obj["id"] = emsg.Msgid
			msg_list = append(msg_list, obj)

		} else if emsg.Msg.Cmd == CommonModel.MSG_CUSTOMER ||
			emsg.Msg.Cmd == CommonModel.MSG_CUSTOMER_SUPPORT {
			im := emsg.Msg.Body.(*CommonModel.CustomerMessage)

			obj := make(map[string]interface{})
			obj["content"] = im.Content
			obj["timestamp"] = im.Timestamp
			obj["customer_appid"] = im.Customer_appid
			obj["customer_id"] = im.Customer_id
			obj["store_id"] = im.Store_id
			obj["seller_id"] = im.Seller_id
			obj["command"] = emsg.Msg.Cmd
			obj["id"] = emsg.Msgid
			msg_list = append(msg_list, obj)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	obj := make(map[string]interface{})
	obj["data"] = msg_list
	b, _ := json.Marshal(obj)
	w.Write(b)
	log.Info("load latest message success")
}

func LoadHistoryMessage(w http.ResponseWriter, req *http.Request) {
	log.Info("load message")
	m, _ := url.ParseQuery(req.URL.RawQuery)

	appid, err := strconv.ParseInt(m.Get("appid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	uid, err := strconv.ParseInt(m.Get("uid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	msgid, err := strconv.ParseInt(m.Get("last_id"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	rpc := model.GetStorageRPCClient(uid)

	s := &model.SyncHistory{
		AppID:     appid,
		Uid:       uid,
		DeviceID:  0,
		LastMsgID: msgid,
	}

	resp, err := rpc.Call("SyncMessage", s)
	if err != nil {
		log.Warning("sync message err:", err)
		return
	}

	ph := resp.(*model.PeerHistoryMessage)
	messages := ph.Messages

	if len(messages) > 0 {
		//reverse
		size := len(messages)
		for i := 0; i < size/2; i++ {
			t := messages[i]
			messages[i] = messages[size-i-1]
			messages[size-i-1] = t
		}
	}

	msg_list := make([]map[string]interface{}, 0, len(messages))
	for _, emsg := range messages {
		msg := &CommonModel.Message{Cmd: int(emsg.Cmd), Version: CommonModel.DEFAULT_VERSION}
		msg.FromData(emsg.Raw)
		if msg.Cmd == CommonModel.MSG_IM ||
			msg.Cmd == CommonModel.MSG_GROUP_IM {
			im := msg.Body.(*CommonModel.IMMessage)

			obj := make(map[string]interface{})
			obj["content"] = im.Content
			obj["timestamp"] = im.Timestamp
			obj["sender"] = im.Sender
			obj["receiver"] = im.Receiver
			obj["command"] = emsg.Cmd
			obj["id"] = emsg.MsgID
			msg_list = append(msg_list, obj)

		} else if msg.Cmd == CommonModel.MSG_CUSTOMER ||
			msg.Cmd == CommonModel.MSG_CUSTOMER_SUPPORT {
			im := msg.Body.(*CommonModel.CustomerMessage)

			obj := make(map[string]interface{})
			obj["content"] = im.Content
			obj["timestamp"] = im.Timestamp
			obj["customer_appid"] = im.Customer_appid
			obj["customer_id"] = im.Customer_id
			obj["store_id"] = im.Store_id
			obj["seller_id"] = im.Seller_id
			obj["command"] = emsg.Cmd
			obj["id"] = emsg.MsgID
			msg_list = append(msg_list, obj)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	obj := make(map[string]interface{})
	obj["data"] = msg_list
	b, _ := json.Marshal(obj)
	w.Write(b)
	log.Info("load history message success")
}

func SendSystemMessage(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		WriteHttpError(400, err.Error(), w)
		return
	}

	m, _ := url.ParseQuery(req.URL.RawQuery)

	appid, err := strconv.ParseInt(m.Get("appid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	uid, err := strconv.ParseInt(m.Get("uid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}
	sys := &CommonModel.SystemMessage{string(body)}
	msg := &CommonModel.Message{Cmd: CommonModel.MSG_SYSTEM, Body: sys}

	msgid, err := model.SaveMessage(appid, uid, 0, msg)
	if err != nil {
		WriteHttpError(500, "internal server error", w)
		return
	}

	//推送通知
	model.PushMessage(appid, uid, msg)

	//发送同步的通知消息
	notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{msgid}}
	model.SendAppMessage(appid, uid, notify)

	w.WriteHeader(200)
}

func SendRoomMessage(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		WriteHttpError(400, err.Error(), w)
		return
	}

	m, _ := url.ParseQuery(req.URL.RawQuery)

	appid, err := strconv.ParseInt(m.Get("appid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	uid, err := strconv.ParseInt(m.Get("uid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}
	room_id, err := strconv.ParseInt(m.Get("room"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	room_im := &CommonModel.RoomMessage{new(CommonModel.RTMessage)}
	room_im.Sender = uid
	room_im.Receiver = room_id
	room_im.Content = string(body)

	msg := &CommonModel.Message{Cmd: CommonModel.MSG_ROOM_IM, Body: room_im}
	route := model.App_route.FindOrAddRoute(appid).(*model.Route)
	clients := route.FindRoomClientSet(room_id)
	for c, _ := range clients {
		c.Wt <- msg
	}

	amsg := &CommonModel.AppMessage{Appid: appid, Receiver: room_id, Msg: msg}
	channel := model.GetRoomChannel(room_id)
	channel.PublishRoom(amsg)

	w.WriteHeader(200)
}

func SendNotification(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		WriteHttpError(400, err.Error(), w)
		return
	}

	m, _ := url.ParseQuery(req.URL.RawQuery)

	appid, err := strconv.ParseInt(m.Get("appid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	uid, err := strconv.ParseInt(m.Get("uid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}
	sys := &CommonModel.SystemMessage{string(body)}
	msg := &CommonModel.Message{Cmd: CommonModel.MSG_NOTIFICATION, Body: sys}
	model.SendAppMessage(appid, uid, msg)

	w.WriteHeader(200)
}

func SendCustomerSupportMessage(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		WriteHttpError(400, err.Error(), w)
		return
	}

	obj, err := simplejson.NewJson(body)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	customer_appid, err := obj.Get("customer_appid").Int64()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	customer_id, err := obj.Get("customer_id").Int64()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	store_id, err := obj.Get("store_id").Int64()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	seller_id, err := obj.Get("seller_id").Int64()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	content, err := obj.Get("content").String()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	cm := &CommonModel.CustomerMessage{}
	cm.Customer_appid = customer_appid
	cm.Customer_id = customer_id
	cm.Store_id = store_id
	cm.Seller_id = seller_id
	cm.Content = content
	cm.Timestamp = int32(time.Now().Unix())

	m := &CommonModel.Message{Cmd: CommonModel.MSG_CUSTOMER_SUPPORT, Body: cm}

	msgid, err := model.SaveMessage(cm.Customer_appid, cm.Customer_id, 0, m)
	if err != nil {
		log.Warning("save message error:", err)
		WriteHttpError(500, "internal server error", w)
		return
	}

	msgid2, err := model.SaveMessage(config.Config.Kefu_appid, cm.Seller_id, 0, m)
	if err != nil {
		log.Warning("save message error:", err)
		WriteHttpError(500, "internal server error", w)
		return
	}

	model.PushMessage(cm.Customer_appid, cm.Customer_id, m)

	//发送给自己的其它登录点
	notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{msgid2}}
	model.SendAppMessage(config.Config.Kefu_appid, cm.Seller_id, notify)

	//发送同步的通知消息
	notify = &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{msgid}}
	model.SendAppMessage(cm.Customer_appid, cm.Customer_id, notify)

	w.WriteHeader(200)
}

func SendCustomerMessage(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		WriteHttpError(400, err.Error(), w)
		return
	}

	obj, err := simplejson.NewJson(body)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	customer_appid, err := obj.Get("customer_appid").Int64()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	customer_id, err := obj.Get("customer_id").Int64()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	store_id, err := obj.Get("store_id").Int64()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	seller_id, err := obj.Get("seller_id").Int64()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	content, err := obj.Get("content").String()
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid json format", w)
		return
	}

	cm := &CommonModel.CustomerMessage{}
	cm.Customer_appid = customer_appid
	cm.Customer_id = customer_id
	cm.Store_id = store_id
	cm.Seller_id = seller_id
	cm.Content = content
	cm.Timestamp = int32(time.Now().Unix())

	m := &CommonModel.Message{Cmd: CommonModel.MSG_CUSTOMER, Body: cm}

	msgid, err := model.SaveMessage(config.Config.Kefu_appid, cm.Seller_id, 0, m)
	if err != nil {
		log.Warning("save message error:", err)
		WriteHttpError(500, "internal server error", w)
		return
	}
	msgid2, err := model.SaveMessage(cm.Customer_appid, cm.Customer_id, 0, m)
	if err != nil {
		log.Warning("save message error:", err)
		WriteHttpError(500, "internal server error", w)
		return
	}

	model.PushMessage(config.Config.Kefu_appid, cm.Seller_id, m)

	//发送同步的通知消息
	notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{msgid}}
	model.SendAppMessage(config.Config.Kefu_appid, cm.Seller_id, notify)

	//发送给自己的其它登录点
	notify = &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{msgid2}}
	model.SendAppMessage(cm.Customer_appid, cm.Customer_id, notify)

	resp := make(map[string]interface{})
	resp["seller_id"] = seller_id
	WriteHttpObj(resp, w)
}
func SendRealtimeMessage(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		WriteHttpError(400, err.Error(), w)
		return
	}

	m, _ := url.ParseQuery(req.URL.RawQuery)

	appid, err := strconv.ParseInt(m.Get("appid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	sender, err := strconv.ParseInt(m.Get("sender"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}
	receiver, err := strconv.ParseInt(m.Get("receiver"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	rt := &CommonModel.RTMessage{}
	rt.Sender = sender
	rt.Receiver = receiver
	rt.Content = string(body)

	msg := &CommonModel.Message{Cmd: CommonModel.MSG_RT, Body: rt}
	model.SendAppMessage(appid, receiver, msg)
	w.WriteHeader(200)
}

func GetOfflineCount(w http.ResponseWriter, req *http.Request) {
	m, _ := url.ParseQuery(req.URL.RawQuery)

	appid, err := strconv.ParseInt(m.Get("appid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	uid, err := strconv.ParseInt(m.Get("uid"), 10, 64)
	if err != nil {
		log.Info("error:", err)
		WriteHttpError(400, "invalid query param", w)
		return
	}

	last_id, err := strconv.ParseInt(m.Get("sync_key"), 10, 64)
	if err != nil || last_id == 0 {
		last_id = model.GetSyncKey(appid, uid)
	}
	sync_key := model.SyncHistory{AppID: appid, Uid: uid, LastMsgID: last_id}

	dc := model.GetStorageRPCClient(uid)

	resp, err := dc.Call("GetNewCount", sync_key)

	if err != nil {
		log.Warning("get new count err:", err)
		WriteHttpError(500, "server internal error", w)
		return
	}
	count := resp.(int64)

	log.Infof("get offline appid:%d uid:%d sync_key:%d count:%d",
		appid, uid, last_id, count)
	obj := make(map[string]interface{})
	obj["count"] = count
	WriteHttpObj(obj, w)
}

func InitMessageQueue(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
}

func DequeueMessage(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
}
