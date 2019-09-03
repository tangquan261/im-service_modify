package model

import (
	"ServerCommon/lru"
	"bytes"
	"encoding/binary"
	"os"

	log "github.com/golang/glog"
)

type UserID struct {
	Appid int64
	Uid   int64
}

type UserIndex struct {
	Last_id       int64
	Last_peer_id  int64
	Last_batch_id int64
	Last_seq_id   int64
}

type GroupID struct {
	Appid int64
	Gid   int64
}

func onFileEvicted(key lru.Key, value interface{}) {
	f := value.(*os.File)
	f.Close()
}

//校验文件结尾是否合法
func checkFile(file_path string) bool {
	file, err := os.Open(file_path)
	if err != nil {
		log.Fatal("open file:", err)
	}

	file_size, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		log.Fatal("seek file")
	}

	if file_size == HEADER_SIZE {
		return true
	}

	if file_size < HEADER_SIZE {
		return false
	}

	_, err = file.Seek(file_size-4, os.SEEK_SET)
	if err != nil {
		log.Fatal("seek file")
	}

	mf := make([]byte, 4)
	n, err := file.Read(mf)
	if err != nil || n != 4 {
		log.Fatal("read file err:", err)
	}
	buffer := bytes.NewBuffer(mf)
	var m int32
	binary.Read(buffer, binary.BigEndian, &m)
	return int(m) == MAGIC
}
