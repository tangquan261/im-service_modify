package util

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"strconv"
	"time"

	"github.com/satori/go.uuid"
	"github.com/zheng-ji/goSnowFlake"
	//Snowflake 算法是Twitter的分布式ID自增算法,用于生成可以跨数据中心的全局唯一ID(不连续)。
)

var (
	iw     *goSnowFlake.IdWorker
	iwSnow *goSnowFlake.IdWorker
)

func InitMD5(snow_work_id int64) {
	var err error

	iw, err = goSnowFlake.NewIdWorker(snow_work_id)
	if err != nil {
		log.Panic("error init goSnowFlake")
	}

	iwSnow, err = goSnowFlake.NewIdWorker(snow_work_id)
	if err != nil {
		log.Panic("error init goSnowFlake")
	}
}

func EncodeMD5(value string) string {
	m := md5.New()
	m.Write([]byte(value))
	return hex.EncodeToString(m.Sum(nil))
}

func Uids() string {
	var ret string

	bytes, err := uuid.NewV4()

	if err == nil {
		return bytes.String()
	}

	u1, err := iw.NextId()

	if err != nil {
		return EncodeMD5("tq" + time.Now().String())
	}

	ret = strconv.FormatInt(u1, 10)

	return ret
}

func UUID() int64 {

	id, err := iw.NextId()
	if err != nil {
		return 0
	}
	return id

}

func SnowFlakeUUID() int64 {
	id, err := iw.NextId()
	if err != nil {
		return 0
	}
	return id
}
