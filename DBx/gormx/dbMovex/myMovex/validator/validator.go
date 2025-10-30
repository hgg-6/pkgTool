package validator

/*
	===========================================
	校验器：此文件主要用来处理数据验证、不一致数据处理、迁移的
	【根据数据库中表的Id字段，亦为自增主键，所以表中默认需有Id自增主键】
	===========================================
*/

import (
	"context"
	"encoding/json"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/gormx/dbMovex/myMovex"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/gormx/dbMovex/myMovex/events"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/slicex"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

type MessageQueueStr[Pdr any] struct {
	//Producer          messageQueuex.ProducerIn[Pdr]
	Producer          mqX.Producer
	MessageQueueTopic string
}

type Validator[T myMovex.Entity, Pdr any] struct {
	// 数据迁移，肯定有
	base   *gorm.DB // 源数据库
	target *gorm.DB // 目标数据库

	l logx.Loggerx

	//producer messageQueuex.ProducerIn[Pdr]
	MessageQueueConf *MessageQueueStr[Pdr]
	topic            string

	direction string
	batchSize int
	utime     int64
	// <= 0 就认为中断
	// > 0 就认为睡眠
	sleepInterval time.Duration
	fromBase      func(ctx context.Context, offset int) (T, error)
}

func NewValidator[T myMovex.Entity, Pdr any](base *gorm.DB, target *gorm.DB, direction string, l logx.Loggerx, producerConf *MessageQueueStr[Pdr]) *Validator[T, Pdr] {
	res := &Validator[T, Pdr]{
		base:             base,
		target:           target,
		l:                l,
		MessageQueueConf: producerConf,
		direction:        direction,
		batchSize:        100,
	}
	res.fromBase = res.fullFromBase
	return res
}

// Validate 验证
func (v *Validator[T, Pdr]) Validate(ctx context.Context) error {
	//err := v.validateBaseToTarget(ctx)
	//if err != nil {
	//	return err
	//}
	//return v.validateTargetToBase(ctx)

	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBaseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.validateTargetToBase(ctx)
	})
	return eg.Wait()
}

// target -> base
func (v *Validator[T, Pdr]) validateBaseToTarget(ctx context.Context) error {
	offset := 0
	for {
		src, err := v.fromBase(ctx, offset)
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound {
			// 你增量校验，要考虑一直运行的
			// 这个就是咩有数据
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			// 查询出错了
			v.l.Error("base -> target 查询 base 失败", logx.Error(err))
			// 在这里，
			offset++
			continue
		}

		// 这边就是正常情况
		var dst T
		err = v.target.WithContext(ctx).
			Where("id = ?", src.ID()).
			First(&dst).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// target 没有
			// 丢一条消息到 Kafka 上
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
			v.l.Warn("base -> target 缺少 target 数据", logx.Int64("id", src.ID()))
		case nil:
			equal := src.CompareTo(dst)
			if !equal {
				// 要丢一条消息到 Kafka 上
				v.notify(src.ID(), events.InconsistentEventTypeNEQ)
				v.l.Warn("base -> target 数据不一致", logx.Int64("id", src.ID()))
			}
		default:
			// 记录日志，然后继续
			// 做好监控
			v.l.Error("base -> target 查询 target 失败",
				logx.Int64("id", src.ID()),
				logx.Error(err))
		}
		offset++
	}
}

func (v *Validator[T, Pdr]) Full() *Validator[T, Pdr] {
	v.fromBase = v.fullFromBase
	return v
}

func (v *Validator[T, Pdr]) Incr() *Validator[T, Pdr] {
	v.fromBase = v.incrFromBase
	return v
}

func (v *Validator[T, Pdr]) Utime(t int64) *Validator[T, Pdr] {
	v.utime = t
	return v
}

func (v *Validator[T, Pdr]) SleepInterval(interval time.Duration) *Validator[T, Pdr] {
	v.sleepInterval = interval
	return v
}

func (v *Validator[T, Pdr]) fullFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Order("id").
		Offset(offset).First(&src).Error
	return src, err
}

func (v *Validator[T, Pdr]) incrFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).
		Where("utime > ?", v.utime).
		Order("utime").
		Offset(offset).First(&src).Error
	return src, err
}

// target -> base
func (v *Validator[T, Pdr]) validateTargetToBase(ctx context.Context) error {
	offset := 0
	for {
		var ts []T
		err := v.target.WithContext(ctx).Select("id").
			//Where("utime > ?", v.utime).
			Order("id").Offset(offset).Limit(v.batchSize).
			Find(&ts).Error
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			v.l.Error("target => base 查询 target 失败", logx.Error(err))
			offset += len(ts)
			continue
		}
		// 在这里
		var srcTs []T
		ids := slicex.Map(ts, func(idx int, t T) int64 {
			return t.ID()
		})
		err = v.base.WithContext(ctx).Select("id").
			Where("id IN ?", ids).Find(&srcTs).Error
		if err == gorm.ErrRecordNotFound || len(srcTs) == 0 {
			// 都代表。base 里面一条对应的数据都没有
			v.notifyBaseMissing(ts)
			offset += len(ts)
			continue
		}
		if err != nil {
			v.l.Error("target => base 查询 base 失败", logx.Error(err))
			// 保守起见，我都认为 base 里面没有数据
			// v.notifyBaseMissing(ts)
			offset += len(ts)
			continue
		}
		// 找差集，diff 里面的，就是 target 有，但是 base 没有的
		diff := slicex.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyBaseMissing(diff)
		// 说明也没了
		if len(ts) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
		}
		offset += len(ts)
	}
}

// 批量发送 base 缺失的消息到 Kafka
func (v *Validator[T, Pdr]) notifyBaseMissing(ts []T) {
	for _, val := range ts {
		v.notify(val.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

// 上报发送不一致消息到 Kafka
func (v *Validator[T, Pdr]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	val, _ := json.Marshal(events.InconsistentEvent{
		ID:        id,
		Type:      typ,
		Direction: v.direction,
	})
	//err := v.MessageQueueConf.Producer.SendMessage(ctx, messageQueuex.Tp{Topic: v.MessageQueueConf.MessageQueueTopic}, val)
	err := v.MessageQueueConf.Producer.Send(ctx, &mqX.Message{Topic: v.MessageQueueConf.MessageQueueTopic, Value: val})
	if err != nil {
		v.l.Error("发送不一致数据消息失败", logx.Error(err), logx.String("topic", v.MessageQueueConf.MessageQueueTopic), logx.String("type", typ), logx.Int64("id", id))
	}
}
