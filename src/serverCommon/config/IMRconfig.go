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

type RouteConfig struct {
	Listen              string
	Mysqldb_datasource  string
	Redis_address       string
	Redis_password      string
	Redis_db            int
	Is_push_system      bool
	Http_listen_address string
}

func Read_route_cfg(cfg_path string) *RouteConfig {
	config := new(RouteConfig)
	app_cfg := make(map[string]string)
	err := cfg.Load(cfg_path, app_cfg)
	if err != nil {
		log.Fatal(err)
	}

	config.Listen = get_string(app_cfg, "listen")
	config.Mysqldb_datasource = get_string(app_cfg, "mysqldb_source")
	config.Redis_address = get_string(app_cfg, "redis_address")
	config.Redis_password = get_opt_string(app_cfg, "redis_password")
	db := get_opt_int(app_cfg, "redis_db", 0)
	config.Redis_db = int(db)
	config.Is_push_system = get_opt_int(app_cfg, "is_push_system", 0) == 1
	config.Http_listen_address = get_opt_string(app_cfg, "http_listen_address")
	return config
}

func (IMR RouteConfig) GetMysqldbdatasource() string {
	return IMR.Mysqldb_datasource
}

func (IMR RouteConfig) GetRedis_address() string {
	return IMR.Redis_address
}

func (IMR RouteConfig) GetRedis_password() string {
	return IMR.Redis_password
}
