package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/inkcure/internal/model"
)

func (a *App) WarmupStatus() (ready bool, detail string) {
	snap := a.Snapshot()
	if snap.Uvbank.DwellStartedAt.IsZero() {
		return false, "dwell not started"
	}
	if !a.dwellWindow.Ready(snap.Uvbank.DwellStartedAt) {
		return false, "dwell window open"
	}
	if !snap.Uvbank.IgnitionAt.IsZero() && !a.warmupWindow.Ready(snap.Uvbank.IgnitionAt) {
		return false, "uvbank warmup window open"
	}
	if !snap.Inkvat.LastSwellAt.IsZero() {
		if err := a.inkvat.RequireSettled(snap.Inkvat); err != nil {
			return false, "inkvat swell settling"
		}
	}
	return true, "ready"
}

func (a *App) WaitWarmup(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w", model.ErrContextDone)
		default:
		}
		ready, _ := a.WarmupStatus()
		if ready {
			return nil
		}
	}
}

func (a *App) DwellRemaining() string {
	snap := a.Snapshot()
	if snap.Uvbank.DwellStartedAt.IsZero() {
		return "not started"
	}
	if a.dwellWindow.Ready(snap.Uvbank.DwellStartedAt) {
		return "complete"
	}
	return "in progress"
}

func (a *App) UvbankWarmupRemaining() string {
	snap := a.Snapshot()
	if snap.Uvbank.IgnitionAt.IsZero() {
		return "not ignited"
	}
	if a.warmupWindow.Ready(snap.Uvbank.IgnitionAt) {
		return "complete"
	}
	return "in progress"
}
