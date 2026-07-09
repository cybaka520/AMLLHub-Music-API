package api

import (
	"context"
	"time"
)

// sleepCtx 在支持 context 取消的前提下休眠；ctx 被取消时立即返回错误。
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
