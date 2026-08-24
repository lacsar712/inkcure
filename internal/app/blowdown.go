package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/inkcure/internal/model"
)

const maxSolventOpeningPct = 100.0

func (a *App) OpenSolvent(ctx context.Context, holder string, openingPct float64) error {
	_ = holder
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if openingPct >= maxSolventOpeningPct {
		return fmt.Errorf("solvent: %w", model.ErrSolventLimit)
	}
	return nil
}

func (a *App) SolventAfterShutdown(ctx context.Context, openingPct float64) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	snap := a.Snapshot()
	if snap.State != model.StateTrip && snap.State != model.StateColdStandby {
		return fmt.Errorf("plant not shut down")
	}
	if openingPct >= maxSolventOpeningPct {
		return fmt.Errorf("unknown fault")
	}
	return nil
}
