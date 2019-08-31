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

import log "github.com/golang/glog"
import "unsafe"
import "sync/atomic"
import "ServerCommon/model"

type RoomClient struct {
	*Connection
	room_id int64
}

func (client *RoomClient) Logout() {
	if client.room_id > 0 {
		channel := GetRoomChannel(client.room_id)
		channel.UnsubscribeRoom(client.appid, client.room_id)
		route := App_route.FindOrAddRoute(client.appid).(*Route)
		route.RemoveRoomClient(client.room_id, client.Client())
	}
}

func (client *RoomClient) HandleMessage(msg *model.Message) {
	switch msg.Cmd {
	case model.MSG_ENTER_ROOM:
		client.HandleEnterRoom(msg.Body.(*model.Room))
	case model.MSG_LEAVE_ROOM:
		client.HandleLeaveRoom(msg.Body.(*model.Room))
	case model.MSG_ROOM_IM:
		client.HandleRoomIM(msg.Body.(*model.RoomMessage), msg.Seq)
	}
}

func (client *RoomClient) HandleEnterRoom(room *model.Room) {
	if client.uid == 0 {
		log.Warning("client has't been authenticated")
		return
	}

	room_id := room.RoomID()
	log.Info("enter room id:", room_id)
	if room_id == 0 || client.room_id == room_id {
		return
	}
	route := App_route.FindOrAddRoute(client.appid).(*Route)
	if client.room_id > 0 {
		channel := GetRoomChannel(client.room_id)
		channel.UnsubscribeRoom(client.appid, client.room_id)

		route.RemoveRoomClient(client.room_id, client.Client())
	}

	client.room_id = room_id
	route.AddRoomClient(client.room_id, client.Client())
	channel := GetRoomChannel(client.room_id)
	channel.SubscribeRoom(client.appid, client.room_id)
}

func (client *RoomClient) Client() *Client {
	p := unsafe.Pointer(client.Connection)
	return (*Client)(p)
}

func (client *RoomClient) HandleLeaveRoom(room *model.Room) {
	if client.uid == 0 {
		log.Warning("client has't been authenticated")
		return
	}

	room_id := room.RoomID()
	log.Info("leave room id:", room_id)
	if room_id == 0 {
		return
	}
	if client.room_id != room_id {
		return
	}

	route := App_route.FindOrAddRoute(client.appid).(*Route)
	route.RemoveRoomClient(client.room_id, client.Client())
	channel := GetRoomChannel(client.room_id)
	channel.UnsubscribeRoom(client.appid, client.room_id)
	client.room_id = 0
}

func (client *RoomClient) HandleRoomIM(room_im *model.RoomMessage, seq int) {
	if client.uid == 0 {
		log.Warning("client has't been authenticated")
		return
	}
	room_id := room_im.Receiver
	if room_id != client.room_id {
		log.Warningf("room id:%d is't client's room id:%d\n", room_id, client.room_id)
		return
	}

	fb := atomic.LoadInt32(&client.forbidden)
	if fb == 1 {
		log.Infof("room id:%d client:%d, %d is forbidden", room_id, client.appid, client.uid)
		return
	}

	m := &model.Message{Cmd: model.MSG_ROOM_IM, Body: room_im}
	route := App_route.FindOrAddRoute(client.appid).(*Route)
	clients := route.FindRoomClientSet(room_id)
	for c, _ := range clients {
		if c == client.Client() {
			continue
		}
		c.EnqueueNonBlockMessage(m)
	}

	amsg := &model.AppMessage{Appid: client.appid, Receiver: room_id, Msg: m}
	channel := GetRoomChannel(client.room_id)
	channel.PublishRoom(amsg)

	client.Wt <- &model.Message{Cmd: model.MSG_ACK, Body: &model.MessageACK{int32(seq)}}
}
