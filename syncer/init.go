package syncer

import (
	"context"
	"distributed-object-storage/pkg/log"
	"reflect"
	"sync"
)

func Init(ctx context.Context) {
	var syncers = []Syncer{
		NewNodeHealthCheckSyncer(),
	}
	wg := new(sync.WaitGroup)
	for _, iter := range syncers {
		wg.Add(1)
		log.Infof("start syncer:[%s]", reflect.TypeOf(iter))
		oneSyncer := iter
		go func() {
			defer wg.Done()
			for { // rerun after recover
				select {
				case <-ctx.Done():
					return
				default:
					Run(ctx, oneSyncer, "test")
				}
			}
		}()
	}
	log.Info("all syncers started")
	wg.Wait()
}
