package model

import (
	CommonModel "ServerCommon/model"
	//"ServerIMS/model"
	"bytes"
	"os"

	log "github.com/golang/glog"
)

//主从同步消息
const MSG_STORAGE_SYNC_BEGIN = 220
const MSG_STORAGE_SYNC_MESSAGE = 221
const MSG_STORAGE_SYNC_MESSAGE_BATCH = 222

//内部文件存储使用
//个人消息队列 代替MSG_OFFLINE_V3
const MSG_OFFLINE_V4 = 248

//个人消息队列 代替MSG_OFFLINE_V2
const MSG_OFFLINE_V3 = 249

//个人消息队列 代替MSG_OFFLINE
//deprecated  兼容性
const MSG_OFFLINE_V2 = 250

//im实例使用
const MSG_PENDING_GROUP_MESSAGE = 251

//超级群消息队列
const MSG_GROUP_IM_LIST = 252

//deprecated
const MSG_GROUP_ACK_IN = 253

//deprecated 兼容性
const MSG_OFFLINE = 254

//deprecated
const MSG_ACK_IN = 255

type Storage struct {
	*StorageFile
	*PeerStorage
	*GroupStorage
}

func (storage *Storage) NextMessageID() int64 {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	offset, err := storage.file.Seek(0, os.SEEK_END)
	if err != nil {
		log.Fatalln(err)
	}
	return offset + int64(storage.block_NO)*BLOCK_SIZE
}

func (storage *Storage) execMessage(msg *CommonModel.Message, msgid int64) {
	storage.PeerStorage.execMessage(msg, msgid)
	storage.GroupStorage.execMessage(msg, msgid)
}

func (storage *Storage) ExecMessage(msg *CommonModel.Message, msgid int64) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	storage.execMessage(msg, msgid)
}

func (storage *Storage) SaveSyncMessageBatch(mb *MessageBatch) error {
	id := mb.First_id
	//all message come from one block
	for _, m := range mb.Msgs {
		emsg := &EMessage{id, 0, m}
		buffer := new(bytes.Buffer)
		storage.WriteMessage(buffer, m)
		id += int64(buffer.Len())
		storage.SaveSyncMessage(emsg)
	}

	log.Infof("save batch sync message first id:%d last id:%d\n",
		mb.First_id, mb.Last_id)
	return nil
}

func (storage *Storage) SaveSyncMessage(emsg *EMessage) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	n := storage.getBlockNO(emsg.Msgid)
	o := storage.getBlockOffset(emsg.Msgid)

	if n < storage.block_NO || (n-storage.block_NO) > 1 {
		log.Warning("skip msg:", emsg.Msgid)
		return nil
	}

	if (n - storage.block_NO) == 1 {
		storage.file.Close()
		storage.openWriteFile(n)
	}

	offset, err := storage.file.Seek(0, os.SEEK_END)
	if err != nil {
		log.Fatalln(err)
	}

	if o < int(offset) {
		log.Warning("skip msg:", emsg.Msgid)
		return nil
	} else if o > int(offset) {
		log.Warning("write padding:", o-int(offset))
		padding := make([]byte, o-int(offset))
		_, err = storage.file.Write(padding)
		if err != nil {
			log.Fatal("file write:", err)
		}
	}

	storage.WriteMessage(storage.file, emsg.Msg)
	storage.execMessage(emsg.Msg, emsg.Msgid)
	log.Info("save sync message:", emsg.Msgid)
	return nil
}

func (storage *Storage) LoadSyncMessagesInBackground(cursor int64) chan *MessageBatch {
	c := make(chan *MessageBatch, 10)
	go func() {
		defer close(c)

		block_NO := storage.getBlockNO(cursor)
		offset := storage.getBlockOffset(cursor)

		n := block_NO
		for {
			file := storage.openReadFile(n)
			if file == nil {
				break
			}

			if n == block_NO {
				file_size, err := file.Seek(0, os.SEEK_END)
				if err != nil {
					log.Fatal("seek file err:", err)
					return
				}

				if file_size < int64(offset) {
					break
				}

				_, err = file.Seek(int64(offset), os.SEEK_SET)
				if err != nil {
					log.Info("seek file err:", err)
					break
				}
			} else {
				file_size, err := file.Seek(0, os.SEEK_END)
				if err != nil {
					log.Fatal("seek file err:", err)
					return
				}

				if file_size < int64(offset) {
					break
				}

				_, err = file.Seek(HEADER_SIZE, os.SEEK_SET)
				if err != nil {
					log.Info("seek file err:", err)
					break
				}
			}

			const BATCH_COUNT = 5000
			batch := &MessageBatch{Msgs: make([]*CommonModel.Message, 0, BATCH_COUNT)}
			for {
				position, err := file.Seek(0, os.SEEK_CUR)
				if err != nil {
					log.Info("seek file err:", err)
					break
				}
				msg := storage.ReadMessage(file)
				if msg == nil {
					break
				}
				msgid := storage.getMsgId(n, int(position))
				if batch.First_id == 0 {
					batch.First_id = msgid
				}

				batch.Last_id = msgid
				batch.Msgs = append(batch.Msgs, msg)

				if len(batch.Msgs) >= BATCH_COUNT {
					c <- batch
					batch = &MessageBatch{Msgs: make([]*CommonModel.Message, 0, BATCH_COUNT)}
				}
			}
			if len(batch.Msgs) > 0 {
				c <- batch
			}

			n++
		}

	}()
	return c
}

func (storage *Storage) SaveIndexFileAndExit() {
	storage.flushIndex()
	os.Exit(0)
}

func (storage *Storage) flushIndex() {
	storage.mutex.Lock()
	last_id := storage.last_id
	peer_index := storage.clonePeerIndex()
	group_index := storage.cloneGroupIndex()
	storage.mutex.Unlock()

	storage.savePeerIndex(peer_index)
	storage.saveGroupIndex(group_index)
	storage.last_saved_id = last_id
}

func (storage *Storage) FlushIndex() {
	do_flush := false
	if storage.last_id-storage.last_saved_id > 2*BLOCK_SIZE {
		do_flush = true
	}
	if do_flush {
		storage.flushIndex()
	}
}
