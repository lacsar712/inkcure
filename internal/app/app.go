package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/inkcure/internal/cureline"
	"github.com/lacsar712/inkcure/internal/clock"
	"github.com/lacsar712/inkcure/internal/uvbank"
	"github.com/lacsar712/inkcure/internal/config"
	"github.com/lacsar712/inkcure/internal/inkvat"
	"github.com/lacsar712/inkcure/internal/fsm"
	"github.com/lacsar712/inkcure/internal/interlock"
	"github.com/lacsar712/inkcure/internal/model"
	"github.com/lacsar712/inkcure/internal/store"
)

type App struct {
	cfg           config.Config
	clk           clock.ProcessClock
	store         *store.PlantStore
	journal       *store.Journal
	fsm           *fsm.CurelineFSM
	cureline        *cureline.Controller
	uvbank    *uvbank.Coordinator
	inkvat          *inkvat.Coordinator
	interlock     *interlock.Interlock
	permissives   *interlock.PermissiveSet
	coordLock     *interlock.CoordinationLock
	scheduler     *clock.Scheduler
	dwellWindow   *clock.DwellWindow
	warmupWindow  *clock.UvbankWarmupWindow
	telemetry     *Telemetry
	tickCancels    map[string]context.CancelFunc
	lampLoopCancels map[string]context.CancelFunc
	mu             sync.RWMutex
}

func New(cfg config.Config, clk clock.ProcessClock) *App {
	return &App{
		cfg:          cfg,
		clk:          clk,
		store:        store.NewPlantStore(),
		journal:      store.NewJournal(cfg.JournalPath, cfg.JournalCapacity),
		fsm:          fsm.NewCurelineFSM(cfg.UnitID),
		cureline:       cureline.NewController(clk),
		uvbank:   uvbank.NewCoordinator(clk),
		inkvat:         inkvat.NewCoordinator(clk),
		interlock:    interlock.NewInterlock(cfg.LeaseTTL),
		permissives:  interlock.NewPermissiveSet(),
		coordLock:    interlock.NewCoordinationLock(),
		scheduler:    clock.NewScheduler(clk),
		dwellWindow:  clock.NewDwellWindow(clk),
		warmupWindow: clock.NewUvbankWarmupWindow(clk),
		telemetry:    NewTelemetry(cfg.UnitID),
		tickCancels:     make(map[string]context.CancelFunc),
		lampLoopCancels: make(map[string]context.CancelFunc),
	}
}

func (a *App) Snapshot() model.PlantSnapshot {
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return model.DefaultSnapshot(a.cfg.UnitID)
	}
	return snap
}

func (a *App) Config() config.Config              { return a.cfg }
func (a *App) Clock() clock.ProcessClock          { return a.clk }
func (a *App) FSM() *fsm.CurelineFSM                { return a.fsm }
func (a *App) UnitID() string                     { return a.cfg.UnitID }
func (a *App) Store() *store.PlantStore           { return a.store }
func (a *App) Interlock() *interlock.Interlock    { return a.interlock }
func (a *App) Telemetry() TelemetrySnapshot       { return a.telemetry.Snapshot() }
func (a *App) Journal() *store.Journal            { return a.journal }

func (a *App) journalEvent(ev, payload string) {
	_, _ = a.journal.Append(a.cfg.UnitID, ev, payload)
}

func (a *App) syncState(state model.PlantState) {
	_ = a.store.UpdateState(a.cfg.UnitID, state)
}

func (a *App) isFiring(state model.PlantState) bool {
	return state == model.StateFiring || state == model.StateLoadFollow || state == model.StateRamp
}

func (a *App) refreshPermissives(snap model.PlantSnapshot) {
	a.permissives.SetInkvat(a.inkvat.Level().WithinLimits(snap.Inkvat.LevelPercent))
	a.permissives.SetPressure(a.cureline.Pressure().WithinTripLimits(snap.Cureline.SteamPressurePSI, a.isFiring(snap.State)))
	a.permissives.SetUvbank(a.uvbank.Burner().CureStable(snap.Uvbank))
	a.permissives.SetLamp(snap.Uvbank.LampFlowTPH > 0 || snap.State == model.StateDwell)
	a.permissives.SetIgnition(snap.Uvbank.BurnerPhase == model.BurnerStable || snap.Uvbank.BurnerPhase == model.BurnerIgnition)
	a.fsm.SetLampPermissive(a.permissives.LampOK())
	a.fsm.SetDwellComplete(a.dwellWindow.Ready(snap.Uvbank.DwellStartedAt))
}

func (a *App) tickLabel() string {
	return fmt.Sprintf("%s-tick", a.cfg.UnitID)
}
