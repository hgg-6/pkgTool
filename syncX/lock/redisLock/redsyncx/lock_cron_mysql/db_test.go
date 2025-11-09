package lock_cron_mysql

import (
	db2 "gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository/dao"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestDb(t *testing.T) {
	db, err := gorm.Open(mysql.Open("root:root@(localhost:13306)/src_db"))
	assert.NoError(t, err)
	err = db.AutoMigrate(&db2.CronJob{})
	assert.NoError(t, err)
}
