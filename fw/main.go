//go:build tinygo

package main

import (
	"machine"
	"runtime"
	"time"

	"github.com/burgrp/bleriot/lib/node/pan211x"
	"github.com/burgrp/bleriot/lib/shared/config"
	"github.com/burgrp/din-rs485-wifi/fw/rs485"
	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

const (
	// --- Pin mapping: DIN RS485 board (vertical/horizontal, PY32F003) ---
	// rs485TxPin   = machine.PA0
	// rs485RxPin   = machine.PA3
	// rs485TxEnPin = machine.PA5
	// statusLedPin = machine.PB0
	// debugPin     = machine.PA7
	// rfDataPin    = machine.PA1
	// rfSckPin     = machine.PA2
	// rfCsPin      = machine.PA4

	// --- Pin mapping: BOB breakout board (PY32F030 + PAN211x) ---
	// RF and LED pins follow bleriot/example/bob and the bob schematic; the
	// RS485 UART uses USART1 on PA2/PA3 (both AF1) with DE/RE and debug on free
	// header GPIO. See PY32F030 datasheet, port-A AF map (PA2=USART1_TX/AF1,
	// PA3=USART1_RX/AF1).
	rs485TxPin   = machine.PA2  // USART1_TX (header)
	rs485RxPin   = machine.PA3  // USART1_RX (header)
	rs485UartAF  = 1            // USART1 alternate function for PA2/PA3
	rs485TxEnPin = machine.PA4  // RS485 DE/RE direction (header GPIO)
	statusLedPin = machine.PB0  // red LED
	debugPin     = machine.PA6  // header GPIO
	rfDataPin    = machine.PA7  // PAN211x DATA (SPI1_MOSI), bidirectional
	rfSckPin     = machine.PA9  // PAN211x SCK  (SPI1_SCK)
	rfCsPin      = machine.PA10 // PAN211x CSN  (SPI1_NSS), active-low

	readRetries   = 5
	modbusBaud    = 9600
	defaultPollMs = 1000
)

// Command (firmware) main wires the generic RS485 bridge device to the BleRiot
// runtime. All meter semantics live on the hub; the device only polls the
// Modbus addresses named in its provisioned Config and notifies encoded tags.
func main() {
	dev := newRuntimeDevice()

	led := statusLedPin
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High()

	n, cfgBytes, err := pan211x.StartNode(&spec.Chip, rfSckPin, rfDataPin, rfCsPin, dev)
	if err != nil {
		if config.IsUnprovisioned(err) {
			haltBlink(led, 1000*time.Millisecond)
		}
		haltBlink(led, 100*time.Millisecond)
	}
	dev.bindNode(n)

	cfg := decodeConfig(cfgBytes)
	pollMs := cfg.PollMs
	if pollMs == 0 {
		pollMs = defaultPollMs
	}

	plan := buildPlan(cfg)
	dev.configure(plan)

	client, err := rs485.New(rs485.Config{
		TX:      rs485TxPin,
		RX:      rs485RxPin,
		AltFunc: rs485UartAF,
		TxEn:    rs485TxEnPin,
		Baud:    modbusBaud,
	})
	if err != nil {
		haltBlink(led, 200*time.Millisecond)
	}
	poll := newPoller(client, dev, plan, readRetries)

	pollMs = 20
	println("poll ms", pollMs)

	go func() {
		for {
			ms := runtime.MemStats{}
			runtime.ReadMemStats(&ms)
			println("mem", ms.HeapAlloc)
			statusLedPin.Set(!statusLedPin.Get())

			poll.pollAll()
			time.Sleep(time.Duration(pollMs) * time.Millisecond)

		}
	}()

	for {
		n.Poll()
		runtime.Gosched()
	}
}

func haltBlink(led machine.Pin, period time.Duration) {
	for {
		led.High()
		time.Sleep(period)
		led.Low()
		time.Sleep(period)
	}
}
