//go:build tinygo

package main

import (
	"encoding/binary"
	"machine"
	"runtime"
	"time"

	"github.com/burgrp/bleriot/lib/node/pan211x"
	"github.com/burgrp/bleriot/lib/shared/config"
	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

const (
	rs485TxPin   = machine.PA0
	rs485RxPin   = machine.PA3
	rs485TxEnPin = machine.PA5
	statusLedPin = machine.PB0
	debugPin     = machine.PA7
	rfDataPin    = machine.PA1
	rfSckPin     = machine.PA2
	rfCsPin      = machine.PA4
	readRetries  = 5
)

// Command (firmware) main wires the RS485 bridge device to the BleRiot runtime.
func main() {
	dev := newRuntimeDevice()

	led := statusLedPin
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High()

	rfSck := rfSckPin
	rfData := rfDataPin
	rfCs := rfCsPin

	n, cfgBytes, err := pan211x.StartNode(&spec.Chip, rfSck, rfData, rfCs, dev)
	if err != nil {
		if config.IsUnprovisioned(err) {
			haltBlink(led, 1000*time.Millisecond)
		}
		haltBlink(led, 100*time.Millisecond)
	}
	dev.bindNode(n)

	bridgeCfg := decodeBridgeConfig(cfgBytes)
	if bridgeCfg.PollMs == 0 {
		bridgeCfg.PollMs = 1000
	}

	devices := configuredDevices(bridgeCfg)
	grouped := make([]groupedDevice, len(devices))
	for i := range devices {
		grouped[i] = newGroupedDevice(devices[i])
	}

	client, err := newUARTRS485Client(bridgeCfg.Baud)
	if err != nil {
		haltBlink(led, 200*time.Millisecond)
	}
	poll := newPoller(client, dev, readRetries)

	go func() {
		for {
			for i := range grouped {
				poll.pollDevice(grouped[i])
			}
			time.Sleep(time.Duration(bridgeCfg.PollMs) * time.Millisecond)
		}
	}()

	for {
		n.Poll()
		runtime.Gosched()
	}
}

func decodeBridgeConfig(raw []byte) spec.Config {
	cfg := spec.DefaultConfig()
	if len(raw) >= 8 {
		cfg.SlaveAddr1 = raw[0]
		cfg.SlaveAddr2 = raw[1]
		cfg.PollMs = binary.LittleEndian.Uint16(raw[2:4])
		cfg.Baud = binary.LittleEndian.Uint32(raw[4:8])
	}
	if cfg.SlaveAddr1 == 0 {
		cfg.SlaveAddr1 = 1
	}
	if cfg.SlaveAddr2 == 0 {
		cfg.SlaveAddr2 = 2
	}
	if cfg.Baud == 0 {
		cfg.Baud = 9600
	}
	return cfg
}

func configuredDevices(cfg spec.Config) []spec.DeviceDef {
	devices := make([]spec.DeviceDef, 0, 2)
	if cfg.SlaveAddr1 != 0 {
		devices = append(devices, spec.DeviceDef{SlaveAddr: cfg.SlaveAddr1, Baud: cfg.Baud, Registers: spec.RegistersForMeterSlot(0)})
	}
	if cfg.SlaveAddr2 != 0 {
		devices = append(devices, spec.DeviceDef{SlaveAddr: cfg.SlaveAddr2, Baud: cfg.Baud, Registers: spec.RegistersForMeterSlot(1)})
	}
	return devices
}

func haltBlink(led machine.Pin, period time.Duration) {
	for {
		led.High()
		time.Sleep(period)
		led.Low()
		time.Sleep(period)
	}
}
