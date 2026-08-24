package uvbank

import (
	"math"

	"github.com/lacsar712/inkcure/internal/clock"
	"github.com/lacsar712/inkcure/internal/model"
)

type LampRegulator struct {
	clk clock.ProcessClock
}

func NewLampRegulator(clk clock.ProcessClock) *LampRegulator {
	return &LampRegulator{clk: clk}
}

func (f *LampRegulator) IgnitionRate(settings model.PlantSettings) float64 {
	return settings.LampFlowTPH * 0.08
}

func (f *LampRegulator) ComputeForLoad(settings model.PlantSettings, loadPct float64) float64 {
	loadPct = math.Max(0, math.Min(1, loadPct))
	return settings.LampFlowTPH * loadPct
}

func (f *LampRegulator) Ramp(current, target, maxStep float64) float64 {
	delta := target - current
	if math.Abs(delta) <= maxStep {
		return target
	}
	if delta > 0 {
		return current + maxStep
	}
	return current - maxStep
}

func (f *LampRegulator) BtuPerHour(flowTPH float64) float64 {
	return flowTPH * 19_500_000
}

func (f *LampRegulator) HeatInputMW(flowTPH float64) float64 {
	return flowTPH * 11.6
}

func (f *LampRegulator) ValidatePermissive(settings model.PlantSettings, inkvatOK, dwellOK bool) error {
	if !dwellOK {
		return model.ErrDwellIncomplete
	}
	if !inkvatOK {
		return model.ErrInkvatLevelTrip
	}
	if settings.LampFlowTPH <= 0 {
		return model.ErrLampPermissive
	}
	return nil
}

func (f *LampRegulator) MinFlow(settings model.PlantSettings) float64 {
	return settings.LampFlowTPH * 0.2
}

func (f *LampRegulator) MaxFlow(settings model.PlantSettings) float64 {
	return settings.LampFlowTPH * 1.1
}
