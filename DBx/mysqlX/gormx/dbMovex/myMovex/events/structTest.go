package events

import (
	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex"
)

/*
	=======================
	仅测试用的结构体
	=======================
*/

type TestUser struct {
	Id        int64  `gorm:"primaryKey, autoIncrement"`
	Name      string `gorm:"column:nick_name;type:varchar(128);size:128"`
	Email     string `gorm:"unique"`
	UpdatedAt int64
	Ctime     int64
	Utime     int64
}

func (i TestUser) ID() int64 {
	return i.Id
}

func (i TestUser) CompareTo(dst myMovex.Entity) bool {
	val, ok := dst.(TestUser)
	if !ok {
		return false
	}
	return i == val
}
func (i TestUser) Types() string {
	return "TestUser"
}
