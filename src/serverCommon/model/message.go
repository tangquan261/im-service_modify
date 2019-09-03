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
import "fmt"

const MSG_AUTH_STATUS = 3

//persistent
const MSG_IM = 4

const MSG_ACK = 5

//deprecated
const MSG_RST = 6

//persistent
const MSG_GROUP_NOTIFICATION = 7 //群聊通知
const MSG_GROUP_IM = 8           //群聊IM

const MSG_PING = 13
const MSG_PONG = 14
const MSG_AUTH_TOKEN = 15

const MSG_RT = 17
const MSG_ENTER_ROOM = 18
const MSG_LEAVE_ROOM = 19
const MSG_ROOM_IM = 20

//persistent
const MSG_SYSTEM = 21

const MSG_UNREAD_COUNT = 22

//persistent, deprecated
const MSG_CUSTOMER_SERVICE_ = 23

//persistent
const MSG_CUSTOMER = 24         //顾客->客服
const MSG_CUSTOMER_SUPPORT = 25 //客服->顾客

//客户端->服务端
const MSG_SYNC = 26 //同步消息
//服务端->客服端
const MSG_SYNC_BEGIN = 27
const MSG_SYNC_END = 28

//通知客户端有新消息
const MSG_SYNC_NOTIFY = 29

//客户端->服务端
const MSG_SYNC_GROUP = 30 //同步超级群消息
//服务端->客服端
const MSG_SYNC_GROUP_BEGIN = 31
const MSG_SYNC_GROUP_END = 32

//通知客户端有新消息
const MSG_SYNC_GROUP_NOTIFY = 33

//客服端->服务端,更新服务器的synckey
const MSG_SYNC_KEY = 34
const MSG_GROUP_SYNC_KEY = 35

//系统通知消息, unpersistent
const MSG_NOTIFICATION = 36

const MSG_VOIP_CONTROL = 64

//消息标志
//文本消息
const MESSAGE_FLAG_TEXT = 0x01

//消息不持久化
const MESSAGE_FLAG_UNPERSISTENT = 0x02

//群组离线消息 MSG_OFFLINE使用
const MESSAGE_FLAG_GROUP = 0x04

//离线消息由当前登录的用户在当前设备发出
const MESSAGE_FLAG_SELF = 0x08

func init() {
	Message_creators[MSG_ACK] = func() IMessage { return new(MessageACK) }
	Message_creators[MSG_GROUP_NOTIFICATION] = func() IMessage { return new(GroupNotification) }

	Message_creators[MSG_AUTH_TOKEN] = func() IMessage { return new(AuthenticationToken) }

	Message_creators[MSG_RT] = func() IMessage { return new(RTMessage) }
	Message_creators[MSG_ENTER_ROOM] = func() IMessage { return new(Room) }
	Message_creators[MSG_LEAVE_ROOM] = func() IMessage { return new(Room) }
	Message_creators[MSG_ROOM_IM] = func() IMessage { return &RoomMessage{new(RTMessage)} }
	Message_creators[MSG_SYSTEM] = func() IMessage { return new(SystemMessage) }
	Message_creators[MSG_UNREAD_COUNT] = func() IMessage { return new(MessageUnreadCount) }
	Message_creators[MSG_CUSTOMER_SERVICE_] = func() IMessage { return new(IgnoreMessage) }

	Message_creators[MSG_CUSTOMER] = func() IMessage { return new(CustomerMessage) }
	Message_creators[MSG_CUSTOMER_SUPPORT] = func() IMessage { return new(CustomerMessage) }

	Message_creators[MSG_SYNC] = func() IMessage { return new(SyncKey) }
	Message_creators[MSG_SYNC_BEGIN] = func() IMessage { return new(SyncKey) }
	Message_creators[MSG_SYNC_END] = func() IMessage { return new(SyncKey) }
	Message_creators[MSG_SYNC_NOTIFY] = func() IMessage { return new(SyncKey) }
	Message_creators[MSG_SYNC_KEY] = func() IMessage { return new(SyncKey) }

	Message_creators[MSG_SYNC_GROUP] = func() IMessage { return new(GroupSyncKey) }
	Message_creators[MSG_SYNC_GROUP_BEGIN] = func() IMessage { return new(GroupSyncKey) }
	Message_creators[MSG_SYNC_GROUP_END] = func() IMessage { return new(GroupSyncKey) }
	Message_creators[MSG_SYNC_GROUP_NOTIFY] = func() IMessage { return new(GroupSyncKey) }
	Message_creators[MSG_GROUP_SYNC_KEY] = func() IMessage { return new(GroupSyncKey) }

	Message_creators[MSG_NOTIFICATION] = func() IMessage { return new(SystemMessage) }

	Message_creators[MSG_VOIP_CONTROL] = func() IMessage { return new(VOIPControl) }

	vmessage_creators[MSG_GROUP_IM] = func() IVersionMessage { return new(IMMessage) }
	vmessage_creators[MSG_IM] = func() IVersionMessage { return new(IMMessage) }

	Message_creators[MSG_AUTH_STATUS] = func() IMessage { return new(AuthenticationStatus) }

	Message_descriptions[MSG_AUTH_STATUS] = "MSG_AUTH_STATUS"
	Message_descriptions[MSG_IM] = "MSG_IM"
	Message_descriptions[MSG_ACK] = "MSG_ACK"
	Message_descriptions[MSG_GROUP_NOTIFICATION] = "MSG_GROUP_NOTIFICATION"
	Message_descriptions[MSG_GROUP_IM] = "MSG_GROUP_IM"
	Message_descriptions[MSG_PING] = "MSG_PING"
	Message_descriptions[MSG_PONG] = "MSG_PONG"
	Message_descriptions[MSG_AUTH_TOKEN] = "MSG_AUTH_TOKEN"
	Message_descriptions[MSG_RT] = "MSG_RT"
	Message_descriptions[MSG_ENTER_ROOM] = "MSG_ENTER_ROOM"
	Message_descriptions[MSG_LEAVE_ROOM] = "MSG_LEAVE_ROOM"
	Message_descriptions[MSG_ROOM_IM] = "MSG_ROOM_IM"
	Message_descriptions[MSG_SYSTEM] = "MSG_SYSTEM"
	Message_descriptions[MSG_UNREAD_COUNT] = "MSG_UNREAD_COUNT"
	Message_descriptions[MSG_CUSTOMER_SERVICE_] = "MSG_CUSTOMER_SERVICE"
	Message_descriptions[MSG_CUSTOMER] = "MSG_CUSTOMER"
	Message_descriptions[MSG_CUSTOMER_SUPPORT] = "MSG_CUSTOMER_SUPPORT"

	Message_descriptions[MSG_SYNC] = "MSG_SYNC"
	Message_descriptions[MSG_SYNC_BEGIN] = "MSG_SYNC_BEGIN"
	Message_descriptions[MSG_SYNC_END] = "MSG_SYNC_END"
	Message_descriptions[MSG_SYNC_NOTIFY] = "MSG_SYNC_NOTIFY"

	Message_descriptions[MSG_SYNC_GROUP] = "MSG_SYNC_GROUP"
	Message_descriptions[MSG_SYNC_GROUP_BEGIN] = "MSG_SYNC_GROUP_BEGIN"
	Message_descriptions[MSG_SYNC_GROUP_END] = "MSG_SYNC_GROUP_END"
	Message_descriptions[MSG_SYNC_GROUP_NOTIFY] = "MSG_SYNC_GROUP_NOTIFY"

	Message_descriptions[MSG_NOTIFICATION] = "MSG_NOTIFICATION"
	Message_descriptions[MSG_VOIP_CONTROL] = "MSG_VOIP_CONTROL"

	external_messages[MSG_AUTH_TOKEN] = true
	external_messages[MSG_IM] = true
	external_messages[MSG_ACK] = true
	external_messages[MSG_GROUP_IM] = true
	external_messages[MSG_PING] = true
	external_messages[MSG_PONG] = true
	external_messages[MSG_RT] = true
	external_messages[MSG_ENTER_ROOM] = true
	external_messages[MSG_LEAVE_ROOM] = true
	external_messages[MSG_ROOM_IM] = true
	external_messages[MSG_UNREAD_COUNT] = true
	external_messages[MSG_CUSTOMER] = true
	external_messages[MSG_CUSTOMER_SUPPORT] = true
	external_messages[MSG_SYNC] = true
	external_messages[MSG_SYNC_GROUP] = true
	external_messages[MSG_SYNC_KEY] = true
	external_messages[MSG_GROUP_SYNC_KEY] = true
}

type Command int

func (cmd Command) String() string {
	c := int(cmd)
	if desc, ok := Message_descriptions[c]; ok {
		return desc
	} else {
		return fmt.Sprintf("%d", c)
	}
}

type IMessage interface {
	ToData() []byte
	FromData(buff []byte) bool
}

type IVersionMessage interface {
	ToData(version int) []byte
	FromData(version int, buff []byte) bool
}

type Message struct {
	Cmd     int
	Seq     int
	Version int
	Flag    int

	Body interface{}
}

func (message *Message) ToData() []byte {
	if message.Body != nil {
		if m, ok := message.Body.(IMessage); ok {
			return m.ToData()
		}
		if m, ok := message.Body.(IVersionMessage); ok {
			return m.ToData(message.Version)
		}
		return nil
	} else {
		return nil
	}
}

func (message *Message) FromData(buff []byte) bool {
	cmd := message.Cmd
	if creator, ok := Message_creators[cmd]; ok {
		c := creator()
		r := c.FromData(buff)
		message.Body = c
		return r
	}
	if creator, ok := vmessage_creators[cmd]; ok {
		c := creator()
		r := c.FromData(message.Version, buff)
		message.Body = c
		return r
	}

	return len(buff) == 0
}

//保存在磁盘中但不再需要处理的消息
type IgnoreMessage struct {
}

func (ignore *IgnoreMessage) ToData() []byte {
	return nil
}

func (ignore *IgnoreMessage) FromData(buff []byte) bool {
	return true
}

type AuthenticationToken struct {
	Token       string
	Platform_id int8
	Device_id   string
}

func (auth *AuthenticationToken) ToData() []byte {
	var l int8
	var tokenlen int16
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, auth.Platform_id)

	tokenlen = int16(len(auth.Token))
	binary.Write(buffer, binary.BigEndian, tokenlen)
	buffer.Write([]byte(auth.Token))

	l = int8(len(auth.Device_id))
	binary.Write(buffer, binary.BigEndian, l)
	buffer.Write([]byte(auth.Device_id))

	buf := buffer.Bytes()
	return buf
}

func (auth *AuthenticationToken) FromData(buff []byte) bool {
	var ltoken int16
	if len(buff) <= 3 {
		return false
	}
	auth.Platform_id = int8(buff[0])
	fmt.Println(auth.Platform_id)
	buffer := bytes.NewBuffer(buff[1:])

	binary.Read(buffer, binary.BigEndian, &ltoken)

	if int(ltoken) > buffer.Len() || int(ltoken) < 0 {
		return false
	}

	token := make([]byte, ltoken)
	buffer.Read(token)

	var device_idl int8
	binary.Read(buffer, binary.BigEndian, &device_idl)
	if int(device_idl) > buffer.Len() || int(device_idl) < 0 {
		return false
	}
	device_id := make([]byte, device_idl)
	buffer.Read(device_id)

	auth.Token = string(token)
	auth.Device_id = string(device_id)
	return true
}

type AuthenticationStatus struct {
	Status int32
}

func (auth *AuthenticationStatus) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, auth.Status)
	buf := buffer.Bytes()
	return buf
}

func (auth *AuthenticationStatus) FromData(buff []byte) bool {
	if len(buff) < 4 {
		return false
	}
	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &auth.Status)
	return true
}

type RTMessage struct {
	Sender   int64
	Receiver int64
	Content  string
}

func (message *RTMessage) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, message.Sender)
	binary.Write(buffer, binary.BigEndian, message.Receiver)
	buffer.Write([]byte(message.Content))
	buf := buffer.Bytes()
	return buf
}

func (rt *RTMessage) FromData(buff []byte) bool {
	if len(buff) < 16 {
		return false
	}
	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &rt.Sender)
	binary.Read(buffer, binary.BigEndian, &rt.Receiver)
	rt.Content = string(buff[16:])
	return true
}

type IMMessage struct {
	Sender    int64
	Receiver  int64
	Timestamp int32
	Msgid     int32
	Content   string
}

func (message *IMMessage) ToDataV0() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, message.Sender)
	binary.Write(buffer, binary.BigEndian, message.Receiver)
	binary.Write(buffer, binary.BigEndian, message.Msgid)
	buffer.Write([]byte(message.Content))
	buf := buffer.Bytes()
	return buf
}

func (im *IMMessage) FromDataV0(buff []byte) bool {
	if len(buff) < 20 {
		return false
	}
	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &im.Sender)
	binary.Read(buffer, binary.BigEndian, &im.Receiver)
	binary.Read(buffer, binary.BigEndian, &im.Msgid)
	im.Content = string(buff[20:])
	return true
}

func (message *IMMessage) ToDataV1() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, message.Sender)
	binary.Write(buffer, binary.BigEndian, message.Receiver)
	binary.Write(buffer, binary.BigEndian, message.Timestamp)
	binary.Write(buffer, binary.BigEndian, message.Msgid)
	buffer.Write([]byte(message.Content))
	buf := buffer.Bytes()
	return buf
}

func (im *IMMessage) FromDataV1(buff []byte) bool {
	if len(buff) < 24 {
		return false
	}
	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &im.Sender)
	binary.Read(buffer, binary.BigEndian, &im.Receiver)
	binary.Read(buffer, binary.BigEndian, &im.Timestamp)
	binary.Read(buffer, binary.BigEndian, &im.Msgid)
	im.Content = string(buff[24:])
	return true
}

func (im *IMMessage) ToData(version int) []byte {
	if version == 0 {
		return im.ToDataV0()
	} else {
		return im.ToDataV1()
	}
}

func (im *IMMessage) FromData(version int, buff []byte) bool {
	if version == 0 {
		return im.FromDataV0(buff)
	} else {
		return im.FromDataV1(buff)
	}
}

type MessageACK struct {
	Seq int32
}

func (ack *MessageACK) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, ack.Seq)
	buf := buffer.Bytes()
	return buf
}

func (ack *MessageACK) FromData(buff []byte) bool {
	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &ack.Seq)
	return true
}

type MessageUnreadCount struct {
	Count int32
}

func (u *MessageUnreadCount) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, u.Count)
	buf := buffer.Bytes()
	return buf
}

func (u *MessageUnreadCount) FromData(buff []byte) bool {
	if len(buff) < 4 {
		return false
	}
	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &u.Count)
	return true
}

type SystemMessage struct {
	Notification string
}

func (sys *SystemMessage) ToData() []byte {
	return []byte(sys.Notification)
}

func (sys *SystemMessage) FromData(buff []byte) bool {
	sys.Notification = string(buff)
	return true
}

type CustomerMessage struct {
	Customer_appid int64 //顾客id所在appid
	Customer_id    int64 //顾客id
	Store_id       int64
	Seller_id      int64
	Timestamp      int32
	Content        string
}

func (cs *CustomerMessage) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, cs.Customer_appid)
	binary.Write(buffer, binary.BigEndian, cs.Customer_id)
	binary.Write(buffer, binary.BigEndian, cs.Store_id)
	binary.Write(buffer, binary.BigEndian, cs.Seller_id)
	binary.Write(buffer, binary.BigEndian, cs.Timestamp)
	buffer.Write([]byte(cs.Content))
	buf := buffer.Bytes()
	return buf
}

func (cs *CustomerMessage) FromData(buff []byte) bool {
	if len(buff) < 36 {
		return false
	}
	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &cs.Customer_appid)
	binary.Read(buffer, binary.BigEndian, &cs.Customer_id)
	binary.Read(buffer, binary.BigEndian, &cs.Store_id)
	binary.Read(buffer, binary.BigEndian, &cs.Seller_id)
	binary.Read(buffer, binary.BigEndian, &cs.Timestamp)

	cs.Content = string(buff[36:])

	return true
}

type GroupNotification struct {
	Notification string
}

func (notification *GroupNotification) ToData() []byte {
	return []byte(notification.Notification)
}

func (notification *GroupNotification) FromData(buff []byte) bool {
	notification.Notification = string(buff)
	return true
}

type Room int64

func (room *Room) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, int64(*room))
	buf := buffer.Bytes()
	return buf
}

func (room *Room) FromData(buff []byte) bool {
	if len(buff) < 8 {
		return false
	}
	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, (*int64)(room))
	return true
}

func (room *Room) RoomID() int64 {
	return int64(*room)
}

type RoomMessage struct {
	*RTMessage
}

type VOIPControl struct {
	sender   int64
	receiver int64
	content  []byte
}

func (ctl *VOIPControl) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, ctl.sender)
	binary.Write(buffer, binary.BigEndian, ctl.receiver)
	buffer.Write([]byte(ctl.content))
	buf := buffer.Bytes()
	return buf
}

func (ctl *VOIPControl) FromData(buff []byte) bool {
	if len(buff) <= 16 {
		return false
	}

	buffer := bytes.NewBuffer(buff[:16])
	binary.Read(buffer, binary.BigEndian, &ctl.sender)
	binary.Read(buffer, binary.BigEndian, &ctl.receiver)
	ctl.content = buff[16:]
	return true
}

type AppUserID struct {
	Appid int64
	Uid   int64
}

func (id *AppUserID) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, id.Appid)
	binary.Write(buffer, binary.BigEndian, id.Uid)
	buf := buffer.Bytes()
	return buf
}

func (id *AppUserID) FromData(buff []byte) bool {
	if len(buff) < 16 {
		return false
	}

	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &id.Appid)
	binary.Read(buffer, binary.BigEndian, &id.Uid)

	return true
}

type AppRoomID struct {
	Appid   int64
	Room_id int64
}

func (id *AppRoomID) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, id.Appid)
	binary.Write(buffer, binary.BigEndian, id.Room_id)
	buf := buffer.Bytes()
	return buf
}

func (id *AppRoomID) FromData(buff []byte) bool {
	if len(buff) < 16 {
		return false
	}

	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &id.Appid)
	binary.Read(buffer, binary.BigEndian, &id.Room_id)

	return true
}

type AppGroupMemberID struct {
	Appid int64
	Gid   int64
	Uid   int64
}

func (id *AppGroupMemberID) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, id.Appid)
	binary.Write(buffer, binary.BigEndian, id.Gid)
	binary.Write(buffer, binary.BigEndian, id.Uid)
	buf := buffer.Bytes()
	return buf
}

func (id *AppGroupMemberID) FromData(buff []byte) bool {
	if len(buff) < 24 {
		return false
	}

	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &id.Appid)
	binary.Read(buffer, binary.BigEndian, &id.Gid)
	binary.Read(buffer, binary.BigEndian, &id.Uid)

	return true
}

type SyncKey struct {
	Sync_key int64
}

func (id *SyncKey) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, id.Sync_key)
	buf := buffer.Bytes()
	return buf
}

func (id *SyncKey) FromData(buff []byte) bool {
	if len(buff) < 8 {
		return false
	}

	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &id.Sync_key)
	return true
}

type GroupSyncKey struct {
	Group_id int64
	Sync_key int64
}

func (id *GroupSyncKey) ToData() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, id.Group_id)
	binary.Write(buffer, binary.BigEndian, id.Sync_key)
	buf := buffer.Bytes()
	return buf
}

func (id *GroupSyncKey) FromData(buff []byte) bool {
	if len(buff) < 16 {
		return false
	}

	buffer := bytes.NewBuffer(buff)
	binary.Read(buffer, binary.BigEndian, &id.Group_id)
	binary.Read(buffer, binary.BigEndian, &id.Sync_key)
	return true
}
