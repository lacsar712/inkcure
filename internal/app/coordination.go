package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/inkcure/internal/model"
)

func (a *App) CoordinateLoadChange(ctx context.Context, holder string, targetMW float64) error {
	if err := a.coordLock.Acquire(holder); err != nil {
		return err
	}
	defer a.coordLock.Release(holder)

	snap := a.Snapshot()
	if snap.Settings.TargetMW <= 0 {
		return fmt.Errorf("invalid target MW")
	}
	loadPct := targetMW / snap.Settings.TargetMW
	if loadPct > 1.1 {
		return fmt.Errorf("load request exceeds rated capacity")
	}
	if err := a.inkvat.RequireSettled(snap.Inkvat); err != nil {
		return err
	}
	settings := snap.Settings
	settings.TargetMW = targetMW
	if err := a.store.UpdateSettings(a.cfg.UnitID, settings); err != nil {
		return err
	}
	return a.RampLoad(ctx, holder, loadPct)
}

func (a *App) CoordinateInkvatLevel(ctx context.Context, holder string, setpoint float64) error {
	if err := a.coordLock.Acquire(holder); err != nil {
		return err
	}
	defer a.coordLock.Release(holder)

	if setpoint < model.MinInkvatLevelPercent || setpoint > model.MaxInkvatLevelPercent {
		return fmt.Errorf("inkvat setpoint out of range")
	}
	snap := a.Snapshot()
	settings := snap.Settings
	settings.InkvatLevelSetpoint = setpoint
	feed := a.inkvat.CoordinateFeedwater(snap, a.isFiring(snap.State))
	settings.FeedwaterFlowTPH = feed
	return a.store.UpdateSettings(a.cfg.UnitID, settings)
}

func (a *App) CoordinateUvbankTrim(ctx context.Context, holder string, o2Setpoint float64) error {
	if err := a.coordLock.Acquire(holder); err != nil {
		return err
	}
	defer a.coordLock.Release(holder)

	if o2Setpoint < model.MinLampframeO2Percent || o2Setpoint > model.MaxLampframeO2Percent {
		return fmt.Errorf("O2 setpoint out of range")
	}
	snap := a.Snapshot()
	settings := snap.Settings
	settings.ExcessO2Setpoint = o2Setpoint
	if err := a.store.UpdateSettings(a.cfg.UnitID, settings); err != nil {
		return err
	}
	comb := snap.Uvbank
	comb.AirflowTPH = a.uvbank.Airflow().TrimAirflow(comb.AirflowTPH, o2Setpoint, comb.ExcessO2Pct)
	comb.ExcessO2Pct = a.uvbank.Airflow().ExcessO2(comb)
	return a.store.UpdateUvbank(a.cfg.UnitID, comb)
}

func (a *App) CoordinationHeld() bool { return a.coordLock.Held() }

func (a *App) PlantHealth() map[string]string {
	snap := a.Snapshot()
	a.refreshPermissives(snap)
	out := map[string]string{
		"state":      string(snap.State),
		"inkvat_ok":    fmt.Sprintf("%v", a.permissives.InkvatOK()),
		"pressure_ok": fmt.Sprintf("%v", a.permissives.PressureOK()),
		"uvbank_ok": fmt.Sprintf("%v", a.permissives.UvbankOK()),
		"lamp_ok":    fmt.Sprintf("%v", a.permissives.LampOK()),
	}
	ready, detail := a.WarmupStatus()
	out["warmup_ready"] = fmt.Sprintf("%v", ready)
	out["warmup_detail"] = detail
	return out
}
