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

import "sync"
import "strconv"
import "strings"
import "time"
import "errors"
import "fmt"
import "math/rand"
import "github.com/gomodule/redigo/redis"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import log "github.com/golang/glog"
import "ServerCommon/config"

//同redis的长链接保持5minute的心跳
const SUBSCRIBE_HEATBEAT = 5 * 60

//群聊房间管理
type GroupManager struct {
	mutex      sync.Mutex
	groups     map[int64]*Group
	ping       string
	action_id  int64
	dirty      bool
	redis_pool *redis.Pool
	config     config.ConfigBase
}

var Group_manager *GroupManager

func CreateAndStartGroup(redis_pool *redis.Pool, config config.ConfigBase) {
	Group_manager = NewGroupManager(redis_pool, config)
	Group_manager.Start()
}

func NewGroupManager(redis_pool *redis.Pool, config config.ConfigBase) *GroupManager {
	now := time.Now().Unix()
	r := fmt.Sprintf("ping_%d", now)
	for i := 0; i < 4; i++ {
		n := rand.Int31n(26)
		r = r + string('a'+n)
	}

	m := new(GroupManager)
	m.config = config
	m.ping = r
	m.action_id = 0
	m.dirty = true
	m.redis_pool = redis_pool
	m.groups = make(map[int64]*Group)
	return m
}

func (g *GroupManager) Start() {
	g.load()
	go g.Run()      //循环等待接收消息处理
	go g.PingLoop() //循环ping
}

func (g *GroupManager) GetGroups() []*Group {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	groups := make([]*Group, 0, len(g.groups))
	for _, group := range g.groups {
		groups = append(groups, group)
	}
	return groups
}

func (g *GroupManager) FindGroup(gid int64) *Group {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if group, ok := g.groups[gid]; ok {
		return group
	}
	return nil
}

func (g *GroupManager) FindUserGroups(appid int64, uid int64) []*Group {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	groups := make([]*Group, 0, 4)
	for _, group := range g.groups {
		if group.Appid == appid && group.IsMember(uid) {
			groups = append(groups, group)
		}
	}
	return groups
}

func (g *GroupManager) HandleCreate(data string) {
	arr := strings.Split(data, ",")
	if len(arr) != 3 {
		log.Info("message error:", data)
		return
	}
	gid, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}
	appid, err := strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}
	super, err := strconv.ParseInt(arr[2], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, ok := g.groups[gid]; ok {
		log.Infof("group:%d exists\n", gid)
	}
	log.Infof("create group:%d appid:%d", gid, appid)
	if super != 0 {
		g.groups[gid] = NewSuperGroup(gid, appid, nil)
	} else {
		g.groups[gid] = NewGroup(gid, appid, nil)
	}
}

func (g *GroupManager) HandleDisband(data string) {
	gid, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()
	if _, ok := g.groups[gid]; ok {
		log.Info("disband group:", gid)
		delete(g.groups, gid)
	} else {
		log.Infof("group:%d nonexists\n", gid)
	}
}

func (g *GroupManager) HandleUpgrade(data string) {
	arr := strings.Split(data, ",")
	if len(arr) != 3 {
		log.Info("message error:", data)
		return
	}
	gid, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}
	appid, err := strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}
	super, err := strconv.ParseInt(arr[2], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}

	if super == 0 {
		log.Warning("super group can't transfer to nomal group")
		return
	}
	group := g.FindGroup(gid)
	if group != nil {
		group.Super = (super == 1)
		log.Infof("upgrade group appid:%d gid:%d super:%d", appid, gid, super)
	} else {
		log.Infof("can't find group:%d\n", gid)
	}
}

func (g *GroupManager) HandleMemberAdd(data string) {
	arr := strings.Split(data, ",")
	if len(arr) != 2 {
		log.Info("message error")
		return
	}
	gid, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}
	uid, err := strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}

	group := g.FindGroup(gid)
	if group != nil {
		timestamp := int(time.Now().Unix())
		group.AddMember(uid, timestamp)
		log.Infof("add group member gid:%d uid:%d", gid, uid)
	} else {
		log.Infof("can't find group:%d\n", gid)
	}
}

func (g *GroupManager) HandleMemberRemove(data string) {
	arr := strings.Split(data, ",")
	if len(arr) != 2 {
		log.Info("message error")
		return
	}
	gid, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}
	uid, err := strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}

	group := g.FindGroup(gid)
	if group != nil {
		group.RemoveMember(uid)
		log.Infof("remove group member gid:%d uid:%d", gid, uid)
	} else {
		log.Infof("can't find group:%d\n", gid)
	}
}

func (g *GroupManager) HandleMute(data string) {
	arr := strings.Split(data, ",")
	if len(arr) != 3 {
		log.Info("message error:", data)
		return
	}
	gid, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}
	uid, err := strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}
	mute, err := strconv.ParseInt(arr[2], 10, 64)
	if err != nil {
		log.Info("error:", err)
		return
	}

	group := g.FindGroup(gid)
	if group != nil {
		group.SetMemberMute(uid, mute != 0)
		log.Infof("set group member gid:%d uid:%d mute:%d", gid, uid, mute)
	} else {
		log.Infof("can't find group:%d\n", gid)
	}
}

//保证action id的顺序性
func (g *GroupManager) parseAction(data string) (bool, int64, int64, string) {
	arr := strings.SplitN(data, ":", 3)
	if len(arr) != 3 {
		log.Warning("group action error:", data)
		return false, 0, 0, ""
	}

	prev_id, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil {
		log.Info("error:", err, data)
		return false, 0, 0, ""
	}

	action_id, err := strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		log.Info("error:", err, data)
		return false, 0, 0, ""
	}
	return true, prev_id, action_id, arr[2]
}

func (g *GroupManager) handleAction(data string, channel string) {
	r, prev_id, action_id, content := g.parseAction(data)
	if r {
		log.Info("group action:", prev_id, action_id, g.action_id, " ", channel)
		if g.action_id != prev_id {
			//reload later
			g.dirty = true
			log.Warning("action nonsequence:",
				g.action_id, prev_id, action_id)
		}

		if channel == "group_create" {
			g.HandleCreate(content)
		} else if channel == "group_disband" {
			g.HandleDisband(content)
		} else if channel == "group_member_add" {
			g.HandleMemberAdd(content)
		} else if channel == "group_member_remove" {
			g.HandleMemberRemove(content)
		} else if channel == "group_upgrade" {
			g.HandleUpgrade(content)
		} else if channel == "group_member_mute" {
			g.HandleMute(content)
		}
		g.action_id = action_id
	}
}

func (g *GroupManager) ReloadGroup() bool {
	log.Info("reload group...")
	db, err := sql.Open("mysql", g.config.GetMysqldbdatasource())
	if err != nil {
		log.Info("error:", err)
		return false
	}
	defer db.Close()

	groups, err := LoadAllGroup(db)
	if err != nil {
		log.Info("load all group error:", err)
		return false
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.groups = groups

	return true
}

func (g *GroupManager) getActionID() (int64, error) {
	conn := g.redis_pool.Get()
	defer conn.Close()

	actions, err := redis.String(conn.Do("GET", "groups_actions"))
	if err != nil && err != redis.ErrNil {
		log.Info("hget error:", err)
		return 0, err
	}
	if actions == "" {
		return 0, nil
	} else {
		arr := strings.Split(actions, ":")
		if len(arr) != 2 {
			log.Error("groups_actions invalid:", actions)
			return 0, errors.New("groups actions invalid")
		}
		_, err := strconv.ParseInt(arr[0], 10, 64)
		if err != nil {
			log.Info("error:", err, actions)
			return 0, err
		}

		action_id, err := strconv.ParseInt(arr[1], 10, 64)
		if err != nil {
			log.Info("error:", err, actions)
			return 0, err
		}
		return action_id, nil
	}
}

func (g *GroupManager) load() {
	//循环直到成功
	for {
		action_id, err := g.getActionID()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		r := g.ReloadGroup()
		if !r {
			time.Sleep(1 * time.Second)
			continue
		}

		g.action_id = action_id
		g.dirty = false
		log.Info("group action id:", action_id)
		break
	}
}

//检查当前的action id 是否变更，变更时则重新加载群组结构
func (g *GroupManager) checkActionID() {
	action_id, err := g.getActionID()
	if err != nil {
		//load later
		g.dirty = true
		return
	}

	if action_id != g.action_id {
		r := g.ReloadGroup()
		if r {
			g.dirty = false
			g.action_id = action_id
		} else {
			//load later
			g.dirty = true
		}
	}
}

func (g *GroupManager) RunOnce() bool {
	t := redis.DialReadTimeout(time.Second * SUBSCRIBE_HEATBEAT)
	c, err := redis.Dial("tcp", g.config.GetRedis_address(), t)
	if err != nil {
		log.Info("dial redis error:", err)
		return false
	}

	password := g.config.GetRedis_password()
	if len(password) > 0 {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			return false
		}
	}

	psc := redis.PubSubConn{c}
	psc.Subscribe("group_create", "group_disband", "group_member_add",
		"group_member_remove", "group_upgrade", "group_member_mute",
		g.ping)

	g.checkActionID()
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			if v.Channel == "group_create" ||
				v.Channel == "group_disband" ||
				v.Channel == "group_member_add" ||
				v.Channel == "group_member_remove" ||
				v.Channel == "group_upgrade" ||
				v.Channel == "group_member_mute" {
				g.handleAction(string(v.Data), v.Channel)
			} else if v.Channel == g.ping {
				//check dirty
				if g.dirty {
					action_id, err := g.getActionID()
					if err == nil {
						r := g.ReloadGroup()
						if r {
							g.dirty = false
							g.action_id = action_id
						}
					} else {
						log.Warning("get action id err:", err)
					}
				} else {
					g.checkActionID()
				}
				log.Info("group manager dirty:", g.dirty)
			} else {
				log.Infof("%s: message: %s\n", v.Channel, v.Data)
			}
		case redis.Subscription:
			log.Infof("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			log.Info("error:", v)
			return true
		}
	}
}

func (g *GroupManager) Run() {
	nsleep := 1
	for {
		connected := g.RunOnce()
		if !connected {
			nsleep *= 2
			if nsleep > 60 {
				nsleep = 60
			}
		} else {
			nsleep = 1
		}
		time.Sleep(time.Duration(nsleep) * time.Second)
	}
}

func (g *GroupManager) Ping() {
	conn := g.redis_pool.Get()
	defer conn.Close()

	_, err := conn.Do("PUBLISH", g.ping, "ping")
	if err != nil {
		log.Info("ping error:", err)
	}
}

func (g *GroupManager) PingLoop() {
	for {
		g.Ping()
		time.Sleep(time.Second * (SUBSCRIBE_HEATBEAT - 10))
	}
}
