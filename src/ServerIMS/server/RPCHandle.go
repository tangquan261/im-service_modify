package server

import (
	Commonconfig "ServerCommon/config"
	CommonModel "ServerCommon/model"
	"ServerIMS/config"
	"ServerIMS/model"
	"sync/atomic"
)

func SyncMessage(addr string, sync_key *model.SyncHistory) *model.PeerHistoryMessage {
	atomic.AddInt64(&model.Server_summary.Nrequests, 1)
	messages, last_msgid, hasMore := model.StorageObj.LoadHistoryMessagesV3(sync_key.AppID,
		sync_key.Uid, sync_key.LastMsgID, config.Config.Limit, config.Config.Hard_limit)

	historyMessages := make([]*model.HistoryMessage, 0, 10)
	for _, emsg := range messages {
		hm := &model.HistoryMessage{}
		hm.MsgID = emsg.Msgid
		hm.DeviceID = emsg.Device_id
		hm.Cmd = int32(emsg.Msg.Cmd)

		emsg.Msg.Version = model.DEFAULT_VERSION
		hm.Raw = emsg.Msg.ToData()
		historyMessages = append(historyMessages, hm)
	}

	return &model.PeerHistoryMessage{historyMessages, last_msgid, hasMore}
}

func SyncGroupMessage(addr string, sync_key *model.SyncGroupHistory) *model.GroupHistoryMessage {
	atomic.AddInt64(&model.Server_summary.Nrequests, 1)
	messages, last_msgid := model.StorageObj.LoadGroupHistoryMessages(sync_key.AppID,
		sync_key.Uid, sync_key.GroupID, sync_key.LastMsgID,
		sync_key.Timestamp, Commonconfig.GROUP_OFFLINE_LIMIT)

	historyMessages := make([]*model.HistoryMessage, 0, 10)
	for _, emsg := range messages {
		hm := &model.HistoryMessage{}
		hm.MsgID = emsg.Msgid
		hm.DeviceID = emsg.Device_id
		hm.Cmd = int32(emsg.Msg.Cmd)

		emsg.Msg.Version = model.DEFAULT_VERSION
		hm.Raw = emsg.Msg.ToData()
		historyMessages = append(historyMessages, hm)
	}

	return &model.GroupHistoryMessage{historyMessages, last_msgid, false}
}

func SavePeerMessage(addr string, m *model.PeerMessage) (int64, error) {
	//保存消息，请求总数数量++，私聊消息总数++
	atomic.AddInt64(&model.Server_summary.Nrequests, 1)
	atomic.AddInt64(&model.Server_summary.Peer_message_count, 1)
	msg := &CommonModel.Message{Cmd: int(m.Cmd), Version: model.DEFAULT_VERSION}
	msg.FromData(m.Raw)
	msgid := model.StorageObj.SavePeerMessage(m.AppID, m.Uid, m.DeviceID, msg)
	return msgid, nil
}

func SaveGroupMessage(addr string, m *model.GroupMessage) (int64, error) {
	atomic.AddInt64(&model.Server_summary.Nrequests, 1)
	atomic.AddInt64(&model.Server_summary.Group_message_count, 1)
	msg := &CommonModel.Message{Cmd: int(m.Cmd), Version: model.DEFAULT_VERSION}
	msg.FromData(m.Raw)
	msgid := model.StorageObj.SaveGroupMessage(m.AppID, m.GroupID, m.DeviceID, msg)
	return msgid, nil
}

func GetNewCount(addr string, sync_key *model.SyncHistory) (int64, error) {
	atomic.AddInt64(&model.Server_summary.Nrequests, 1)
	count := model.StorageObj.GetNewCount(sync_key.AppID, sync_key.Uid, sync_key.LastMsgID)
	return int64(count), nil
}

func GetLatestMessage(addr string, r *model.HistoryRequest) []*model.HistoryMessage {
	atomic.AddInt64(&model.Server_summary.Nrequests, 1)
	messages := model.StorageObj.LoadLatestMessages(r.AppID, r.Uid, int(r.Limit))

	historyMessages := make([]*model.HistoryMessage, 0, 10)
	for _, emsg := range messages {
		hm := &model.HistoryMessage{}
		hm.MsgID = emsg.Msgid
		hm.DeviceID = emsg.Device_id
		hm.Cmd = int32(emsg.Msg.Cmd)

		emsg.Msg.Version = model.DEFAULT_VERSION
		hm.Raw = emsg.Msg.ToData()
		historyMessages = append(historyMessages, hm)
	}

	return historyMessages
}
