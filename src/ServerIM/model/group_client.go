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
	"ServerIM/config"
	"sync/atomic"
	"time"

	CommonFilter "ServerCommon/pkg/filter"

	log "github.com/golang/glog"
)

var Group_sync_c chan *SyncGroupHistory

var Sync_c chan *SyncHistory

func init() {
	Group_sync_c = make(chan *SyncGroupHistory, 100)
	Sync_c = make(chan *SyncHistory, 100)
}

type GroupClient struct {
	*Connection
}

func (client *GroupClient) HandleSuperGroupMessage(msg *CommonModel.IMMessage) {
	m := &CommonModel.Message{Cmd: CommonModel.MSG_GROUP_IM,
		Version: CommonModel.DEFAULT_VERSION, Body: msg}
	msgid, err := SaveGroupMessage(client.appid, msg.Receiver, client.device_ID, m)
	if err != nil {
		log.Errorf("save group message:%d %d err:%s", msg.Sender, msg.Receiver, err)
		return
	}

	//推送外部通知
	PushGroupMessage(client.appid, msg.Receiver, m)

	//发送同步的通知消息
	notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_GROUP_NOTIFY,
		Body: &CommonModel.GroupSyncKey{Group_id: msg.Receiver, Sync_key: msgid}}
	client.SendGroupMessage(msg.Receiver, notify)
}

func (client *GroupClient) HandleGroupMessage(im *CommonModel.IMMessage,
	group *CommonModel.Group) {
	gm := &PendingGroupMessage{}
	gm.appid = client.appid
	gm.sender = im.Sender
	gm.device_ID = client.device_ID
	gm.gid = im.Receiver
	gm.timestamp = im.Timestamp

	members := group.Members()
	gm.members = make([]int64, len(members))
	i := 0
	for uid := range members {
		gm.members[i] = uid
		i += 1
	}

	gm.content = im.Content
	deliver := GetGroupMessageDeliver(group.Gid)
	m := &CommonModel.Message{Cmd: MSG_PENDING_GROUP_MESSAGE, Body: gm}
	deliver.SaveMessage(m)
}

func (client *GroupClient) HandleGroupIMMessage(message *CommonModel.Message) {
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

	group := CommonModel.Group_manager.FindGroup(msg.Receiver)
	if group == nil {
		log.Warning("can't find group:", msg.Receiver)
		return
	}

	if !group.IsMember(msg.Sender) {
		log.Warningf("sender:%d is not group member", msg.Sender)
		return
	}

	if group.GetMemberMute(msg.Sender) {
		log.Warningf("sender:%d is mute in group", msg.Sender)
		return
	}

	if group.Super {
		client.HandleSuperGroupMessage(msg)
	} else {
		client.HandleGroupMessage(msg, group)
	}
	ack := &CommonModel.Message{Cmd: CommonModel.MSG_ACK,
		Body: &CommonModel.MessageACK{int32(seq)}}
	r := client.EnqueueMessage(ack)
	if !r {
		log.Warning("send group message ack error")
	}

	atomic.AddInt64(&Server_summary.In_message_count, 1)
	log.Infof("group message sender:%d group id:%d", msg.Sender, msg.Receiver)
}

func (client *GroupClient) HandleGroupSync(group_sync_key *CommonModel.GroupSyncKey) {
	if client.uid == 0 {
		return
	}

	group_id := group_sync_key.Group_id

	group := CommonModel.Group_manager.FindGroup(group_id)
	if group == nil {
		log.Warning("can't find group:", group_id)
		return
	}

	if !group.IsMember(client.uid) {
		log.Warningf("sender:%d is not group member", client.uid)
		return
	}

	ts := group.GetMemberTimestamp(client.uid)

	rpc := GetGroupStorageRPCClient(group_id)

	last_id := group_sync_key.Sync_key
	if last_id == 0 {
		last_id = GetGroupSyncKey(client.appid, client.uid, group_id)
	}

	s := &SyncGroupHistory{
		AppID:     client.appid,
		Uid:       client.uid,
		DeviceID:  client.device_ID,
		GroupID:   group_sync_key.Group_id,
		LastMsgID: last_id,
		Timestamp: int32(ts),
	}

	log.Info("sync group message...", group_sync_key.Sync_key, last_id)
	resp, err := rpc.Call("SyncGroupMessage", s)
	if err != nil {
		log.Warning("sync message err:", err)
		return
	}

	gh := resp.(*GroupHistoryMessage)
	messages := gh.Messages

	sk := &CommonModel.GroupSyncKey{Sync_key: last_id, Group_id: group_id}
	client.EnqueueMessage(&CommonModel.Message{Cmd: CommonModel.MSG_SYNC_GROUP_BEGIN,
		Body: sk})
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		log.Info("message:", msg.MsgID, CommonModel.Command(msg.Cmd))
		m := &CommonModel.Message{Cmd: int(msg.Cmd),
			Version: CommonModel.DEFAULT_VERSION}
		m.FromData(msg.Raw)
		sk.Sync_key = msg.MsgID

		if config.Config.Sync_self {
			//连接成功后的首次同步，自己发送的消息也下发给客户端
			//过滤掉所有自己在当前设备发出的消息
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
		client.EnqueueMessage(m)
	}

	if gh.LastMsgID < last_id && gh.LastMsgID > 0 {
		sk.Sync_key = gh.LastMsgID
		log.Warningf("group:%d client last id:%d server last id:%d", group_id, last_id, gh.LastMsgID)
	}
	client.EnqueueMessage(&CommonModel.Message{Cmd: CommonModel.MSG_SYNC_GROUP_END,
		Body: sk})
}

func (client *GroupClient) HandleGroupSyncKey(group_sync_key *CommonModel.GroupSyncKey) {
	if client.uid == 0 {
		return
	}

	group_id := group_sync_key.Group_id
	last_id := group_sync_key.Sync_key

	log.Info("group sync key:", group_sync_key.Sync_key, last_id)
	if last_id > 0 {
		s := &SyncGroupHistory{
			AppID:     client.appid,
			Uid:       client.uid,
			GroupID:   group_id,
			LastMsgID: last_id,
		}
		Group_sync_c <- s
	}
}

func (client *GroupClient) HandleMessage(msg *CommonModel.Message) {
	switch msg.Cmd {
	case CommonModel.MSG_GROUP_IM:
		client.HandleGroupIMMessage(msg)
	case CommonModel.MSG_SYNC_GROUP:
		client.HandleGroupSync(msg.Body.(*CommonModel.GroupSyncKey))
	case CommonModel.MSG_GROUP_SYNC_KEY:
		client.HandleGroupSyncKey(msg.Body.(*CommonModel.GroupSyncKey))
	}
}
