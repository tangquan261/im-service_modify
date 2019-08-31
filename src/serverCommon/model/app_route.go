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

//路由管理
func NewAppRoute() *AppRoute {
	app_route := new(AppRoute)
	app_route.apps = make(map[int64]RouteBase)
	return app_route
}

type RouteBase interface {
	GetAppID() int64
	GetUserIDs() IntSet
	ContainUserID(uid int64) bool
	IsUserOnline(uid int64) bool
	ContainRoomID(room_id int64) bool
	AddUserID(uid int64, online bool)
	RemoveUserID(uid int64)
	AddRoomID(room_id int64)
	RemoveRoomID(room_id int64)
}

type AppRoute struct {
	mutex    sync.Mutex
	apps     map[int64]RouteBase
	FnCreate func(appid int64) RouteBase //创建route
}

//根据id获取路由，不在路由中则创建路由加入路由管理，并返回创建的路由
func (app_route *AppRoute) FindOrAddRoute(appid int64) RouteBase {
	app_route.mutex.Lock()
	defer app_route.mutex.Unlock()
	if route, ok := app_route.apps[appid]; ok {
		return route
	}

	route := app_route.FnCreate(appid)
	app_route.apps[appid] = route
	return route
}

//根据id获取路由
func (app_route *AppRoute) FindRoute(appid int64) RouteBase {
	app_route.mutex.Lock()
	defer app_route.mutex.Unlock()
	return app_route.apps[appid]
}

//将路由加入路由管理
func (app_route *AppRoute) AddRoute(route RouteBase) {
	app_route.mutex.Lock()
	defer app_route.mutex.Unlock()
	app_route.apps[route.GetAppID()] = route
}

//获取路由管理下的所有appid对应的用户id 的intset
func (app_route *AppRoute) GetUsers() map[int64]IntSet {
	app_route.mutex.Lock()
	defer app_route.mutex.Unlock()

	r := make(map[int64]IntSet)
	for appid, route := range app_route.apps {
		uids := route.GetUserIDs()
		r[appid] = uids
	}
	return r
}
