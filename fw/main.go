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

// Command (firmware) main wires the RS485 bridge device to the BleRiot runtime.
func main() {
	pins := boardConfig()
	dev := newRuntimeDevice()

	ledPin, err := pinByName(pins.LedPin)
	if err == nil {
		ledPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		ledPin.High()
	}

	rfSck, err := pinByName(pins.RFSckPin)
	if err != nil {
		haltBlink(ledPin, 100*time.Millisecond)
	}
	rfData, err := pinByName(pins.RFDataPin)
	if err != nil {
		haltBlink(ledPin, 100*time.Millisecond)
	}
	rfCs, err := pinByName(pins.RFCsPin)
	if err != nil {
		haltBlink(ledPin, 100*time.Millisecond)
	}

	n, cfgBytes, err := pan211x.StartNode(&spec.Chip, rfSck, rfData, rfCs, dev)
	if err != nil {
		if config.IsUnprovisioned(err) {
			haltBlink(ledPin, 1000*time.Millisecond)
		}
		haltBlink(ledPin, 100*time.Millisecond)
	}
	dev.bindNode(n)

	bridgeCfg := decodeBridgeConfig(cfgBytes)
	if bridgeCfg.PollMs == 0 {
		bridgeCfg.PollMs = 1000
	}

	devices := []spec.DeviceDef{
		{Name: "em.grid", SlaveAddr: bridgeCfg.GridAddr, Baud: bridgeCfg.Baud, Registers: spec.GridRegisters},
		{Name: "em.house", SlaveAddr: bridgeCfg.HouseAddr, Baud: bridgeCfg.Baud, Registers: spec.HouseRegisters},
	}

	client, err := newUARTRS485Client(pins, bridgeCfg.Baud)
	if err != nil {
		haltBlink(ledPin, 200*time.Millisecond)
	}
	poll := newPoller(client, dev, pins.ReadRetries)

	go func() {
		for {
			for i := range devices {
				poll.pollDevice(devices[i])
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
		cfg.GridAddr = raw[0]
		cfg.HouseAddr = raw[1]
		cfg.PollMs = binary.LittleEndian.Uint16(raw[2:4])
		cfg.Baud = binary.LittleEndian.Uint32(raw[4:8])
	}
	if cfg.GridAddr == 0 {
		cfg.GridAddr = 1
	}
	if cfg.HouseAddr == 0 {
		cfg.HouseAddr = 2
	}
	if cfg.Baud == 0 {
		cfg.Baud = 9600
	}
	return cfg
}

func haltBlink(led machine.Pin, period time.Duration) {
	for {
		led.High()
		time.Sleep(period)
		led.Low()
		time.Sleep(period)
	}
}
