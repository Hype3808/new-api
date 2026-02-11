package common

import (
	"context"
	"fmt"

	"github.com/bytedance/gopkg/util/gopool"
)

var relayGoPool gopool.Pool

func init() {
	poolSize := GetEnvOrDefault("RELAY_GOPOOL_SIZE", 100)
	if poolSize < 1 {
		poolSize = 100
	}
	relayGoPool = gopool.NewPool("gopool.RelayPool", int32(poolSize), gopool.NewConfig())
	relayGoPool.SetPanicHandler(func(ctx context.Context, i interface{}) {
		if stopChan, ok := ctx.Value("stop_chan").(chan bool); ok {
			SafeSendBool(stopChan, true)
		}
		SysError(fmt.Sprintf("panic in gopool.RelayPool: %v", i))
	})
}

func RelayCtxGo(ctx context.Context, f func()) {
	relayGoPool.CtxGo(ctx, f)
}
