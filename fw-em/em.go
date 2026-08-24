//go:build tinygo

package main

import (
	"machine"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/burgrp/bleriot-rs485/fw-em/spec"
	"github.com/burgrp/bleriot/lib/node"
	"github.com/burgrp/bleriot/lib/node/pan211x"
)

const (
	pinRadioCSN  = machine.PF3
	pinRadioSCK  = machine.PA2
	pinRadioData = machine.PA3
	pinStatus    = machine.PF0

	defaultPollMs  = 5000
	readRetries    = 3
	notifyInterval = 75 * time.Millisecond
)

var measurementTags = [...]uint16{
	spec.RegVoltage1, spec.RegVoltage2, spec.RegVoltage3,
	spec.RegCurrent1, spec.RegCurrent2, spec.RegCurrent3,
	spec.RegPowerActiveTotal, spec.RegPowerActive1, spec.RegPowerActive2, spec.RegPowerActive3,
	spec.RegPowerReactiveTotal, spec.RegPowerReactive1, spec.RegPowerReactive2, spec.RegPowerReactive3,
	spec.RegPowerFactor1, spec.RegPowerFactor2, spec.RegPowerFactor3,
	spec.RegFrequency, spec.RegEnergyActive, spec.RegEnergyReactive,
}

type firmwareConfigError string

func (err firmwareConfigError) Error() string { return string(err) }

const (
	errMeterAddress firmwareConfigError = "meter address must be between 1 and 247"
	errBaud         firmwareConfigError = "meter baud must be 1200, 2400, 4800, or 9600"
	errParity       firmwareConfigError = "invalid meter parity"
	errWordOrder    firmwareConfigError = "invalid meter word order"
)

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

type meterPoller struct {
	client *modbusClient
	device *Device
	config spec.Config
}

func (poller *meterPoller) run() {
	for {
		started := monotonicNanos()
		for _, run := range spec.PollRuns {
			poller.pollRun(run)
			runtime.Gosched()
		}
		remaining := int64(time.Duration(poller.config.PollMs)*time.Millisecond) - (monotonicNanos() - started)
		if remaining > 0 {
			time.Sleep(time.Duration(remaining))
		}
	}
}

func (poller *meterPoller) pollRun(run spec.Run) {
	var values []int32
	var err error
	for attempt := uint8(0); attempt < readRetries; attempt++ {
		values, err = poller.client.readFloat32s(poller.config.MeterAddress, run.Base, run.Count, poller.config.WordOrder)
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}

	for offset := uint8(0); offset < run.Count; offset++ {
		index := run.First + offset
		if !finiteFloat32Bits(values[offset]) {
			continue
		}
		poller.device.update(index, values[offset], true)
	}
}

func finiteFloat32Bits(raw int32) bool {
	return uint32(raw)&0x7F800000 != 0x7F800000
}

func bleriotMain(provisioning node.Provisioning, config spec.Config) {
	pinStatus.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinStatus.Low()

	config, err := normalizedConfig(config)
	if err != nil {
		println(err.Error())
		haltBlink(100 * time.Millisecond)
	}
	transport, err := newRS485Transport(config)
	if err != nil {
		println(err.Error())
		haltBlink(100 * time.Millisecond)
	}

	device := &Device{}
	n, err := pan211x.StartNode(provisioning, pinRadioSCK, pinRadioData, pinRadioCSN, device)
	if err != nil {
		println("failed to start BleRiot node:", err.Error())
		haltBlink(100 * time.Millisecond)
	}

	println("Bleriot RS485 energy meter starting")
	println("meter", config.MeterAddress, "baud", config.Baud, "poll ms", config.PollMs)
	go (&meterPoller{client: newModbusClient(transport), device: device, config: config}).run()

	var pending uint32
	var nextNotify int64
	for {
		n.Poll()
		pending |= device.dirty.Swap(0)
		if pending != 0 && monotonicNanos() >= nextNotify {
			index := firstSetBit(pending)
			mask := uint32(1) << index
			value, null := device.Read(measurementTags[index])
			n.Notify(measurementTags[index], value, null)
			pending &^= mask
			nextNotify = monotonicNanos() + int64(notifyInterval)
		}
		runtime.Gosched()
	}
}

func normalizedConfig(config spec.Config) (spec.Config, error) {
	if config.MeterAddress == 0 {
		config.MeterAddress = 1
	}
	if config.MeterAddress > 247 {
		return config, errMeterAddress
	}
	if config.Baud == 0 {
		config.Baud = 9600
	}
	switch config.Baud {
	case 1200, 2400, 4800, 9600:
	default:
		return config, errBaud
	}
	if config.Parity > spec.ParityNone {
		return config, errParity
	}
	if config.WordOrder > spec.WordOrderLowFirst {
		return config, errWordOrder
	}
	if config.PollMs == 0 {
		config.PollMs = defaultPollMs
	}
	return config, nil
}

func firstSetBit(mask uint32) uint8 {
	for index := uint8(0); index < 32; index++ {
		if mask&(uint32(1)<<index) != 0 {
			return index
		}
	}
	return 0
}

func haltBlink(period time.Duration) {
	for {
		pinStatus.High()
		time.Sleep(period)
		pinStatus.Low()
		time.Sleep(period)
	}
}
