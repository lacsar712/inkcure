package fsm

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/inkcure/internal/model"
)

type CurelineFSM struct {
	mu            sync.RWMutex
	state         model.PlantState
	lampPermissive bool
	dwellComplete  bool
	hooks          *HookChain
}

func NewCurelineFSM(unitID string) *CurelineFSM {
	_ = unitID
	return &CurelineFSM{state: model.StateColdStandby, hooks: NewHookChain()}
}

func (f *CurelineFSM) Hooks() *HookChain { return f.hooks }

func (f *CurelineFSM) State() model.PlantState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

func (f *CurelineFSM) SetLampPermissive(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lampPermissive = ok
}

func (f *CurelineFSM) SetDwellComplete(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dwellComplete = ok
}

func (f *CurelineFSM) LampPermissive() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.lampPermissive
}

func (f *CurelineFSM) Dispatch(ctx context.Context, event PlantEvent) (model.PlantState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case <-ctx.Done():
		return f.state, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if event == EvTrip {
		from := f.state
		if f.hooks != nil {
			if err := f.hooks.RunBefore(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		f.state = model.StateTrip
		if f.hooks != nil {
			if err := f.hooks.RunAfter(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		return f.state, nil
	}
	next, ok := NextState(f.state, event)
	if !ok {
		if f.hooks != nil {
			_ = f.hooks.RunAfter(ctx, f.state, f.state, event)
		}
		return f.state, fmt.Errorf("%s from %s: %w", event, f.state, ErrIllegalTransition)
	}
	if event == EvIgnite && !f.lampPermissive {
		return f.state, fmt.Errorf("%w", model.ErrLampPermissive)
	}
	if event == EvDwellComplete && !f.dwellComplete {
		return f.state, fmt.Errorf("%w", model.ErrDwellIncomplete)
	}
	from := f.state
	if f.hooks != nil {
		if err := f.hooks.RunBefore(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	f.state = next
	if f.hooks != nil {
		if err := f.hooks.RunAfter(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	return f.state, nil
}

func (f *CurelineFSM) ForceState(state model.PlantState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}
