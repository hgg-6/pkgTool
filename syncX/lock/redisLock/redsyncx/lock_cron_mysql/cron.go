package lock_cron_mysql

import (
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CronMysql struct {
	web     *gin.Engine
	db      *gorm.DB
	redSync redsyncx.RedSyncIn

	l logx.Loggerx
}

func NewCronMysql(web *gin.Engine, db *gorm.DB, redSync redsyncx.RedSyncIn, l logx.Loggerx) *CronMysql {
	return &CronMysql{
		web:     web,
		db:      db,
		redSync: redSync,
		l:       l,
	}
}
