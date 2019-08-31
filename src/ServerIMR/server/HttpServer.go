package server

import (
	"net/http"

	"ServerIMR/config"

	log "github.com/golang/glog"
)

func InitHttpServer() {
	if len(config.Config.Http_listen_address) > 0 {
		go StartHttpServer(config.Config.Http_listen_address)
	}
}

func StartHttpServer(addr string) {
	http.HandleFunc("/online", GetOnlineStatus)
	http.HandleFunc("/all_online", GetOnlineClients)

	handler := loggingHandler{http.DefaultServeMux}

	err := http.ListenAndServe(addr, handler)
	if err != nil {
		log.Fatal("http server err:", err)
	}
}

type loggingHandler struct {
	handler http.Handler
}

func (h loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infof("http request:%s %s %s", r.RemoteAddr, r.Method, r.URL)
	h.handler.ServeHTTP(w, r)
}
