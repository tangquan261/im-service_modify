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
import "ServerCommon/model"

//路由本身有routemanger管理
type Route struct {
	appid    int64 //每个服务器IM的id
	mutex    sync.Mutex
	uids     map[int64]bool
	room_ids model.IntSet
}

func NewRoute(appid int64) model.RouteBase {
	r := new(Route)
	r.appid = appid
	r.uids = make(map[int64]bool)
	r.room_ids = model.NewIntSet()
	return r
}

func (route *Route) GetAppID() int64 {
	return route.appid
}

//判断用户是否在该路由下
func (route *Route) ContainUserID(uid int64) bool {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	_, ok := route.uids[uid]
	return ok
}

//判断用户是否在线
func (route *Route) IsUserOnline(uid int64) bool {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	return route.uids[uid]
}

//将用户加入该路由
func (route *Route) AddUserID(uid int64, online bool) {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	route.uids[uid] = online
}

func (route *Route) RemoveUserID(uid int64) {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	delete(route.uids, uid)
}

//获取该路由下所有用户
func (route *Route) GetUserIDs() model.IntSet {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	uids := model.NewIntSet()
	for uid, _ := range route.uids {
		uids.Add(uid)
	}
	return uids
}

//判断该路由下是否包含该房间
func (route *Route) ContainRoomID(room_id int64) bool {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	return route.room_ids.IsMember(room_id)
}

//将当前房间加入本路由
func (route *Route) AddRoomID(room_id int64) {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	route.room_ids.Add(room_id)
}

//将当前房间移除本路由
func (route *Route) RemoveRoomID(room_id int64) {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	route.room_ids.Remove(room_id)
}
