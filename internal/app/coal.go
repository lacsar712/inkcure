package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/inkcure/internal/clock"
	"github.com/lacsar712/inkcure/internal/model"
)

func (a *App) advanceClock(d time.Duration) {
	if mc, ok := a.clk.(*clock.ManualClock); ok {
		mc.Advance(d)
		time.Sleep(time.Millisecond)
	} else {
		time.Sleep(d)
	}
}

var activeLampCancel context.CancelFunc

func (a *App) bindLampLoop(holder string, ctx context.Context) context.Context {
	a.mu.Lock()
	if activeLampCancel != nil {
		activeLampCancel()
	}
	child, cancel := context.WithCancel(ctx)
	activeLampCancel = cancel
	a.mu.Unlock()
	return child
}

func (a *App) cancelLampLoop(holder string) {
	a.mu.Lock()
	if activeLampCancel != nil {
		activeLampCancel()
		activeLampCancel = nil
	}
	a.mu.Unlock()
}

func (a *App) cancelAllLampLoops() {
	a.mu.Lock()
	for holder, cancel := range a.lampLoopCancels {
		cancel()
		delete(a.lampLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) CoalFeedTPH() float64 {
	return a.Snapshot().Uvbank.LampFlowTPH
}

func (a *App) RunLampRamp(ctx context.Context, holder string, targetTPH float64) error {
	loopCtx := a.bindLampLoop(holder, ctx)
	defer a.cancelLampLoop(holder)
	for {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		current := snap.Uvbank.LampFlowTPH
		if current >= targetTPH {
			return nil
		}
		comb := snap.Uvbank
		comb.LampFlowTPH = current + 1.0
		_ = a.store.UpdateUvbank(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.LampFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
}

func (a *App) RunCoalFeed(ctx context.Context, holder string, steps int) error {
	loopCtx := a.bindLampLoop(holder, ctx)
	defer a.cancelLampLoop(holder)
	for i := 0; steps <= 0 || i < steps; i++ {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		comb := snap.Uvbank
		comb.LampFlowTPH += 0.5
		_ = a.store.UpdateUvbank(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.LampFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
	return nil
}
