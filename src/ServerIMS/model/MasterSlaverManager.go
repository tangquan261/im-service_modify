package model

//"ServerIMS/model"

var MasterObj *Master

func InitMaster() {
	MasterObj = NewMaster()
	MasterObj.Start()
}

func NewMaster() *Master {
	master := new(Master)
	master.Clients = make(map[*SyncClient]struct{})
	master.Ewt = make(chan *EMessage, 10)
	return master
}

func InitSlaver(mastr_Address string) {

	if len(mastr_Address) > 0 {
		slaver := NewSlaver(mastr_Address)
		go slaver.Run()
	}
}
