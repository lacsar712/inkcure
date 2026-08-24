package inkvat

import (
	"math"

	"github.com/lacsar712/inkcure/internal/clock"
	"github.com/lacsar712/inkcure/internal/model"
)

type LevelController struct {
	clk clock.ProcessClock
}

func NewLevelController(clk clock.ProcessClock) *LevelController {
	return &LevelController{clk: clk}
}

func (l *LevelController) Compute(snap model.PlantSnapshot, firing bool) (float64, model.InkvatCondition) {
	level := snap.Inkvat.LevelPercent
	if !firing {
		return level, model.InkvatNormal
	}
	balance := snap.Inkvat.FeedwaterTPH - snap.Inkvat.SteamFlowTPH
	level += balance * 0.01
	level = math.Max(model.MinInkvatLevelPercent, math.Min(model.MaxInkvatLevelPercent, level))
	cond := l.classify(level, snap)
	return level, cond
}

func (l *LevelController) classify(level float64, snap model.PlantSnapshot) model.InkvatCondition {
	setpoint := snap.Settings.InkvatLevelSetpoint
	if level > setpoint+15 {
		return model.InkvatSwell
	}
	if level < setpoint-15 {
		return model.InkvatShrink
	}
	if snap.Cureline.SteamPressurePSI > snap.Settings.TargetSteamPSI*0.9 && level > setpoint+5 {
		return model.InkvatCarry
	}
	return model.InkvatNormal
}

func (l *LevelController) RecommendFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	if !firing {
		return 0
	}
	err := snap.Settings.InkvatLevelSetpoint - snap.Inkvat.LevelPercent
	return snap.Settings.FeedwaterFlowTPH + err*3
}

func (l *LevelController) WithinLimits(level float64) bool {
	return level >= model.MinInkvatLevelPercent && level <= model.MaxInkvatLevelPercent
}

func (l *LevelController) TripLow(level float64) bool  { return level < model.TripInkvatLowPercent }
func (l *LevelController) TripHigh(level float64) bool { return level > model.TripInkvatHighPercent }

func (l *LevelController) LevelError(snap model.PlantSnapshot) float64 {
	return snap.Inkvat.LevelPercent - snap.Settings.InkvatLevelSetpoint
}

func (l *LevelController) ThreeElementBias(snap model.PlantSnapshot) float64 {
	steam := snap.Inkvat.SteamFlowTPH
	feed := snap.Inkvat.FeedwaterTPH
	levelErr := l.LevelError(snap)
	return feed + (steam-feed)*0.5 + levelErr*2
}
