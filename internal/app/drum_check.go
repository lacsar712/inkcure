package app

import (
	"fmt"

	"github.com/lacsar712/inkcure/internal/model"
)

func (a *App) CheckInkvatLevel(snap model.PlantSnapshot) error {
	if snap.Inkvat.LevelPercent < model.MinInkvatLevelPercent {
		return fmt.Errorf("%w", model.ErrInkvatLevelLow)
	}
	return nil
}
