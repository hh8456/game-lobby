package login

import (
	"servers/model"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func TestLogin(t *testing.T) {
	db, err := gorm.Open("mysql", "dev:dev123@tcp(192.168.0.155)/games?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		t.Logf("连接数据库错误: %v\n", err)
		return
	}
	t.Log("连接数据库成功")

	defer db.Close()

	db.SingularTable(true)

	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(20)

	type PlayerBaseInfo struct {
		ID         int64
		Wxid       string
		WxidCrc32  int
		UId        int
		HeadPic    string
		InviteCode int
		Diamond    int
		Gold       int
		Sex        int8
		RegDate    time.Time
	}

	playerBaseInfo := &model.PlayerBaseInfo{}
	//playerBaseInfo := &model.PlayerBaseInfo{Wxid: "111", WxidCrc32: 3,
	//UId: int(time.Now().Unix()), HeadPic: "1", InviteCode: 33,
	//Diamond: 100000000, Gold: 100000000, RegDate: time.Now()}

	if err := db.Where("wxid_crc32 = ? and wxid = ?", 333, "123").First(playerBaseInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			t.Logf("没有查询到结果")
		} else {
			t.Logf("查询错误: %v\n", err)
		}
	}
}
