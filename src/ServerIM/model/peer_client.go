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
	CommonModel "ServerCommon/model"
	CommonFilter "ServerCommon/pkg/filter"
	"ServerIM/config"
	"sync/atomic"
	"time"

	log "github.com/golang/glog"
)

type PeerClient struct {
	*Connection
}

func (client *PeerClient) Login() {
	channel := GetChannel(client.uid)

	channel.Subscribe(client.appid, client.uid, client.online)

	for _, c := range Group_route_channels {
		if c == channel {
			continue
		}

		c.Subscribe(client.appid, client.uid, client.online)
	}

	SetUserUnreadCount(client.appid, client.uid, 0)
}

func (client *PeerClient) Logout() {
	if client.uid > 0 {
		channel := GetChannel(client.uid)
		channel.Unsubscribe(client.appid, client.uid, client.online)

		for _, c := range Group_route_channels {
			if c == channel {
				continue
			}

			c.Unsubscribe(client.appid, client.uid, client.online)
		}
	}
}

func (client *PeerClient) HandleSync(sync_key *CommonModel.SyncKey) {
	if client.uid == 0 {
		return
	}
	last_id := sync_key.Sync_key

	if last_id == 0 {
		last_id = GetSyncKey(client.appid, client.uid)
	}

	rpc := GetStorageRPCClient(client.uid)

	s := &SyncHistory{
		AppID:     client.appid,
		Uid:       client.uid,
		DeviceID:  client.device_ID,
		LastMsgID: last_id,
	}

	log.Infof("syncing message:%d %d %d %d", client.appid, client.uid, client.device_ID, last_id)

	resp, err := rpc.Call("SyncMessage", s)
	if err != nil {
		log.Warning("sync message err:", err)
		return
	}
	client.sync_count += 1

	ph := resp.(*PeerHistoryMessage)
	messages := ph.Messages

	msgs := make([]*CommonModel.Message, 0, len(messages)+2)

	sk := &CommonModel.SyncKey{last_id}
	msgs = append(msgs, &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_BEGIN, Body: sk})

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		log.Info("message:", msg.MsgID, CommonModel.Command(msg.Cmd))
		m := &CommonModel.Message{Cmd: int(msg.Cmd), Version: CommonModel.DEFAULT_VERSION}
		m.FromData(msg.Raw)
		sk.Sync_key = msg.MsgID

		if config.Config.Sync_self {
			//连接成功后的首次同步，自己发送的消息也下发给客户端
			//之后的同步则过滤掉所有自己在当前设备发出的消息
			//这是为了解决服务端已经发出消息，但是对发送端的消息ack丢失的问题
			if client.sync_count > 1 && client.isSender(m, msg.DeviceID) {
				continue
			}
		} else {
			//过滤掉所有自己在当前设备发出的消息
			if client.isSender(m, msg.DeviceID) {
				continue
			}
		}
		if client.isSender(m, msg.DeviceID) {
			m.Flag |= CommonModel.MESSAGE_FLAG_SELF
		}
		msgs = append(msgs, m)
	}

	if ph.LastMsgID < last_id && ph.LastMsgID > 0 {
		sk.Sync_key = ph.LastMsgID
		log.Warningf("client last id:%d server last id:%d", last_id, ph.LastMsgID)
	}

	msgs = append(msgs, &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_END, Body: sk})

	client.EnqueueMessages(msgs)

	if ph.HasMore {
		notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
			Body: &CommonModel.SyncKey{ph.LastMsgID + 1}}
		client.EnqueueMessage(notify)
	}
}

func (client *PeerClient) HandleSyncKey(sync_key *CommonModel.SyncKey) {
	if client.uid == 0 {
		return
	}

	last_id := sync_key.Sync_key
	log.Infof("sync key:%d %d %d %d", client.appid, client.uid, client.device_ID, last_id)
	if last_id > 0 {
		s := &SyncHistory{
			AppID:     client.appid,
			Uid:       client.uid,
			LastMsgID: last_id,
		}
		Sync_c <- s
	}
}

func (client *PeerClient) HandleIMMessage(message *CommonModel.Message) {
	msg := message.Body.(*CommonModel.IMMessage)
	seq := message.Seq
	if client.uid == 0 {
		log.Warning("client has't been authenticated")
		return
	}

	if msg.Sender != client.uid {
		log.Warningf("im message sender:%d client uid:%d\n", msg.Sender, client.uid)
		return
	}
	if message.Flag&CommonModel.MESSAGE_FLAG_TEXT != 0 {
		CommonFilter.FilterDirtyWord(msg)
	}
	msg.Timestamp = int32(time.Now().Unix())
	m := &CommonModel.Message{Cmd: CommonModel.MSG_IM,
		Version: CommonModel.DEFAULT_VERSION, Body: msg}

	msgid, err := SaveMessage(client.appid, msg.Receiver, client.device_ID, m)
	if err != nil {
		log.Errorf("save peer message:%d %d err:", msg.Sender, msg.Receiver, err)
		return
	}

	//保存到自己的消息队列，这样用户的其它登陆点也能接受到自己发出的消息
	msgid2, err := SaveMessage(client.appid, msg.Sender, client.device_ID, m)
	if err != nil {
		log.Errorf("save peer message:%d %d err:", msg.Sender, msg.Receiver, err)
		return
	}

	//推送外部通知
	PushMessage(client.appid, msg.Receiver, m)

	//发送同步的通知消息
	notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{msgid}}
	client.SendMessage(msg.Receiver, notify)

	//发送给自己的其它登录点
	notify = &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{msgid2}}
	client.SendMessage(client.uid, notify)

	ack := &CommonModel.Message{Cmd: CommonModel.MSG_ACK,
		Body: &CommonModel.MessageACK{int32(seq)}}
	r := client.EnqueueMessage(ack)
	if !r {
		log.Warning("send peer message ack error")
	}

	atomic.AddInt64(&Server_summary.In_message_count, 1)
	log.Infof("peer message sender:%d receiver:%d msgid:%d\n", msg.Sender, msg.Receiver, msgid)
}

func (client *PeerClient) HandleUnreadCount(u *CommonModel.MessageUnreadCount) {
	SetUserUnreadCount(client.appid, client.uid, u.Count)
}

func (client *PeerClient) HandleRTMessage(msg *CommonModel.Message) {
	rt := msg.Body.(*CommonModel.RTMessage)
	if rt.Sender != client.uid {
		log.Warningf("rt message sender:%d client uid:%d\n", rt.Sender, client.uid)
		return
	}

	m := &CommonModel.Message{Cmd: CommonModel.MSG_RT, Body: rt}
	client.SendMessage(rt.Receiver, m)

	atomic.AddInt64(&Server_summary.In_message_count, 1)
	log.Infof("realtime message sender:%d receiver:%d", rt.Sender, rt.Receiver)
}

func (client *PeerClient) HandleMessage(msg *CommonModel.Message) {
	switch msg.Cmd {
	case CommonModel.MSG_IM:
		client.HandleIMMessage(msg)
	case CommonModel.MSG_RT:
		client.HandleRTMessage(msg)
	case CommonModel.MSG_UNREAD_COUNT:
		client.HandleUnreadCount(msg.Body.(*CommonModel.MessageUnreadCount))
	case CommonModel.MSG_SYNC:
		client.HandleSync(msg.Body.(*CommonModel.SyncKey))
	case CommonModel.MSG_SYNC_KEY:
		client.HandleSyncKey(msg.Body.(*CommonModel.SyncKey))
	}
}
