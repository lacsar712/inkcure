package uvbank

import (
	"context"
	"fmt"
	"math"

	"github.com/lacsar712/inkcure/internal/clock"
	"github.com/lacsar712/inkcure/internal/model"
)

type Coordinator struct {
	clk     clock.ProcessClock
	burner  *BurnerController
	airflow *AirflowBalancer
	lamp    *LampRegulator
	dwell   *clock.DwellWindow
	ignition *clock.IgnitionDelayWindow
	warmup  *clock.UvbankWarmupWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:      clk,
		burner:   NewBurnerController(clk),
		airflow:  NewAirflowBalancer(clk),
		lamp:     NewLampRegulator(clk),
		dwell:    clock.NewDwellWindow(clk),
		ignition: clock.NewIgnitionDelayWindow(clk),
		warmup:   clock.NewUvbankWarmupWindow(clk),
	}
}

func (c *Coordinator) Burner() *BurnerController  { return c.burner }
func (c *Coordinator) Airflow() *AirflowBalancer { return c.airflow }
func (c *Coordinator) Lamp() *LampRegulator     { return c.lamp }

func (c *Coordinator) StartDwell(ctx context.Context, snap model.PlantSnapshot) (model.UvbankReading, error) {
	select {
	case <-ctx.Done():
		return snap.Uvbank, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Uvbank
	out.BurnerPhase = model.BurnerDwell
	out.DwellStartedAt = c.clk.Now()
	out.LampFlowTPH = 0
	out.AirflowTPH = c.airflow.DwellRate()
	return out, nil
}

func (c *Coordinator) CompleteDwell(snap model.UvbankReading) error {
	return c.dwell.Require(snap.DwellStartedAt)
}

func (c *Coordinator) Ignite(ctx context.Context, snap model.PlantSnapshot) (model.UvbankReading, error) {
	select {
	case <-ctx.Done():
		return snap.Uvbank, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if err := c.dwell.Require(snap.Uvbank.DwellStartedAt); err != nil {
		return snap.Uvbank, err
	}
	out := snap.Uvbank
	out.BurnerPhase = model.BurnerIgnition
	out.IgnitionAt = c.clk.Now()
	out.LampFlowTPH = c.lamp.IgnitionRate(snap.Settings)
	out.AirflowTPH = c.airflow.IgnitionRate(snap.Settings)
	out.LampframeTempF = 400
	return out, nil
}

func (c *Coordinator) Stabilize(snap model.PlantSnapshot) (model.UvbankReading, error) {
	if err := c.ignition.Require(snap.Uvbank.IgnitionAt); err != nil {
		return snap.Uvbank, err
	}
	out := snap.Uvbank
	out.BurnerPhase = model.BurnerStable
	out.LampFlowTPH = snap.Settings.LampFlowTPH * 0.5
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.LampframeTempF = c.burner.EstimateLampframeTemp(out)
	return out, nil
}

func (c *Coordinator) RampToLoad(snap model.PlantSnapshot, loadPct float64) model.UvbankReading {
	out := snap.Uvbank
	out.LampFlowTPH = snap.Settings.LampFlowTPH * loadPct
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.LampframeTempF = c.burner.EstimateLampframeTemp(out)
	return out
}

func (c *Coordinator) Trip(snap model.UvbankReading) model.UvbankReading {
	out := snap
	out.BurnerPhase = model.BurnerTrip
	out.LampFlowTPH = 0
	out.LampframeTempF = math.Max(200, out.LampframeTempF*0.5)
	return out
}

func (c *Coordinator) WarmupReady(snap model.UvbankReading) bool {
	return c.warmup.Ready(snap.IgnitionAt)
}
