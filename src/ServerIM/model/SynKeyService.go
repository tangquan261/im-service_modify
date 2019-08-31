package model

import (
	log "github.com/golang/glog"
)

func SyncKeyService() {
	for {
		select {
		case s := <-Sync_c:
			origin := GetSyncKey(s.AppID, s.Uid)
			if s.LastMsgID > origin {
				log.Infof("save sync key:%d %d %d", s.AppID, s.Uid, s.LastMsgID)
				SaveSyncKey(s.AppID, s.Uid, s.LastMsgID)
			}
			break
		case s := <-Group_sync_c:
			origin := GetGroupSyncKey(s.AppID, s.Uid, s.GroupID)
			if s.LastMsgID > origin {
				log.Infof("save group sync key:%d %d %d %d",
					s.AppID, s.Uid, s.GroupID, s.LastMsgID)
				SaveGroupSyncKey(s.AppID, s.Uid, s.GroupID, s.LastMsgID)
			}
			break
		}
	}
}
