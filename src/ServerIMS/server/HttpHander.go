package server

import (
	"ServerIMS/model"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"

	log "github.com/golang/glog"
)

//统计协程，请求数量，私聊和群聊数量
func Summary(rw http.ResponseWriter, req *http.Request) {
	obj := make(map[string]interface{})
	obj["goroutine_count"] = runtime.NumGoroutine()
	obj["request_count"] = model.Server_summary.Nrequests
	obj["peer_message_count"] = model.Server_summary.Peer_message_count
	obj["group_message_count"] = model.Server_summary.Group_message_count

	res, err := json.Marshal(obj)
	if err != nil {
		log.Info("json marshal:", err)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	_, err = rw.Write(res)
	if err != nil {
		log.Info("write err:", err)
	}
	return
}

//显示协程到os.stderr屏幕
func Stack(rw http.ResponseWriter, req *http.Request) {
	pprof.Lookup("goroutine").WriteTo(os.Stderr, 1)
	WriteHttpError(200, "success", rw)
}

func WriteHttpError(status int, err string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	obj := make(map[string]interface{})
	meta := make(map[string]interface{})
	meta["code"] = status
	meta["message"] = err
	obj["meta"] = meta
	b, _ := json.Marshal(obj)
	w.WriteHeader(status)
	w.Write(b)
}

func WriteHttpObj(data map[string]interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	obj := make(map[string]interface{})
	obj["data"] = data
	b, _ := json.Marshal(obj)
	w.Write(b)
}
