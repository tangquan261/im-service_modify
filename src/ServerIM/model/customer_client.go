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
	"time"

	"ServerIM/config"

	log "github.com/golang/glog"
)

type CustomerClient struct {
	*Connection
}

// func NewCustomerClient(conn *Connection) *CustomerClient {
// 	c := &CustomerClient{Connection: conn}
// 	return c
// }

func (client *CustomerClient) HandleMessage(msg *CommonModel.Message) {
	switch msg.Cmd {
	case CommonModel.MSG_CUSTOMER:
		client.HandleCustomerMessage(msg)
	case CommonModel.MSG_CUSTOMER_SUPPORT:
		client.HandleCustomerSupportMessage(msg)
	}
}

//客服->顾客
func (client *CustomerClient) HandleCustomerSupportMessage(msg *CommonModel.Message) {
	cm := msg.Body.(*CommonModel.CustomerMessage)
	if client.appid != config.Config.Kefu_appid {
		log.Warningf("client appid:%d kefu appid:%d",
			client.appid, config.Config.Kefu_appid)
		return
	}
	if client.uid != cm.Seller_id {
		log.Warningf("uid:%d seller id:%d", client.uid, cm.Seller_id)
		return
	}

	cm.Timestamp = int32(time.Now().Unix())

	if (msg.Flag & CommonModel.MESSAGE_FLAG_UNPERSISTENT) > 0 {
		log.Info("customer support message unpersistent")
		SendAppMessage(cm.Customer_appid, cm.Customer_id, msg)
		ack := &CommonModel.Message{Cmd: CommonModel.MSG_ACK,
			Body: &CommonModel.MessageACK{int32(msg.Seq)}}
		client.EnqueueMessage(ack)
		return
	}

	msgid, err := SaveMessage(cm.Customer_appid, cm.Customer_id, client.device_ID, msg)
	if err != nil {
		log.Warning("save customer support message err:", err)
		return
	}

	msgid2, err := SaveMessage(client.appid, cm.Seller_id, client.device_ID, msg)
	if err != nil {
		log.Warning("save customer support message err:", err)
		return
	}

	PushMessage(cm.Customer_appid, cm.Customer_id, msg)

	//发送同步的通知消息
	notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{msgid}}
	SendAppMessage(cm.Customer_appid, cm.Customer_id, notify)

	//发送给自己的其它登录点
	notify = &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{msgid2}}
	client.SendMessage(client.uid, notify)

	ack := &CommonModel.Message{Cmd: CommonModel.MSG_ACK,
		Body: &CommonModel.MessageACK{int32(msg.Seq)}}
	client.EnqueueMessage(ack)
}

//顾客->客服
func (client *CustomerClient) HandleCustomerMessage(msg *CommonModel.Message) {
	cm := msg.Body.(*CommonModel.CustomerMessage)
	cm.Timestamp = int32(time.Now().Unix())

	log.Infof("customer message customer appid:%d customer id:%d store id:%d seller id:%d",
		cm.Customer_appid, cm.Customer_id, cm.Store_id, cm.Seller_id)
	if cm.Customer_appid != client.appid {
		log.Warningf("message appid:%d client appid:%d",
			cm.Customer_appid, client.appid)
		return
	}
	if cm.Customer_id != client.uid {
		log.Warningf("message customer id:%d client uid:%d",
			cm.Customer_id, client.uid)
		return
	}

	if cm.Seller_id == 0 {
		log.Warningf("message seller id:0")
		return
	}

	if (msg.Flag & CommonModel.MESSAGE_FLAG_UNPERSISTENT) > 0 {
		log.Info("customer message unpersistent")
		SendAppMessage(config.Config.Kefu_appid, cm.Seller_id, msg)
		ack := &CommonModel.Message{Cmd: CommonModel.MSG_ACK,
			Body: &CommonModel.MessageACK{int32(msg.Seq)}}
		client.EnqueueMessage(ack)
		return
	}

	msgid, err := SaveMessage(config.Config.Kefu_appid, cm.Seller_id, client.device_ID, msg)
	if err != nil {
		log.Warning("save customer message err:", err)
		return
	}

	msgid2, err := SaveMessage(cm.Customer_appid, cm.Customer_id, client.device_ID, msg)
	if err != nil {
		log.Warning("save customer message err:", err)
		return
	}

	PushMessage(config.Config.Kefu_appid, cm.Seller_id, msg)

	//发送同步的通知消息
	notify := &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{msgid}}
	SendAppMessage(config.Config.Kefu_appid, cm.Seller_id, notify)

	//发送给自己的其它登录点
	notify = &CommonModel.Message{Cmd: CommonModel.MSG_SYNC_NOTIFY,
		Body: &CommonModel.SyncKey{msgid2}}
	client.SendMessage(client.uid, notify)

	ack := &CommonModel.Message{Cmd: CommonModel.MSG_ACK,
		Body: &CommonModel.MessageACK{int32(msg.Seq)}}
	client.EnqueueMessage(ack)
}
