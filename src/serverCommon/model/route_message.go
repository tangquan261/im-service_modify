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

import "bytes"
import "encoding/binary"

//路由服务器消息
const MSG_SUBSCRIBE = 130   //在线请求
const MSG_UNSUBSCRIBE = 131 //离线请求
const MSG_PUBLISH = 132     //私聊消息

const MSG_PUBLISH_GROUP = 135 //广播给其他所有群聊服务器

const MSG_SUBSCRIBE_ROOM = 136   //加入房间
const MSG_UNSUBSCRIBE_ROOM = 137 //离开房间
const MSG_PUBLISH_ROOM = 138     //广播房间消息

func init() {
	Message_creators[MSG_SUBSCRIBE] = func() IMessage { return new(SubscribeMessage) }
	Message_creators[MSG_UNSUBSCRIBE] = func() IMessage { return new(AppUserID) }
	Message_creators[MSG_PUBLISH] = func() IMessage { return new(AppMessage) }

	Message_creators[MSG_PUBLISH_GROUP] = func() IMessage { return new(AppMessage) }

	Message_creators[MSG_SUBSCRIBE_ROOM] = func() IMessage { return new(AppRoomID) }
	Message_creators[MSG_UNSUBSCRIBE_ROOM] = func() IMessage { return new(AppRoomID) }
	Message_creators[MSG_PUBLISH_ROOM] = func() IMessage { return new(AppMessage) }

	Message_descriptions[MSG_SUBSCRIBE] = "MSG_SUBSCRIBE"
	Message_descriptions[MSG_UNSUBSCRIBE] = "MSG_UNSUBSCRIBE"
	Message_descriptions[MSG_PUBLISH] = "MSG_PUBLISH"

	Message_descriptions[MSG_PUBLISH_GROUP] = "MSG_PUBLISH_GROUP"

	Message_descriptions[MSG_SUBSCRIBE_ROOM] = "MSG_SUBSCRIBE_ROOM"
	Message_descriptions[MSG_UNSUBSCRIBE_ROOM] = "MSG_UNSUBSCRIBE_ROOM"
	Message_descriptions[MSG_PUBLISH_ROOM] = "MSG_PUBLISH_ROOM"
}

type AppMessage struct {
	Appid     int64
	Receiver  int64
	Msgid     int64
	Device_id int64
	Timestamp int64 //纳秒,测试消息从im->imr->im的时间
	Msg       *Message
}

func (amsg *AppMessage) ToData() []byte {
	if amsg.Msg == nil {
		return nil
	}

	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, amsg.Appid)
	binary.Write(buffer, binary.BigEndian, amsg.Receiver)
	binary.Write(buffer, binary.BigEndian, amsg.Msgid)
	binary.Write(buffer, binary.BigEndian, amsg.Device_id)
	binary.Write(buffer, binary.BigEndian, amsg.Timestamp)
	mbuffer := new(bytes.Buffer)
	WriteMessage(mbuffer, amsg.Msg)
	msg_buf := mbuffer.Bytes()
	var l int16 = int16(len(msg_buf))
	binary.Write(buffer, binary.BigEndian, l)
	buffer.Write(msg_buf)

	buf := buffer.Bytes()
	return buf
}

func (amsg *AppMessage) FromData(buff []byte) bool {
	if len(buff) < 42 {
		return false
	}

	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &amsg.Appid)
	binary.Read(buffer, binary.BigEndian, &amsg.Receiver)
	binary.Read(buffer, binary.BigEndian, &amsg.Msgid)
	binary.Read(buffer, binary.BigEndian, &amsg.Device_id)
	binary.Read(buffer, binary.BigEndian, &amsg.Timestamp)

	var l int16
	binary.Read(buffer, binary.BigEndian, &l)
	if int(l) > buffer.Len() || l < 0 {
		return false
	}

	msg_buf := make([]byte, l)
	buffer.Read(msg_buf)

	mbuffer := bytes.NewBuffer(msg_buf)
	//recusive
	msg := ReceiveMessage(mbuffer)
	if msg == nil {
		return false
	}
	amsg.Msg = msg

	return true
}

type SubscribeMessage struct {
	Appid  int64
	Uid    int64
	Online int8 //1 or 0
}

func (sub *SubscribeMessage) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, sub.Appid)
	binary.Write(buffer, binary.BigEndian, sub.Uid)
	binary.Write(buffer, binary.BigEndian, sub.Online)
	buf := buffer.Bytes()
	return buf
}

func (sub *SubscribeMessage) FromData(buff []byte) bool {
	if len(buff) < 17 {
		return false
	}

	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &sub.Appid)
	binary.Read(buffer, binary.BigEndian, &sub.Uid)
	binary.Read(buffer, binary.BigEndian, &sub.Online)

	return true
}
