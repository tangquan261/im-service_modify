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
	Commonmodel "ServerCommon/model"
	"container/list"
	"net"
	"sync/atomic"
	"time"

	"ServerIM/DB"
	"fmt"

	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
)

type Client struct {
	Connection //必须放在结构体首部
	*PeerClient
	*GroupClient
	*RoomClient
	*CustomerClient
	public_ip int32
}

func NewClient(conn interface{}) *Client {
	client := new(Client)

	//初始化Connection
	client.conn = conn // conn is net.Conn or engineio.Conn

	if net_conn, ok := conn.(net.Conn); ok {
		addr := net_conn.LocalAddr()
		if taddr, ok := addr.(*net.TCPAddr); ok {
			ip4 := taddr.IP.To4()
			client.public_ip = int32(ip4[0])<<24 | int32(ip4[1])<<16 | int32(ip4[2])<<8 | int32(ip4[3])
		}
	}

	client.Wt = make(chan *Commonmodel.Message, 300)
	client.lwt = make(chan int, 1) //only need 1
	//'10'对于用户拥有非常多的超级群，读线程还是有可能会阻塞
	client.pwt = make(chan []*Commonmodel.Message, 10)
	client.messages = list.New()

	atomic.AddInt64(&Server_summary.Nconnections, 1)

	client.PeerClient = &PeerClient{&client.Connection}
	client.GroupClient = &GroupClient{&client.Connection}
	client.RoomClient = &RoomClient{Connection: &client.Connection}
	client.CustomerClient = &CustomerClient{&client.Connection}
	return client
}

func (client *Client) Read() {
	for {
		tc := atomic.LoadInt32(&client.tc)
		if tc > 0 {
			log.Infof("quit read goroutine, client:%d write goroutine blocked", client.uid)
			client.HandleClientClosed()
			break
		}

		t1 := time.Now().Unix()
		msg := client.read()
		t2 := time.Now().Unix()
		if t2-t1 > 6*60 {
			log.Infof("client:%d socket read timeout:%d %d", client.uid, t1, t2)
		}
		if msg == nil {
			client.HandleClientClosed()
			break
		}

		client.HandleMessage(msg)
		t3 := time.Now().Unix()
		if t3-t2 > 2 {
			log.Infof("client:%d handle message is too slow:%d %d", client.uid, t2, t3)
		}
	}
}

func (client *Client) RemoveClient() {

	route := App_route.FindRoute(client.appid)
	if route == nil {
		log.Warning("can't find app route")
		return
	}

	route.(*Route).RemoveClient(client)

	if client.room_id > 0 {
		route.(*Route).RemoveRoomClient(client.room_id, client)
	}
}

func (client *Client) HandleClientClosed() {
	atomic.AddInt64(&Server_summary.Nconnections, -1)
	if client.uid > 0 {
		atomic.AddInt64(&Server_summary.Nclients, -1)
	}
	atomic.StoreInt32(&client.closed, 1)

	client.RemoveClient()

	//quit when write goroutine received
	client.Wt <- nil

	client.RoomClient.Logout()
	client.PeerClient.Logout()
}

func (client *Client) HandleMessage(msg *Commonmodel.Message) {
	log.Info("msg cmd:", Commonmodel.Command(msg.Cmd))
	switch msg.Cmd {
	case Commonmodel.MSG_AUTH_TOKEN:
		client.HandleAuthToken(msg.Body.(*Commonmodel.AuthenticationToken),
			msg.Version)
	case Commonmodel.MSG_ACK:
		client.HandleACK(msg.Body.(*Commonmodel.MessageACK))
	case Commonmodel.MSG_PING:
		client.HandlePing()
	}

	client.PeerClient.HandleMessage(msg)
	client.GroupClient.HandleMessage(msg)
	client.RoomClient.HandleMessage(msg)
	client.CustomerClient.HandleMessage(msg)
}

func (client *Client) AuthToken(token string) (int64, int64, int, bool, error) {
	appid, uid, forbidden, notification_on, err := LoadUserAccessToken(token)

	if err != nil {
		return 0, 0, 0, false, err
	}
	fmt.Println(appid, uid, forbidden, notification_on)
	return appid, uid, forbidden, notification_on, nil
}

func (client *Client) HandleAuthToken(login *Commonmodel.AuthenticationToken, version int) {
	if client.uid > 0 {
		log.Info("repeat login")
		return
	}

	var err error
	appid, uid, fb, on, err := client.AuthToken(login.Token)
	if err != nil {
		log.Infof("auth token:%s err:%s", login.Token, err)
		msg := &Commonmodel.Message{Cmd: Commonmodel.MSG_AUTH_STATUS,
			Version: version, Body: &Commonmodel.AuthenticationStatus{1}}
		client.EnqueueMessage(msg)
		return
	}
	if uid == 0 {
		log.Info("auth token uid==0")
		msg := &Commonmodel.Message{Cmd: Commonmodel.MSG_AUTH_STATUS,
			Version: version, Body: &Commonmodel.AuthenticationStatus{1}}
		client.EnqueueMessage(msg)
		return
	}

	if login.Platform_id != Commonmodel.PLATFORM_WEB && len(login.Device_id) > 0 {
		client.device_ID, err = GetDeviceID(login.Device_id, int(login.Platform_id))
		if err != nil {
			log.Info("auth token uid==0")
			msg := &Commonmodel.Message{Cmd: Commonmodel.MSG_AUTH_STATUS, Version: version, Body: &Commonmodel.AuthenticationStatus{1}}
			client.EnqueueMessage(msg)
			return
		}
	}

	is_mobile := login.Platform_id == Commonmodel.PLATFORM_IOS || login.Platform_id == Commonmodel.PLATFORM_ANDROID
	online := true
	if on && !is_mobile {
		online = false
	}

	client.appid = appid
	client.uid = uid
	client.forbidden = int32(fb)
	client.notification_on = on
	client.online = online
	client.version = version
	client.device_id = login.Device_id
	client.platform_id = login.Platform_id
	client.tm = time.Now()
	log.Infof("auth token:%s appid:%d uid:%d device id:%s:%d forbidden:%d notification on:%t online:%t",
		login.Token, client.appid, client.uid, client.device_id,
		client.device_ID, client.forbidden, client.notification_on, client.online)

	msg := &Commonmodel.Message{Cmd: Commonmodel.MSG_AUTH_STATUS,
		Version: version, Body: &Commonmodel.AuthenticationStatus{0}}
	client.EnqueueMessage(msg)

	client.AddClient()

	client.PeerClient.Login()

	CountDAU(client.appid, client.uid)
	atomic.AddInt64(&Server_summary.Nclients, 1)
}

func (client *Client) AddClient() {
	route := App_route.FindOrAddRoute(client.appid).(*Route)
	route.AddClient(client)
}

func (client *Client) HandlePing() {
	m := &Commonmodel.Message{Cmd: Commonmodel.MSG_PONG}
	client.EnqueueMessage(m)
	if client.uid == 0 {
		log.Warning("client has't been authenticated")
		return
	}
}

func (client *Client) HandleACK(ack *Commonmodel.MessageACK) {
	log.Info("ack:", ack.Seq)
}

//发送等待队列中的消息
func (client *Client) SendMessages(seq int) int {
	var messages *list.List
	client.mutex.Lock()
	if client.messages.Len() == 0 {
		client.mutex.Unlock()
		return seq
	}
	messages = client.messages
	client.messages = list.New()
	client.mutex.Unlock()

	e := messages.Front()
	for e != nil {
		msg := e.Value.(*Commonmodel.Message)
		if msg.Cmd == Commonmodel.MSG_RT || msg.Cmd == Commonmodel.MSG_IM ||
			msg.Cmd == Commonmodel.MSG_GROUP_IM {
			atomic.AddInt64(&Server_summary.Out_message_count, 1)
		}
		seq++
		//以当前客户端所用版本号发送消息
		vmsg := &Commonmodel.Message{msg.Cmd, seq, client.version, msg.Flag, msg.Body}
		client.send(vmsg)

		e = e.Next()
	}
	return seq
}

func (client *Client) Write() {
	seq := 0
	running := true

	//发送在线消息
	for running {
		select {
		case msg := <-client.Wt:
			if msg == nil {
				client.close()
				running = false
				log.Infof("client:%d socket closed", client.uid)
				break
			}
			if msg.Cmd == Commonmodel.MSG_RT || msg.Cmd == Commonmodel.MSG_IM ||
				msg.Cmd == Commonmodel.MSG_GROUP_IM {
				atomic.AddInt64(&Server_summary.Out_message_count, 1)
			}
			seq++

			//以当前客户端所用版本号发送消息
			vmsg := &Commonmodel.Message{msg.Cmd, seq, client.version,
				msg.Flag, msg.Body}
			client.send(vmsg)
		case messages := <-client.pwt:
			for _, msg := range messages {
				if msg.Cmd == Commonmodel.MSG_RT || msg.Cmd == Commonmodel.MSG_IM ||
					msg.Cmd == Commonmodel.MSG_GROUP_IM {
					atomic.AddInt64(&Server_summary.Out_message_count, 1)
				}
				seq++

				//以当前客户端所用版本号发送消息
				vmsg := &Commonmodel.Message{msg.Cmd, seq, client.version,
					msg.Flag, msg.Body}
				client.send(vmsg)
			}
		case <-client.lwt:
			seq = client.SendMessages(seq)
			break

		}
	}

	//等待200ms,避免发送者阻塞
	t := time.After(200 * time.Millisecond)
	running = true
	for running {
		select {
		case <-t:
			running = false
		case <-client.Wt:
			log.Warning("msg is dropped")
		}
	}

	log.Info("write goroutine exit")
}

func (client *Client) Run() {
	go client.Write()
	go client.Read()
}

func GetDeviceID(device_id string, platform_id int) (int64, error) {
	conn := DB.Redis_pool.Get()
	defer conn.Close()
	key := fmt.Sprintf("devices_%s_%d", device_id, platform_id)
	device_ID, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		k := "devices_id"
		device_ID, err = redis.Int64(conn.Do("INCR", k))
		if err != nil {
			return 0, err
		}
		_, err = conn.Do("SET", key, device_ID)
		if err != nil {
			return 0, err
		}
	}
	return device_ID, err
}
