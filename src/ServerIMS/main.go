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
//保存数据点的管理服务器
package main

import (
	"ServerCommon/pkg/util"
	"ServerIMS/config"
	"ServerIMS/model"
	"ServerIMS/server"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// Signal handler
func waitSignal() error {
	ch := make(chan os.Signal, 1)
	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	for {
		sig := <-ch
		fmt.Println("singal:", sig.String())
		switch sig {
		case syscall.SIGTERM, syscall.SIGINT:
			model.StorageObj.Flush()
			model.StorageObj.SaveIndexFileAndExit()
		}
	}
	return nil // It'll never get here.
}

//flush storage file每隔一秒刷新一次消息文件
func FlushLoop() {
	ticker := time.NewTicker(time.Second * 1)
	for range ticker.C {
		model.StorageObj.Flush()
	}
}

//flush message index每隔5分钟刷新一次索引文件，看源码：
func FlushIndexLoop() {
	//5 min
	ticker := time.NewTicker(time.Minute * 5)
	for range ticker.C {
		model.StorageObj.FlushIndex()
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	//监控统计请求
	model.InitSummary()
	//配置信息
	config.InitConfig()

	util.InitMD5(1)

	//本地存储管理
	model.InitStorage()
	//主DB服务器
	model.InitMaster()

	//从DB服务器
	model.InitSlaver(config.Config.Master_address)

	//刷新storage file，定时处理事件
	go FlushLoop()
	go FlushIndexLoop()
	//信号监听处理事件
	go waitSignal()

	//http处理
	go server.StartHttpServer(config.Config.Http_listen_address)

	//处理GRPC之后的请求发送push消息
	go server.Listen(config.Config.Sync_listen)

	//GRPC 处理，提供远程条用查询，存储
	server.ListenRPCClient(config.Config.Rpc_listen)
}
