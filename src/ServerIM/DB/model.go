package DB

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var (
	DB *gorm.DB
)

func init() {

}

func ConfigDB() {

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "tq_" + defaultTableName
	}

	var err error
	DB, err = gorm.Open("mysql", "root:tangquan@tcp(127.0.0.1:3306)/gobelieve?charset=utf8&parseTime=True&loc=Local")

	if nil != err {
		panic(err)
	}

	DB.DB().SetMaxIdleConns(10)
	DB.DB().SetMaxOpenConns(100)
	DB.LogMode(true)
	DB.SingularTable(true)
	DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&ServerItemDB{})

	fmt.Println("db init success")
}
