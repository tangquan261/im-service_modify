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

package config

import "log"
import "strings"
import "github.com/richmonkey/cfg"

const DEFAULT_GROUP_DELIVER_COUNT = 4

type Config struct {
	Port                  int
	Ssl_port              int
	Mysqldb_datasource    string
	Mysqldb_appdatasource string
	Pending_root          string

	Kefu_appid int64

	Redis_address  string
	Redis_password string
	Redis_db       int

	Http_listen_address string
	Rpc_listen_address  string

	//engine io listen address
	//todo remove engine io
	Socket_io_address string
	Tls_address       string

	Cert_file string
	Key_file  string

	//websocket listen address
	Ws_address  string
	Wss_address string

	Storage_rpc_addrs       []string
	Group_storage_rpc_addrs []string
	Route_addrs             []string
	Group_route_addrs       []string //可选配置项， 超群群的route server

	Group_deliver_count int    //群组消息投递并发数量,默认4
	Word_file           string //关键词字典文件
	Sync_self           bool   //是否同步自己发送的消息
}

func Read_cfg(cfg_path string) *Config {
	config := new(Config)
	app_cfg := make(map[string]string)
	err := cfg.Load(cfg_path, app_cfg)
	if err != nil {
		log.Fatal(err)
	}

	config.Port = int(get_int(app_cfg, "port"))
	config.Ssl_port = int(get_opt_int(app_cfg, "ssl_port", 0))
	config.Http_listen_address = get_string(app_cfg, "http_listen_address")
	config.Rpc_listen_address = get_string(app_cfg, "rpc_listen_address")
	config.Redis_address = get_string(app_cfg, "redis_address")
	config.Redis_password = get_opt_string(app_cfg, "redis_password")
	db := get_opt_int(app_cfg, "redis_db", 0)
	config.Redis_db = int(db)

	config.Pending_root = get_string(app_cfg, "pending_root")
	config.Mysqldb_datasource = get_string(app_cfg, "mysqldb_source")
	config.Socket_io_address = get_string(app_cfg, "socket_io_address")
	config.Tls_address = get_opt_string(app_cfg, "tls_address")

	config.Ws_address = get_string(app_cfg, "ws_address")
	config.Wss_address = get_opt_string(app_cfg, "wss_address")

	config.Cert_file = get_opt_string(app_cfg, "cert_file")
	config.Key_file = get_opt_string(app_cfg, "key_file")

	config.Kefu_appid = get_opt_int(app_cfg, "kefu_appid", 0)

	str := get_string(app_cfg, "storage_rpc_pool")
	array := strings.Split(str, " ")
	config.Storage_rpc_addrs = array
	if len(config.Storage_rpc_addrs) == 0 {
		log.Fatal("storage pool config")
	}

	str = get_opt_string(app_cfg, "group_storage_rpc_pool")
	if str != "" {
		array = strings.Split(str, " ")
		config.Group_storage_rpc_addrs = array
		//check repeat
		for _, addr := range config.Group_storage_rpc_addrs {
			for _, addr2 := range config.Storage_rpc_addrs {
				if addr == addr2 {
					log.Fatal("stroage and group storage address repeat")
				}
			}
		}
	}

	str = get_string(app_cfg, "route_pool")
	array = strings.Split(str, " ")
	config.Route_addrs = array
	if len(config.Route_addrs) == 0 {
		log.Fatal("route pool config")
	}

	str = get_opt_string(app_cfg, "group_route_pool")
	if str != "" {
		array = strings.Split(str, " ")
		config.Group_route_addrs = array

		//check repeat group_route_addrs and route_addrs
		for _, addr := range config.Group_route_addrs {
			for _, addr2 := range config.Route_addrs {
				if addr == addr2 {
					log.Fatal("route and group route repeat")
				}
			}
		}
	}

	config.Group_deliver_count = int(get_opt_int(app_cfg, "group_deliver_count", 0))
	if config.Group_deliver_count == 0 {
		config.Group_deliver_count = DEFAULT_GROUP_DELIVER_COUNT
	}

	config.Word_file = get_opt_string(app_cfg, "word_file")
	config.Sync_self = get_opt_int(app_cfg, "sync_self", 0) != 0
	return config
}

func (IM Config) GetMysqldbdatasource() string {
	return IM.Mysqldb_datasource
}

func (IM Config) GetRedis_address() string {
	return IM.Redis_address
}

func (IM Config) GetRedis_password() string {
	return IM.Redis_password
}
