package config

import (
	IMconfig "ServerCommon/config"

	log "github.com/golang/glog"
)

var Config *IMconfig.Config

func InitConfig() {
	Config = IMconfig.Read_cfg("/Users/tq/gowork/src/ServerIM/im.cfg")
	log.Infof("port:%d\n", Config.Port)

	log.Infof("redis address:%s password:%s db:%d\n",
		Config.Redis_address, Config.Redis_password, Config.Redis_db)
	log.Info("storage addresses:", Config.Storage_rpc_addrs)
	log.Info("route addressed:", Config.Route_addrs)
	log.Info("group route addressed:", Config.Group_route_addrs)
	log.Info("kefu appid:", Config.Kefu_appid)
	log.Info("pending root:", Config.Pending_root)

	log.Infof("socket io address:%s tls_address:%s cert file:%s key file:%s",
		Config.Socket_io_address, Config.Tls_address, Config.Cert_file, Config.Key_file)
	log.Infof("ws address:%s wss address:%s", Config.Ws_address, Config.Wss_address)
	log.Info("group deliver count:", Config.Group_deliver_count)
	log.Info("sync self:", Config.Sync_self)
}
