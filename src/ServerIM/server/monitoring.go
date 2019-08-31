package server

import "net/http"
import "encoding/json"
import "os"
import "runtime"
import "runtime/pprof"
import (
	"ServerIM/model"

	log "github.com/golang/glog"
)

func Summary(rw http.ResponseWriter, req *http.Request) {
	obj := make(map[string]interface{})
	obj["goroutine_count"] = runtime.NumGoroutine()
	obj["connection_count"] = model.Server_summary.Nconnections
	obj["client_count"] = model.Server_summary.Nclients
	obj["in_message_count"] = model.Server_summary.In_message_count
	obj["out_message_count"] = model.Server_summary.Out_message_count

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

func Stack(rw http.ResponseWriter, req *http.Request) {
	pprof.Lookup("goroutine").WriteTo(os.Stderr, 1)
	rw.WriteHeader(200)
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
