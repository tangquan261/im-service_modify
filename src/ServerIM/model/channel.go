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

import "net"
import "time"
import "sync"
import log "github.com/golang/glog"
import "ServerCommon/model"

type Subscriber struct {
	uids     map[int64]int
	room_ids map[int64]int
}

func NewSubscriber() *Subscriber {
	s := new(Subscriber)
	s.uids = make(map[int64]int)
	s.room_ids = make(map[int64]int)
	return s
}

type Channel struct {
	addr string
	wt   chan *model.Message

	mutex       sync.Mutex
	subscribers map[int64]*Subscriber

	dispatch       func(*model.AppMessage)
	dispatch_group func(*model.AppMessage)
	dispatch_room  func(*model.AppMessage)
}

func NewChannel(addr string, f func(*model.AppMessage),
	f2 func(*model.AppMessage), f3 func(*model.AppMessage)) *Channel {
	channel := new(Channel)
	channel.subscribers = make(map[int64]*Subscriber)
	channel.dispatch = f
	channel.dispatch_group = f2
	channel.dispatch_room = f3
	channel.addr = addr
	channel.wt = make(chan *model.Message, 10)
	return channel
}

//返回添加前的计数
func (channel *Channel) AddSubscribe(appid, uid int64, online bool) (int, int) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	subscriber, ok := channel.subscribers[appid]
	if !ok {
		subscriber = NewSubscriber()
		channel.subscribers[appid] = subscriber
	}
	//不存在时count==0
	count := subscriber.uids[uid]

	//低16位表示总数量 高16位表示online的数量
	c1 := count & 0xffff
	c2 := count >> 16 & 0xffff

	if online {
		c2 += 1
	}
	c1 += 1
	subscriber.uids[uid] = c2<<16 | c1
	return count & 0xffff, count >> 16 & 0xffff
}

//返回删除前的计数
func (channel *Channel) RemoveSubscribe(appid, uid int64, online bool) (int, int) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	subscriber, ok := channel.subscribers[appid]
	if !ok {
		return 0, 0
	}

	count, ok := subscriber.uids[uid]
	//低16位表示总数量 高16位表示online的数量
	c1 := count & 0xffff
	c2 := count >> 16 & 0xffff
	if ok {
		if online {
			c2 -= 1
			//assert c2 >= 0
			if c2 < 0 {
				log.Warning("online count < 0")
			}
		}
		c1 -= 1
		if c1 > 0 {
			subscriber.uids[uid] = c2<<16 | c1
		} else {
			delete(subscriber.uids, uid)
		}
	}
	return count & 0xffff, count >> 16 & 0xffff
}

func (channel *Channel) GetAllSubscriber() map[int64]*Subscriber {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	subs := make(map[int64]*Subscriber)
	for appid, s := range channel.subscribers {
		sub := NewSubscriber()
		for uid, c := range s.uids {
			sub.uids[uid] = c
		}

		subs[appid] = sub
	}
	return subs
}

//online表示用户不再接受推送通知(apns, gcm)
func (channel *Channel) Subscribe(appid int64, uid int64, online bool) {
	count, online_count := channel.AddSubscribe(appid, uid, online)
	log.Info("sub count:", count, online_count)
	if count == 0 {
		//新用户上线
		on := 0
		if online {
			on = 1
		}
		id := &model.SubscribeMessage{Appid: appid, Uid: uid, Online: int8(on)}
		msg := &model.Message{Cmd: model.MSG_SUBSCRIBE, Body: id}
		channel.wt <- msg
	} else if online_count == 0 && online {
		//手机端上线
		id := &model.SubscribeMessage{Appid: appid, Uid: uid, Online: 1}
		msg := &model.Message{Cmd: model.MSG_SUBSCRIBE, Body: id}
		channel.wt <- msg
	}
}

func (channel *Channel) Unsubscribe(appid int64, uid int64, online bool) {
	count, online_count := channel.RemoveSubscribe(appid, uid, online)
	log.Info("unsub count:", count, online_count)
	if count == 1 {
		//用户断开全部连接
		id := &model.AppUserID{Appid: appid, Uid: uid}
		msg := &model.Message{Cmd: model.MSG_UNSUBSCRIBE, Body: id}
		channel.wt <- msg
	} else if count > 1 && online_count == 1 && online {
		//手机端断开连接,pc/web端还未断开连接
		id := &model.SubscribeMessage{Appid: appid, Uid: uid, Online: 0}
		msg := &model.Message{Cmd: model.MSG_SUBSCRIBE, Body: id}
		channel.wt <- msg
	}
}

func (channel *Channel) Publish(amsg *model.AppMessage) {
	msg := &model.Message{Cmd: model.MSG_PUBLISH, Body: amsg}
	channel.wt <- msg
}

func (channel *Channel) PublishGroup(amsg *model.AppMessage) {
	msg := &model.Message{Cmd: model.MSG_PUBLISH_GROUP, Body: amsg}
	channel.wt <- msg
}

//返回添加前的计数
func (channel *Channel) AddSubscribeRoom(appid, room_id int64) int {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	subscriber, ok := channel.subscribers[appid]
	if !ok {
		subscriber = NewSubscriber()
		channel.subscribers[appid] = subscriber
	}
	//不存在count==0
	count := subscriber.room_ids[room_id]
	subscriber.room_ids[room_id] = count + 1
	return count
}

//返回删除前的计数
func (channel *Channel) RemoveSubscribeRoom(appid, room_id int64) int {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	subscriber, ok := channel.subscribers[appid]
	if !ok {
		return 0
	}

	count, ok := subscriber.room_ids[room_id]
	if ok {
		if count > 1 {
			subscriber.room_ids[room_id] = count - 1
		} else {
			delete(subscriber.room_ids, room_id)
		}
	}
	return count
}

func (channel *Channel) GetAllRoomSubscriber() []*model.AppRoomID {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	subs := make([]*model.AppRoomID, 0, 100)
	for appid, s := range channel.subscribers {
		for room_id, _ := range s.room_ids {
			id := &model.AppRoomID{Appid: appid, Room_id: room_id}
			subs = append(subs, id)
		}
	}
	return subs
}

func (channel *Channel) SubscribeRoom(appid int64, room_id int64) {
	count := channel.AddSubscribeRoom(appid, room_id)
	log.Info("sub room count:", count)
	if count == 0 {
		id := &model.AppRoomID{Appid: appid, Room_id: room_id}
		msg := &model.Message{Cmd: model.MSG_SUBSCRIBE_ROOM, Body: id}
		channel.wt <- msg
	}
}

func (channel *Channel) UnsubscribeRoom(appid int64, room_id int64) {
	count := channel.RemoveSubscribeRoom(appid, room_id)
	log.Info("unsub room count:", count)
	if count == 1 {
		id := &model.AppRoomID{Appid: appid, Room_id: room_id}
		msg := &model.Message{Cmd: model.MSG_UNSUBSCRIBE_ROOM, Body: id}
		channel.wt <- msg
	}
}

func (channel *Channel) PublishRoom(amsg *model.AppMessage) {
	msg := &model.Message{Cmd: model.MSG_PUBLISH_ROOM, Body: amsg}
	channel.wt <- msg
}

func (channel *Channel) ReSubscribe(conn *net.TCPConn, seq int) int {
	subs := channel.GetAllSubscriber()
	for appid, sub := range subs {
		for uid, count := range sub.uids {
			//低16位表示总数量 高16位表示online的数量
			c2 := count >> 16 & 0xffff
			on := 0
			if c2 > 0 {
				on = 1
			}

			id := &model.SubscribeMessage{Appid: appid, Uid: uid, Online: int8(on)}
			msg := &model.Message{Cmd: model.MSG_SUBSCRIBE, Body: id}

			seq = seq + 1
			msg.Seq = seq
			model.SendMessage(conn, msg)
		}
	}
	return seq
}

func (channel *Channel) ReSubscribeRoom(conn *net.TCPConn, seq int) int {
	subs := channel.GetAllRoomSubscriber()
	for _, id := range subs {
		msg := &model.Message{Cmd: model.MSG_SUBSCRIBE_ROOM, Body: id}
		seq = seq + 1
		msg.Seq = seq
		model.SendMessage(conn, msg)
	}
	return seq
}

func (channel *Channel) RunOnce(conn *net.TCPConn) {
	defer conn.Close()

	closed_ch := make(chan bool)
	seq := 0
	seq = channel.ReSubscribe(conn, seq)
	seq = channel.ReSubscribeRoom(conn, seq)

	go func() {
		for {
			msg := model.ReceiveMessage(conn)
			if msg == nil {
				close(closed_ch)
				return
			}
			log.Info("channel recv message:", model.Command(msg.Cmd))
			if msg.Cmd == model.MSG_PUBLISH {
				amsg := msg.Body.(*model.AppMessage)
				if channel.dispatch != nil {
					channel.dispatch(amsg)
				}
			} else if msg.Cmd == model.MSG_PUBLISH_ROOM {
				amsg := msg.Body.(*model.AppMessage)
				if channel.dispatch_room != nil {
					channel.dispatch_room(amsg)
				}
			} else if msg.Cmd == model.MSG_PUBLISH_GROUP {
				amsg := msg.Body.(*model.AppMessage)
				if channel.dispatch_group != nil {
					channel.dispatch_group(amsg)
				}
			} else {
				log.Error("unknown message cmd:", msg.Cmd)
			}
		}
	}()

	for {
		select {
		case _ = <-closed_ch:
			log.Info("channel closed")
			return
		case msg := <-channel.wt:
			seq = seq + 1
			msg.Seq = seq
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := model.SendMessage(conn, msg)
			if err != nil {
				log.Info("channel send message:", err)
			}
		}
	}
}

func (channel *Channel) Run() {
	nsleep := 100
	for {
		conn, err := net.Dial("tcp", channel.addr)
		if err != nil {
			log.Info("connect route server error:", err)
			nsleep *= 2
			if nsleep > 60*1000 {
				nsleep = 60 * 1000
			}
			log.Info("channel sleep:", nsleep)
			time.Sleep(time.Duration(nsleep) * time.Millisecond)
			continue
		}
		tconn := conn.(*net.TCPConn)
		tconn.SetKeepAlive(true)
		tconn.SetKeepAlivePeriod(time.Duration(10 * 60 * time.Second))
		log.Info("channel connected")
		nsleep = 100
		channel.RunOnce(tconn)
	}
}

func (channel *Channel) Start() {
	go channel.Run()
}
