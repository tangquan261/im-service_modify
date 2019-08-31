package config

import (
	IMSconfig "ServerCommon/config"

	log "github.com/golang/glog"
)

var Config *IMSconfig.StorageConfig

func InitConfig() {
	Config = IMSconfig.Read_storage_cfg("./ims.cfg")
	log.Infof("rpc listen:%s storage root:%s sync listen:%s master address:%s is push system:%t group limit:%d offline message limit:%d hard limit:%d\n",
		Config.Rpc_listen, Config.Storage_root, Config.Sync_listen,
		Config.Master_address, Config.Is_push_system, Config.Group_limit,
		Config.Limit, Config.Hard_limit)
	log.Infof("http listen address:%s", Config.Http_listen_address)

	if Config.Limit == 0 {
		log.Fatal("config limit is 0")
		return
	}
	if Config.Hard_limit > 0 && Config.Hard_limit/Config.Limit < 2 {
		log.Fatal("config limit:%d, hard limit:%d invalid, hard limit/limit must gte 2",
			Config.Limit, Config.Hard_limit)
		return
	}
}
