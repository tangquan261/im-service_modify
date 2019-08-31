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
import "github.com/richmonkey/cfg"

//超级群离线消息数量限制,超过的部分会被丢弃
const GROUP_OFFLINE_LIMIT = 100

//离线消息返回的数量限制
const OFFLINE_DEFAULT_LIMIT = 3000

const GROUP_OFFLINE_DEFAULT_LIMIT = 0

//unlimit
const OFFLINE_DEFAULT_HARD_LIMIT = 0

type StorageConfig struct {
	Rpc_listen          string
	Storage_root        string
	Kefu_appid          int64
	Http_listen_address string

	Sync_listen    string
	Master_address string
	Is_push_system bool
	Group_limit    int //普通群离线消息的数量限制
	Limit          int //单次离线消息的数量限制
	Hard_limit     int //离线消息总的数量限制
}

func Read_storage_cfg(cfg_path string) *StorageConfig {
	config := new(StorageConfig)
	app_cfg := make(map[string]string)
	err := cfg.Load(cfg_path, app_cfg)
	if err != nil {
		log.Fatal(err)
	}

	config.Rpc_listen = get_string(app_cfg, "rpc_listen")
	config.Http_listen_address = get_opt_string(app_cfg, "http_listen_address")
	config.Storage_root = get_string(app_cfg, "storage_root")
	config.Kefu_appid = get_int(app_cfg, "kefu_appid")
	config.Sync_listen = get_string(app_cfg, "sync_listen")
	config.Master_address = get_opt_string(app_cfg, "master_address")
	config.Is_push_system = get_opt_int(app_cfg, "is_push_system", 0) == 1
	config.Limit = int(get_opt_int(app_cfg, "limit", OFFLINE_DEFAULT_LIMIT))
	config.Group_limit = int(get_opt_int(app_cfg, "group_limit", GROUP_OFFLINE_DEFAULT_LIMIT))
	config.Hard_limit = int(get_opt_int(app_cfg, "hard_limit", OFFLINE_DEFAULT_HARD_LIMIT))
	return config
}
