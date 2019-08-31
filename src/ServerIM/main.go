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

package main

//import "net"
import "fmt"
import "time"
import "runtime"
import "math/rand"

//import "net/http"
import "path"

import (
	CommonModel "ServerCommon/model"
	CommonFilter "ServerCommon/pkg/filter"
	"ServerIM/DB"
	"ServerIM/config"
	"ServerIM/model"
	"ServerIM/server"

	log "github.com/golang/glog"
)

func init() {

	//路由管理
	model.App_route = CommonModel.NewAppRoute()
	//统计消息
	model.NewServerSummary()
}

func main() {

	rand.Seed(time.Now().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	config.InitConfig()

	//屏蔽词初始化
	if len(config.Config.Word_file) > 0 {
		CommonFilter.InitFilter(config.Config.Word_file)
	}

	DB.InitRedis()
	server.ListenRPCClient()

	CommonModel.CreateAndStartGroup(DB.Redis_pool, config.Config)

	model.Group_message_delivers = make([]*model.GroupMessageDeliver, config.Config.Group_deliver_count)
	for i := 0; i < config.Config.Group_deliver_count; i++ {
		q := fmt.Sprintf("q%d", i)
		r := path.Join(config.Config.Pending_root, q)
		deliver := model.NewGroupMessageDeliver(r)
		deliver.Start()
		model.Group_message_delivers[i] = deliver
	}

	go model.ListenRedis()
	go model.SyncKeyService()

	server.ListenClient()
	log.Infof("exit")
}
