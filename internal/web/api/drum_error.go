package api

import (
	"errors"

	"github.com/lacsar712/inkcure/internal/model"
)

func classifyInkvatError(err error) (string, bool) {
	if errors.Is(err, model.ErrInkvatLevelLow) {
		return "inkvat_level_low", true
	}
	return "", false
}
