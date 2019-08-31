package config

import (
	IMRconfig "ServerCommon/config"

	log "github.com/golang/glog"
)

var Config *IMRconfig.RouteConfig

func InitConfig() {
	Config = IMRconfig.Read_route_cfg("./imr.cfg")
	log.Infof("listen:%s\n", Config.Listen)

	log.Infof("redis address:%s password:%s db:%d\n",
		Config.Redis_address, Config.Redis_password, Config.Redis_db)
}
