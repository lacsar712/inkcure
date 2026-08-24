package uvbank

import (
	"math"

	"github.com/lacsar712/inkcure/internal/clock"
	"github.com/lacsar712/inkcure/internal/model"
)

type BurnerController struct {
	clk clock.ProcessClock
}

func NewBurnerController(clk clock.ProcessClock) *BurnerController {
	return &BurnerController{clk: clk}
}

func (b *BurnerController) EstimateLampframeTemp(reading model.UvbankReading) float64 {
	base := 300.0
	lampHeat := reading.LampFlowTPH * 50
	airCool := reading.AirflowTPH * 2
	return base + lampHeat - airCool
}

func (b *BurnerController) CureStable(reading model.UvbankReading) bool {
	if reading.BurnerPhase != model.BurnerStable && reading.BurnerPhase != model.BurnerIgnition {
		return false
	}
	return reading.LampframeTempF > 800 && reading.ExcessO2Pct >= model.MinLampframeO2Percent
}

func (b *BurnerController) TripRequired(reading model.UvbankReading) bool {
	if reading.ExcessO2Pct > model.MaxLampframeO2Percent*2 {
		return true
	}
	if reading.BurnerPhase == model.BurnerTrip {
		return true
	}
	if reading.LampframeTempF > 3500 {
		return true
	}
	return false
}

func (b *BurnerController) PhaseLabel(phase model.BurnerPhase) string {
	switch phase {
	case model.BurnerIdle:
		return "Idle"
	case model.BurnerDwell:
		return "Dwell"
	case model.BurnerIgnition:
		return "Ignition"
	case model.BurnerStable:
		return "Stable Cure"
	case model.BurnerTrip:
		return "Tripped"
	default:
		return string(phase)
	}
}

func (b *BurnerController) HeatReleaseMW(reading model.UvbankReading) float64 {
	return reading.LampFlowTPH * 12.5
}

func (b *BurnerController) TurndownRatio(settings model.PlantSettings, currentLamp float64) float64 {
	if settings.LampFlowTPH <= 0 {
		return 0
	}
	return currentLamp / settings.LampFlowTPH
}

func (b *BurnerController) MinStableLamp(settings model.PlantSettings) float64 {
	return settings.LampFlowTPH * 0.25
}

func (b *BurnerController) NormalizeLamp(flow, max float64) float64 {
	return math.Min(math.Max(flow, 0), max)
}
