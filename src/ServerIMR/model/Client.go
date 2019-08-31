/**
 * Copyright (c) 2014-2015, GoBelieve
 * All rights reserved.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA  02111-1307  USA
 */

package model

import (
	"ServerCommon/model"

	"ServerIMR/config"
	"net"

	CommonModel "ServerCommon/model"

	log "github.com/golang/glog"
)

type Push struct {
	queue_name string
	content    []byte
}

type Client struct {
	//message chan，一般用户对在线用户
	wt chan *model.Message

	//push chan，一般用于对离线用户,push消息通过redis发送出去
	pwt chan *Push

	conn      *net.TCPConn
	app_route *model.AppRoute
}

func (client *Client) Run() {
	//创建三个协程，分别读写，push
	go client.Write()
	go client.Read()
	go client.Push()
}

func NewClient(conn *net.TCPConn) *Client {
	client := new(Client)
	client.conn = conn
	client.pwt = make(chan *Push, 10000)
	client.wt = make(chan *model.Message, 10)
	client.app_route = model.NewAppRoute()
	client.app_route.FnCreate = NewRoute
	return client
}

func (client *Client) GetAppRoute() *model.AppRoute {
	return client.app_route
}

func (client *Client) ContainAppUserID(id *model.AppUserID) bool {
	route := client.app_route.FindRoute(id.Appid)
	if route == nil {
		return false
	}

	return route.ContainUserID(id.Uid)
}

func (client *Client) IsAppUserOnline(id *model.AppUserID) bool {
	route := client.app_route.FindRoute(id.Appid)
	if route == nil {
		return false
	}

	return route.IsUserOnline(id.Uid)
}

func (client *Client) ContainAppRoomID(id *model.AppRoomID) bool {
	route := client.app_route.FindRoute(id.Appid)
	if route == nil {
		return false
	}

	return route.ContainRoomID(id.Room_id)
}

func (client *Client) HandleMessage(msg *model.Message) {
	log.Info("msg cmd:", model.Command(msg.Cmd))
	switch msg.Cmd {
	case model.MSG_SUBSCRIBE:
		client.HandleSubscribe(msg.Body.(*model.SubscribeMessage))
	case model.MSG_UNSUBSCRIBE:
		client.HandleUnsubscribe(msg.Body.(*model.AppUserID))
	case model.MSG_PUBLISH:
		client.HandlePublish(msg.Body.(*model.AppMessage))
	case model.MSG_PUBLISH_GROUP:
		client.HandlePublishGroup(msg.Body.(*model.AppMessage))
	case model.MSG_SUBSCRIBE_ROOM:
		client.HandleSubscribeRoom(msg.Body.(*model.AppRoomID))
	case model.MSG_UNSUBSCRIBE_ROOM:
		client.HandleUnsubscribeRoom(msg.Body.(*model.AppRoomID))
	case model.MSG_PUBLISH_ROOM:
		client.HandlePublishRoom(msg.Body.(*model.AppMessage))
	default:
		log.Warning("unknown message cmd:", msg.Cmd)
	}
}

//在线请求
func (client *Client) HandleSubscribe(id *model.SubscribeMessage) {
	log.Infof("subscribe appid:%d uid:%d online:%d", id.Appid, id.Uid, id.Online)
	route := client.app_route.FindOrAddRoute(id.Appid)
	on := id.Online != 0
	route.AddUserID(id.Uid, on)
}

//离线请求
func (client *Client) HandleUnsubscribe(id *model.AppUserID) {
	log.Infof("unsubscribe appid:%d uid:%d", id.Appid, id.Uid)
	route := client.app_route.FindOrAddRoute(id.Appid)
	route.RemoveUserID(id.Uid)
}

//群聊消息广播
func (client *Client) HandlePublishGroup(amsg *model.AppMessage) {
	log.Infof("publish message appid:%d uid:%d msgid:%d cmd:%s", amsg.Appid, amsg.Receiver, amsg.Msgid, model.Command(amsg.Msg.Cmd))
	gid := amsg.Receiver
	group := CommonModel.Group_manager.FindGroup(gid)

	if group != nil && amsg.Msg.Cmd == model.MSG_GROUP_IM {
		msg := amsg.Msg
		members := group.Members()
		im := msg.Body.(*model.IMMessage)
		off_members := make([]int64, 0)
		for uid, _ := range members {
			if im.Sender != uid && !IsUserOnline(amsg.Appid, uid) {
				//不是自己同时不在线
				off_members = append(off_members, uid)
			}
		}
		if len(off_members) > 0 {
			//广播离线的用户状态push
			client.PublishGroupMessage(amsg.Appid, off_members, im)
		}
	}

	//当前只有MSG_SYNC_GROUP_NOTIFY可以发给终端
	if amsg.Msg.Cmd != model.MSG_SYNC_GROUP_NOTIFY {
		return
	}

	//群发给所有接入服务器
	s := GetClientSet()

	msg := &model.Message{Cmd: model.MSG_PUBLISH_GROUP, Body: amsg}
	for c := range s {
		//不发送给自身
		if client == c {
			continue
		}
		c.wt <- msg
	}
}

func (client *Client) HandlePublish(amsg *model.AppMessage) {
	log.Infof("publish message appid:%d uid:%d msgid:%d cmd:%s", amsg.Appid, amsg.Receiver, amsg.Msgid, model.Command(amsg.Msg.Cmd))

	cmd := amsg.Msg.Cmd
	receiver := &model.AppUserID{Appid: amsg.Appid, Uid: amsg.Receiver}
	s := FindClientSet(receiver)

	offline := true
	for c := range s {
		if c.IsAppUserOnline(receiver) {
			offline = false
		}
	}

	if offline {
		//用户不在线,推送消息到终端
		if cmd == model.MSG_IM {
			client.PublishPeerMessage(amsg.Appid, amsg.Msg.Body.(*model.IMMessage))
		} else if cmd == model.MSG_GROUP_IM {
			client.PublishGroupMessage(amsg.Appid, []int64{amsg.Receiver},
				amsg.Msg.Body.(*model.IMMessage))
		} else if cmd == model.MSG_CUSTOMER ||
			cmd == model.MSG_CUSTOMER_SUPPORT {
			client.PublishCustomerMessage(amsg.Appid, amsg.Receiver,
				amsg.Msg.Body.(*model.CustomerMessage), amsg.Msg.Cmd)
		} else if cmd == model.MSG_SYSTEM {
			sys := amsg.Msg.Body.(*model.SystemMessage)
			if config.Config.Is_push_system {
				client.PublishSystemMessage(amsg.Appid, amsg.Receiver, sys.Notification)
			}
		}
	}

	if cmd == model.MSG_IM || cmd == model.MSG_GROUP_IM ||
		cmd == model.MSG_CUSTOMER || cmd == model.MSG_CUSTOMER_SUPPORT ||
		cmd == model.MSG_SYSTEM {
		if amsg.Msg.Flag&model.MESSAGE_FLAG_UNPERSISTENT == 0 {
			//持久化的消息不主动推送消息到客户端
			return
		}
	}

	msg := &model.Message{Cmd: model.MSG_PUBLISH, Body: amsg}
	for c := range s {
		//不发送给自身
		if client == c {
			continue
		}
		c.wt <- msg
	}
}

func (client *Client) HandleSubscribeRoom(id *model.AppRoomID) {
	log.Infof("subscribe appid:%d room id:%d", id.Appid, id.Room_id)
	route := client.app_route.FindOrAddRoute(id.Appid)
	route.AddRoomID(id.Room_id)
}

func (client *Client) HandleUnsubscribeRoom(id *model.AppRoomID) {
	log.Infof("unsubscribe appid:%d room id:%d", id.Appid, id.Room_id)
	route := client.app_route.FindOrAddRoute(id.Appid)
	route.RemoveRoomID(id.Room_id)
}

func (client *Client) HandlePublishRoom(amsg *model.AppMessage) {
	log.Infof("publish room message appid:%d room id:%d cmd:%s", amsg.Appid, amsg.Receiver, model.Command(amsg.Msg.Cmd))
	receiver := &model.AppRoomID{Appid: amsg.Appid, Room_id: amsg.Receiver}
	s := FindRoomClientSet(receiver)

	msg := &model.Message{Cmd: model.MSG_PUBLISH_ROOM, Body: amsg}
	for c := range s {
		//不发送给自身
		if client == c {
			continue
		}
		log.Info("publish room message")
		c.wt <- msg
	}
}

func (client *Client) Read() {
	AddClient(client)
	for {
		msg := client.read()
		if msg == nil {
			RemoveClient(client)
			client.pwt <- nil
			client.wt <- nil
			break
		}
		client.HandleMessage(msg)
	}
}
func (client *Client) Write() {
	seq := 0
	for {
		msg := <-client.wt
		if msg == nil {
			client.close()
			log.Infof("client socket closed")
			break
		}
		seq++
		msg.Seq = seq
		client.send(msg)
	}
}

func (client *Client) read() *model.Message {
	return model.ReceiveMessage(client.conn)
}

func (client *Client) send(msg *model.Message) {
	model.SendMessage(client.conn, msg)
}

func (client *Client) close() {
	client.conn.Close()
}
