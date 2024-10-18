package redis

import (
	"context"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/prometheus/common/log"
	"time"
)

var (
	DefaultExpire = 12 * time.Second
)

type Lock struct {
	key    string
	m      *redsync.Mutex
	expire time.Duration
	cancel context.CancelFunc
}

func NewRedisLock(key string) *Lock {
	return NewRedisLockWithExpire(key, DefaultExpire)
}

func NewRedisLockWithExpire(key string, expire time.Duration) *Lock {
	pool := goredis.NewPool(Redis())
	rs := redsync.New(pool)
	m := rs.NewMutex(key, redsync.WithTries(1), redsync.WithExpiry(expire))
	return &Lock{
		key:    key,
		m:      m,
		expire: DefaultExpire,
		cancel: nil,
	}
}

func (l *Lock) Expire() time.Duration {
	return l.expire
}

func (l *Lock) Key() string {
	return l.key
}

func (l *Lock) Lock(ctx context.Context) error {
	err := l.m.LockContext(ctx)
	if err != nil {
		return err
	}

	log.Debugf("get lock %s success", l.key)

	if l.cancel == nil {
		c, cancel := context.WithCancel(context.TODO())
		l.cancel = cancel
		l.RenewLock(c)
	}

	return nil
}

func (l *Lock) UnLock(ctx context.Context) error {
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
	_, err := l.m.UnlockContext(ctx)
	if err != nil {
		return err
	}
	log.Debugf("unlock %s success", l.key)
	return nil
}

func (l *Lock) RenewLock(ctx context.Context) {
	go func() {
		for {
			time.Sleep(l.expire * 2 / 3)
			select {
			case <-ctx.Done():
				return
			default:
				_, err := l.m.ExtendContext(ctx)
				if err != nil {
					log.Warnf("extend err: %v", err)
				} else {
					log.Debugf("extend lock: %s success", l.key)
				}
			}
		}
	}()
}
