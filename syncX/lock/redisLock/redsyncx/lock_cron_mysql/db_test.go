package lock_cron_mysql

import (
	"testing"

	db2 "github.com/hgg-6/pkgTool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository/dao"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestDb(t *testing.T) {
	db, err := gorm.Open(mysql.Open("root:root@(localhost:13306)/cron_db"))
	assert.NoError(t, err)
	err = db.AutoMigrate(&db2.CronJob{})
	assert.NoError(t, err)
}
