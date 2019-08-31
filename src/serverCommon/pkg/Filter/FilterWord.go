package filter

import (
	"github.com/bitly/go-simplejson"
	"github.com/importcjj/sensitive"

	"ServerCommon/model"

	log "github.com/golang/glog"
)

var Filter *sensitive.Filter

func InitFilter(word_fild string) {
	if len(word_fild) > 0 {
		Filter = sensitive.New()
		Filter.LoadWordDict(word_fild)
	}
}

//过滤敏感词
func FilterDirtyWord(msg *model.IMMessage) {
	if Filter == nil {
		return
	}

	obj, err := simplejson.NewJson([]byte(msg.Content))
	if err != nil {
		return
	}

	text, err := obj.Get("text").String()
	if err != nil {
		return
	}

	if exist, _ := Filter.FindIn(text); exist {
		t := Filter.RemoveNoise(text)
		replacedText := Filter.Replace(t, '*')

		obj.Set("text", replacedText)
		c, err := obj.Encode()
		if err != nil {
			log.Errorf("json encode err:%s", err)
			return
		}
		msg.Content = string(c)
	}
}
