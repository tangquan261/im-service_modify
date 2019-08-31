package model

import (
	//CommonModel "ServerCommon/model"
	"ServerIMS/config"

	//"encoding/binary"

	//"ServerIMS/storage"
	//"bytes"
	//"os"

	log "github.com/golang/glog"
)

//存储对象管理,对应storage_func.go
var StorageObj *Storage

func InitStorage() {
	StorageObj = NewStorage(config.Config.Storage_root)
}

func NewStorage(root string) *Storage {
	f := NewStorageFile(root)
	ps := NewPeerStorage(f)
	gs := NewGroupStorage(f)

	storage := &Storage{f, ps, gs}

	r1 := storage.readPeerIndex()
	r2 := storage.readGroupIndex()
	storage.last_saved_id = storage.last_id

	if r1 {
		storage.repairPeerIndex()
	}
	if r2 {
		storage.repairGroupIndex()
	}

	if !r1 {
		storage.createPeerIndex()
	}
	if !r2 {
		storage.createGroupIndex()
	}

	log.Infof("last id:%d last saved id:%d", storage.last_id, storage.last_saved_id)
	storage.FlushIndex()
	return storage
}
