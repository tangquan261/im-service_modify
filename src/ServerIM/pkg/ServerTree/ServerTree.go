package ServerTree

import (
	"ServerIM/DB"
	"fmt"
	"log"
)

var ServerItemMap map[string]*ServerItem
var HeadNodeList *ServerItem

type ServerItem struct {
	IP     string //不存在IP的是虚拟节点
	Left   *ServerItem
	Right  *ServerItem
	Parent *ServerItem
}

func init() {
	ServerItemMap = make(map[string]*ServerItem)

	HeadNodeList = new(ServerItem)
	HeadNodeList.Parent = nil
	HeadNodeList.Left = nil
	HeadNodeList.Right = nil
	HeadNodeList.IP = ""

}

func InitLoadServerIM() {
	models := DB.LoadAll_IMS_DB()

	for i := 0; i < len(models); i++ {
		nodeDB := models[i]
		node := new(ServerItem)
		node.IP = nodeDB.IP

		if !AddServerItem(nodeDB.BrotherIP, node) {
			log.Fatal("error add node ", nodeDB.BrotherIP)
		}
	}
}

func AddServerItem(brotherIP string, Newitem *ServerItem) bool {

	if len(Newitem.IP) <= 0 {
		fmt.Println("插入的节点不能是虚拟节点")
		return false
	}
	if len(brotherIP) <= 0 {

		if HeadNodeList.Left == nil {
			HeadNodeList.Left = Newitem
			Newitem.Parent = HeadNodeList
			ServerItemMap[Newitem.IP] = Newitem
			return true
		}

		if HeadNodeList.Right == nil {
			HeadNodeList.Right = Newitem
			Newitem.Parent = HeadNodeList
			ServerItemMap[Newitem.IP] = Newitem
			return true
		}

		fmt.Println("根节点插入失败!", Newitem.IP)
		return false
	}

	var brotherItem *ServerItem
	var ok bool

	if brotherItem, ok = ServerItemMap[brotherIP]; !ok {
		fmt.Println("没有找到 brothIP 的节点", brotherIP)
		return false
	}

	if len(brotherItem.IP) <= 0 {
		fmt.Println("兄弟 已经是虚拟节点，不能继续添加节点")
		return false
	}

	if brotherItem.Parent == nil {
		fmt.Println("兄弟节点没有父亲节点")
		return false
	}

	NewVNode := new(ServerItem)
	NewVNode.Parent = brotherItem.Parent
	NewVNode.Left = brotherItem
	NewVNode.Right = Newitem

	brotherItem.Parent = NewVNode
	Newitem.Parent = NewVNode

	ServerItemMap[Newitem.IP] = Newitem
	return true
}

func FindServer(GUID int64) *ServerItem {

	return findServerItem(1, GUID, HeadNodeList)
}

//GUID 为两个用户ID只和，计算的出的服务器
func findServerItem(depth int64, GUID int64, parent *ServerItem) *ServerItem {

	if len(parent.IP) < 0 {
		//是虚拟节点，则二分到下一节点
		if (GUID/depth)%2 == 0 {
			//左边节点
			return findServerItem(depth*2, GUID, parent.Left)
		} else {
			//右边节点
			return findServerItem(depth*2, GUID, parent.Right)
		}

		return nil
	} else {
		//是实体节点，直接返回该节点
		return parent
	}
}
