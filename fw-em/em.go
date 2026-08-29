//go:build tinygo

package main

import (
	"machine"
	"runtime"
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

	defaultPollMs       = 5000
	defaultDisconnectMs = 15000
	readRetries         = 3
)

type firmwareConfigError string

func (err firmwareConfigError) Error() string { return string(err) }

const (
	errMeterAddress firmwareConfigError = "meter address must be between 1 and 247"
	errBaud         firmwareConfigError = "meter baud must be 1200, 2400, 4800, or 9600"
	errParity       firmwareConfigError = "invalid meter parity"
	errWordOrder    firmwareConfigError = "invalid meter word order"
)

type meterPoller struct {
	client     *modbusClient
	device     *Device
	config     spec.Config
	freshness  [len(measurementTags)]measurementFreshness
	disconnect int64
}

func (poller *meterPoller) run() {
	for {
		started := monotonicNanos()
		for _, run := range spec.PollRuns {
			poller.pollRun(run)
			runtime.Gosched()
		}
		poller.expireMeasurements(monotonicNanos())
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

	now := monotonicNanos()
	for offset := uint8(0); offset < run.Count; offset++ {
		index := run.First + offset
		if !finiteFloat32Bits(values[offset]) {
			continue
		}
		poller.freshness[index].record(now)
		poller.device.update(index, values[offset], true)
	}
}

func (poller *meterPoller) expireMeasurements(now int64) {
	for index := range poller.freshness {
		if poller.freshness[index].expire(now, poller.disconnect) {
			poller.device.update(uint8(index), 0, false)
		}
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
	go (&meterPoller{
		client:     newModbusClient(transport),
		device:     device,
		config:     config,
		disconnect: int64(time.Duration(config.DisconnectMs) * time.Millisecond),
	}).run()
	for {
		n.Poll()
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
	if config.DisconnectMs == 0 {
		config.DisconnectMs = defaultDisconnectMs
	}
	return config, nil
}

func haltBlink(period time.Duration) {
	for {
		pinStatus.High()
		time.Sleep(period)
		pinStatus.Low()
		time.Sleep(period)
	}
}
