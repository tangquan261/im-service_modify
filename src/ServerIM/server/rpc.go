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

package server

//import "net/http"
//import "encoding/json"
//import "time"
//import "net/url"
//import "strconv"
import "sync/atomic"
import log "github.com/golang/glog"

//import "io/ioutil"

import (
	CommonModel "ServerCommon/model"
	"ServerIM/model"
)

func SendGroupIMMessage(im *CommonModel.IMMessage, appid int64) {
	m := &CommonModel.Message{Cmd: CommonModel.MSG_GROUP_IM,
		Version: CommonModel.DEFAULT_VERSION, Body: im}
	group := CommonModel.Group_manager.FindGroup(im.Receiver)
	if group == nil {
		log.Warning("can't find group:", im.Receiver)
		return
	}
	if group.Super {
		msgid, err := model.SaveGroupMessage(appid, im.Receiver, 0, m)
		if err != nil {
			return
		}

		//推送外部通知
		model.PushGroupMessage(appid, im.Receiver, m)

		//发送同步的通知消息
		notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_GROUP_NOTIFY,
			Body: &CommonModel.GroupSyncKey{Group_id: im.Receiver, Sync_key: msgid}}
		model.SendAppGroupMessage(appid, im.Receiver, notify)

	} else {
		members := group.Members()
		for member := range members {
			msgid, err := model.SaveMessage(appid, member, 0, m)
			if err != nil {
				continue
			}

			//推送外部通知
			model.PushMessage(appid, member, m)

			//发送同步的通知消息
			notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
				Body: &CommonModel.SyncKey{Sync_key: msgid}}
			model.SendAppMessage(appid, member, notify)
		}
	}
	atomic.AddInt64(&model.Server_summary.In_message_count, 1)
}

func SendIMMessage(im *CommonModel.IMMessage, appid int64) {
	m := &CommonModel.Message{Cmd: CommonModel.MSG_IM,
		Version: CommonModel.DEFAULT_VERSION, Body: im}
	msgid, err := model.SaveMessage(appid, im.Receiver, 0, m)
	if err != nil {
		return
	}

	//保存到发送者自己的消息队列
	msgid2, err := model.SaveMessage(appid, im.Sender, 0, m)
	if err != nil {
		return
	}

	//推送外部通知
	model.PushMessage(appid, im.Receiver, m)

	//发送同步的通知消息
	notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{Sync_key: msgid}}
	model.SendAppMessage(appid, im.Receiver, notify)

	//发送同步的通知消息
	notify = &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{Sync_key: msgid2}}
	model.SendAppMessage(appid, im.Sender, notify)

	atomic.AddInt64(&model.Server_summary.In_message_count, 1)
}

func SyncMessageInterface(addr string, sync_key *model.SyncHistory) *model.PeerHistoryMessage {
	return nil
}

func SyncGroupMessageInterface(addr string, sync_key *model.SyncGroupHistory) *model.GroupHistoryMessage {
	return nil
}

func SavePeerMessageInterface(addr string, m *model.PeerMessage) (int64, error) {
	return 0, nil
}

func SaveGroupMessageInterface(addr string, m *model.GroupMessage) (int64, error) {
	return 0, nil
}

//获取是否接收到新消息,只会返回0/1
func GetNewCountInterface(addr string, s *model.SyncHistory) (int64, error) {
	return 0, nil
}

func GetLatestMessageInterface(addr string, r *model.HistoryRequest) []*model.HistoryMessage {
	return nil
}
