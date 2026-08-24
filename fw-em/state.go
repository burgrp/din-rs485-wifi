package main

import (
	"sync/atomic"

	"github.com/burgrp/bleriot-rs485/fw-em/spec"
)

var measurementTags = [...]uint16{
	spec.RegVoltage1, spec.RegVoltage2, spec.RegVoltage3,
	spec.RegCurrent1, spec.RegCurrent2, spec.RegCurrent3,
	spec.RegPowerActiveTotal, spec.RegPowerActive1, spec.RegPowerActive2, spec.RegPowerActive3,
	spec.RegPowerReactiveTotal, spec.RegPowerReactive1, spec.RegPowerReactive2, spec.RegPowerReactive3,
	spec.RegPowerFactor1, spec.RegPowerFactor2, spec.RegPowerFactor3,
	spec.RegFrequency, spec.RegEnergyActive, spec.RegEnergyReactive,
}

type Device struct {
	values [len(measurementTags)]atomic.Int32
	valid  atomic.Uint32
	dirty  atomic.Uint32
}

func (device *Device) Read(tag uint16) (int32, bool) {
	for index, candidate := range measurementTags {
		if candidate == tag {
			if device.valid.Load()&(uint32(1)<<index) == 0 {
				return 0, true
			}
			return device.values[index].Load(), false
		}
	}
	return 0, true
}

func (device *Device) Write(tag uint16, value int32, null bool) {}

func (device *Device) update(index uint8, value int32, valid bool) {
	mask := uint32(1) << index
	wasValid := device.valid.Load()&mask != 0
	oldValue := device.values[index].Load()

	if valid {
		device.values[index].Store(value)
		setAtomicBit(&device.valid, mask, true)
	} else {
		setAtomicBit(&device.valid, mask, false)
	}
	if wasValid != valid || valid && oldValue != value {
		setAtomicBit(&device.dirty, mask, true)
	}
}

func setAtomicBit(value *atomic.Uint32, mask uint32, set bool) {
	for {
		old := value.Load()
		updated := old | mask
		if !set {
			updated = old &^ mask
		}
		if old == updated || value.CompareAndSwap(old, updated) {
			return
		}
	}
}

type measurementFreshness struct {
	lastSuccess int64
	valid       bool
}

func (freshness *measurementFreshness) record(now int64) {
	freshness.lastSuccess = now
	freshness.valid = true
}

func (freshness *measurementFreshness) expire(now int64, timeout int64) bool {
	if freshness.valid && now-freshness.lastSuccess >= timeout {
		freshness.valid = false
		return true
	}
	return false
}

type notificationState struct {
	values [len(measurementTags)]int32
	known  uint32
	valid  uint32
}

func (state *notificationState) changed(index uint8, value int32, null bool) bool {
	mask := uint32(1) << index
	valid := !null
	wasKnown := state.known&mask != 0
	wasValid := state.valid&mask != 0
	if wasKnown && wasValid == valid && (!valid || state.values[index] == value) {
		return false
	}

	state.known |= mask
	if valid {
		state.values[index] = value
		state.valid |= mask
	} else {
		state.valid &^= mask
	}
	return true
}
