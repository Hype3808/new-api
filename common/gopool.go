package common

import (
	"context"
	"fmt"

	"github.com/bytedance/gopkg/util/gopool"
)

var relayGoPool gopool.Pool

func init() {
	// Limit concurrent goroutines to prevent CPU thrashing on small servers
	// For 1 vCPU: 50-100 concurrent goroutines is reasonable
	// For 2+ vCPUs: can increase to 200-500
	// Configurable via GOPOOL_WORKER_SIZE env var
	poolSize := GetEnvOrDefault("GOPOOL_WORKER_SIZE", 100)
	
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
