package main

import (
	"math/rand"
	"runtime"
	"time"

	CommonModel "ServerCommon/model"
	"ServerIMR/DB"
	"ServerIMR/config"
	"ServerIMR/server"
)

//内网路由服务器,处理ServerIM消息
func main() {
	rand.Seed(time.Now().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	config.InitConfig()

	DB.InitRedis()

	//通过redis监听redis的message进行群聊group的内存级管理
	CommonModel.CreateAndStartGroup(DB.Redis_pool, config.Config)
	//http管理，返回用户的在线信息
	server.InitHttpServer()
	//作为转发服务器，等待IM连接，处理数据
	server.ListenClient()
}
