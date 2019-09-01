package DB

type ServerItemDB struct {
	ID        int64
	IP        string `gorm:"not null;unique"` //IP:PORT 定位唯一节点
	BrotherIP string //兄弟节点,根节点没有兄弟节点
}

func LoadAll_IMS_DB() []ServerItemDB {

	var modes []ServerItemDB

	err := DB.Model(ServerItemDB{}).Find(&modes).Error
	if err != nil {
		return nil
	}
	return modes
}

func Add_IMS_DB(IP, BrotherIP string) error {

	var model ServerItemDB
	model.IP = IP
	model.BrotherIP = BrotherIP
	return DB.Model(ServerItemDB{}).FirstOrCreate(&model).Error
}
