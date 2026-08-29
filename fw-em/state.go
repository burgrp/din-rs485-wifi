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

	if valid {
		device.values[index].Store(value)
		setAtomicBit(&device.valid, mask, true)
	} else {
		setAtomicBit(&device.valid, mask, false)
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
