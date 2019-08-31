package server

import (
	"net/http"

	log "github.com/golang/glog"
)

func StartHttpServer(addr string) {

	if len(addr) <= 0 {
		return
	}

	//统计协程，请求数量，私聊和群聊数量
	http.HandleFunc("/summary", Summary)
	http.HandleFunc("/stack", Stack)

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
