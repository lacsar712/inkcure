package model

import "time"

func CloneSnapshot(s PlantSnapshot) PlantSnapshot {
	out := s
	out.Alarms = append([]AlarmEvent(nil), s.Alarms...)
	return out
}

func DefaultSnapshot(unitID string) PlantSnapshot {
	now := time.Now()
	return PlantSnapshot{
		UnitID: unitID,
		State:  StateColdStandby,
		Settings: PlantSettings{
			Mode:              ModeBaseLoad,
			TargetMW:          150,
			TargetSteamPSI:    NormalSteamPressurePSI,
			InkvatLevelSetpoint: 55,
			FeedwaterFlowTPH:  400,
			LampFlowTPH:       35,
			ExcessO2Setpoint:  3.5,
		},
		Plant: PlantRef{UnitLabel: unitID, PlantCode: "STEAM-PLT"},
		Inkvat: InkvatReading{
			LevelPercent: 50,
			Condition:    InkvatNormal,
			FeedwaterTPH: 0,
			SteamFlowTPH: 0,
		},
		Uvbank: UvbankReading{
			BurnerPhase: BurnerIdle,
		},
		Cureline: CurelineReading{
			SteamPressurePSI: 0,
			SteamTempF:       70,
		},
		UpdatedAt: now,
	}
}

func (s PlantSnapshot) IsFiring() bool {
	return s.State == StateFiring || s.State == StateLoadFollow || s.State == StateRamp
}

func (s PlantSnapshot) InkvatWithinLimits() bool {
	return s.Inkvat.LevelPercent >= MinInkvatLevelPercent && s.Inkvat.LevelPercent <= MaxInkvatLevelPercent
}

func (s PlantSnapshot) PressureWithinLimits() bool {
	if !s.IsFiring() {
		return true
	}
	return s.Cureline.SteamPressurePSI <= MaxSteamPressurePSI
}
