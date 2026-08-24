package interlock

import (
	"fmt"

	"github.com/lacsar712/inkcure/internal/model"
)

type PermissiveSet struct {
	lampOK       bool
	ignitionOK   bool
	inkvatOK       bool
	pressureOK   bool
	uvbankOK bool
}

func NewPermissiveSet() *PermissiveSet { return &PermissiveSet{} }

func (p *PermissiveSet) SetLamp(ok bool)       { p.lampOK = ok }
func (p *PermissiveSet) SetIgnition(ok bool)   { p.ignitionOK = ok }
func (p *PermissiveSet) SetInkvat(ok bool)       { p.inkvatOK = ok }
func (p *PermissiveSet) SetPressure(ok bool)   { p.pressureOK = ok }
func (p *PermissiveSet) SetUvbank(ok bool) { p.uvbankOK = ok }

func (p *PermissiveSet) LampOK() bool       { return p.lampOK }
func (p *PermissiveSet) IgnitionOK() bool   { return p.ignitionOK }
func (p *PermissiveSet) InkvatOK() bool       { return p.inkvatOK }
func (p *PermissiveSet) PressureOK() bool   { return p.pressureOK }
func (p *PermissiveSet) UvbankOK() bool { return p.uvbankOK }

func (p *PermissiveSet) AllFiring() bool {
	return p.lampOK && p.ignitionOK && p.inkvatOK && p.pressureOK && p.uvbankOK
}

func (p *PermissiveSet) CheckIgnition() error {
	if !p.lampOK {
		return fmt.Errorf("%w", model.ErrLampPermissive)
	}
	if !p.ignitionOK {
		return fmt.Errorf("%w", model.ErrIgnitionBlocked)
	}
	return nil
}

func CheckCureLoss(reading model.UvbankReading) error {
	if reading.BurnerPhase == model.BurnerStable && reading.LampframeTempF < 600 {
		return fmt.Errorf("%w", model.ErrCureLoss)
	}
	return nil
}

func (p *PermissiveSet) CheckFiring() error {
	if err := p.CheckIgnition(); err != nil {
		return err
	}
	if !p.inkvatOK {
		return fmt.Errorf("%w", model.ErrInkvatLevelTrip)
	}
	if !p.pressureOK {
		return fmt.Errorf("%w", model.ErrPressureTrip)
	}
	if !p.uvbankOK {
		return fmt.Errorf("%w", model.ErrUvbankTrip)
	}
	return nil
}

type CoordinationLock struct {
	holder string
	held   bool
}

func NewCoordinationLock() *CoordinationLock { return &CoordinationLock{} }

func (c *CoordinationLock) Acquire(holder string) error {
	if c.held {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	c.holder = holder
	c.held = true
	return nil
}

func (c *CoordinationLock) Release(holder string) {
	if c.held && c.holder == holder {
		c.held = false
		c.holder = ""
	}
}

func (c *CoordinationLock) Require(holder string) error {
	if !c.held || c.holder != holder {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	return nil
}

func (c *CoordinationLock) Held() bool { return c.held }
