package app

import (
	"context"
	"fmt"
)

var CalibrateProbe func(ctx context.Context) error

func (a *App) Calibrate(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		if CalibrateProbe != nil {
			if err := CalibrateProbe(ctx); err != nil {
				return fmt.Errorf("calibrate: %w", err)
			}
		}
		a.journalEvent("calibrate", fmt.Sprintf("{\"holder\":\"%s\"}", holder))
		return nil
	})
}
