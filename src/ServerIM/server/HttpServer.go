package server

import (
	"net/http"

	"ServerIM/config"

	log "github.com/golang/glog"
)

func InitHttpServer() {
	if len(config.Config.Http_listen_address) > 0 {
		go StartHttpServer(config.Config.Http_listen_address)
	}
}

func StartHttpServer(addr string) {
	http.HandleFunc("/summary", Summary)
	http.HandleFunc("/stack", Stack)

	//rpc function
	http.HandleFunc("/post_group_notification", PostGroupNotification)
	http.HandleFunc("/post_im_message", PostIMMessage)
	http.HandleFunc("/load_latest_message", LoadLatestMessage)
	http.HandleFunc("/load_history_message", LoadHistoryMessage)
	http.HandleFunc("/post_system_message", SendSystemMessage)
	http.HandleFunc("/post_notification", SendNotification)
	http.HandleFunc("/post_room_message", SendRoomMessage)
	http.HandleFunc("/post_customer_message", SendCustomerMessage)
	http.HandleFunc("/post_customer_support_message", SendCustomerSupportMessage)
	http.HandleFunc("/post_realtime_message", SendRealtimeMessage)
	http.HandleFunc("/init_message_queue", InitMessageQueue)
	http.HandleFunc("/get_offline_count", GetOfflineCount)
	http.HandleFunc("/dequeue_message", DequeueMessage)

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
