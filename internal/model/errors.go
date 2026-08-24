package model

import "errors"

var (
	ErrContextDone      = errors.New("operation cancelled")
	ErrPlantNotFound    = errors.New("plant unit not found")
	ErrLeaseHeld        = errors.New("interlock lease held by another operator")
	ErrLeaseMissing     = errors.New("interlock lease missing or expired")
	ErrGateBlocked      = errors.New("safety gate blocked")
	ErrLampPermissive   = errors.New("lamp permissive not satisfied")
	ErrIgnitionBlocked  = errors.New("ignition sequence blocked")
	ErrInkvatLevelTrip    = errors.New("inkvat level trip condition")
	ErrPressureTrip     = errors.New("steam pressure trip condition")
	ErrUvbankTrip   = errors.New("uvbank trip condition")
	ErrIllegalState     = errors.New("illegal plant state transition")
	ErrSnapshotStale    = errors.New("snapshot revision stale")
	ErrWindowOpen       = errors.New("timing window still open")
	ErrDwellIncomplete  = errors.New("lampframe dwell incomplete")
	ErrCoordinationLock = errors.New("coordination lock held")
	ErrInkvatLevelLow     = errors.New("inkvat level below low limit")
	ErrCureLoss        = errors.New("lampframe cure lost")
	ErrSolventLimit    = errors.New("solvent valve at limit")
)
