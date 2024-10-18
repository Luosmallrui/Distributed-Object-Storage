package syncer

import (
	"context"
	"distributed-object-storage/pkg/log"
	"distributed-object-storage/redis"
	"fmt"
	"math/rand"
	"reflect"
	"runtime/debug"
	"time"
)

type Syncer interface {
	// Sync 执行函数，可取消
	Sync(ctx context.Context) error
	// Interval 执行间隔，默认10秒
	Interval() time.Duration
	// BeforeStart 第一次开始前
	BeforeStart(ctx context.Context)
	// RunOnce 只执行一次
	RunOnce() bool

	EnvIsolation() bool
}

func Run(ctx context.Context, s Syncer, envGroup string) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic recover: %v", err)
			debug.PrintStack()
		}
	}()
	s.BeforeStart(ctx)
	ticker := time.NewTicker(s.Interval())

	sName := getTypeName(s)
	if s.EnvIsolation() && envGroup != "" {
		sName = fmt.Sprintf("%s_%s", envGroup, sName)
	}

	lock := getRedisLock(sName)

	run := func() {
		err := lock.Lock(ctx)
		if err != nil {
			//logs.Debugf("get lock: %s failed", lock.Key())
			time.Sleep(delayFunc())
			return
		}

		defer func() {
			err := lock.UnLock(ctx)
			if err != nil {
				log.Warnf("unlock :%s failed", lock.Key())
			}
		}()

		start := time.Now()
		runCtx := ctx
		log.Infof("%s syncing", sName)
		err = s.Sync(runCtx)
		if err != nil {
			log.Errorf("sync failed: %v", err)
		}

		log.Infof("%s sync end with %d ms", sName, time.Since(start).Milliseconds())
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}

	// run once的任务会阻塞, 长时运行
	if s.RunOnce() {
		run()
		return
	}

	for {
		run()
		// 让其他等待的有机会获取锁
		time.Sleep(delayFunc())
	}
}

func delayFunc() time.Duration {
	return time.Duration(rand.Intn(300)+200) * time.Millisecond
}

func getRedisLock(name string) *redis.Lock {
	key := fmt.Sprintf("nice-offer%s", name)
	return redis.NewRedisLock(key)
}

func getTypeName(s Syncer) string {
	r := reflect.TypeOf(s)
	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}
	return r.Name()
}
