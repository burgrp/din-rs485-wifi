//go:build tinygo

// Package rig is the shared node runtime for the DIN RS485 boards. It owns the
// fixed vertical-board pin mapping, the BleRiot node lifecycle and the poll
// loop, so each device firmware only supplies a Source: which tags it exposes
// and how to poll them. One node drives one slave, so there is no addressing
// map to decode; the only runtime knob is the poll interval.
package rig

import (
	"machine"
	"runtime"
	"time"

	"github.com/burgrp/bleriot/lib/node/pan211x"
	"github.com/burgrp/bleriot/lib/shared/config"
	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

// Vertical board (PY32F003W1xSx SOP16 + PAN211x) pin mapping. The RS485 bus runs
// on USART2 (TX PA0/AF9, RX PA3/AF4) with DE/RE on PA5; TX and RX need different
// alternate functions because PA0 only exposes USART2_TX on AF9. The debug
// console lives on USART1_TX (PA7, AF8) so a USB-serial adapter shows the
// `--serial uart` log. The PAN211x radio uses a bit-banged SPI on plain GPIO,
// and the status LED is on PB0. See the PY32F003 datasheet port-A AF table
// (PA0=USART2_TX/AF9, PA3=USART2_RX/AF4, PA7=USART1_TX/AF8).
const (
	UartTx   = machine.PA0 // RS485_TX, USART2_TX (AF9)
	UartRx   = machine.PA3 // RS485_RX, USART2_RX (AF4)
	UartTxAF = 9           // USART2_TX alternate function for PA0
	UartRxAF = 4           // USART2_RX alternate function for PA3
	TxEn     = machine.PA5 // RS485 DE/RE direction (RS485_TXEN)
	Led      = machine.PB0 // status LED
	DebugTx  = machine.PA7 // USART1_TX (DEBUG) for the --serial uart console
	DebugAF  = 8           // USART1_TX alternate function for PA7
	RfData   = machine.PA1 // PAN211x DATA (bit-banged SPI), bidirectional
	RfSck    = machine.PA2 // PAN211x SCK  (bit-banged SPI)
	RfCs     = machine.PA4 // PAN211x CSN  (bit-banged SPI), active-low

	defaultPollMs = 1000
)

// UartBus is the USART peripheral the RS485 transport drives. USART2 keeps the
// bus off USART1, which the debug console owns.
var UartBus = machine.UART2

// Source is a device-specific data source driven by the runtime. Tags reports
// the fixed set of wire tags the device exposes (used to size the register
// table); Poll reads the slave once and writes results via dev.Set.
type Source interface {
	Tags() []uint16
	Poll(dev *Device)
}

// Run wires a Source to the BleRiot runtime and never returns. It configures
// the status LED, starts the node (blinking on a provisioning error), decodes
// the poll interval, then polls the source on a timer while servicing RF.
func Run(src Source) {
	led := Led
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High()

	// Route the debug console's USART1_TX onto its header pin. The runtime has
	// already configured USART1 (machine.Serial) for the `--serial uart` build;
	// the py32 UART driver does not map the peripheral to pins itself.
	DebugTx.Configure(machine.PinConfig{Mode: machine.PinAlternate})
	DebugTx.SetAltFunc(DebugAF)

	dev := newDevice(src.Tags())

	n, cfgBytes, err := pan211x.StartNode(&spec.Chip, RfSck, RfData, RfCs, dev)
	if err != nil {
		if config.IsUnprovisioned(err) {
			haltBlink(led, 1000*time.Millisecond)
		}
		haltBlink(led, 100*time.Millisecond)
	}
	dev.bindNode(n)

	cfg := spec.DecodeConfig(cfgBytes)
	pollMs := cfg.PollMs
	if pollMs == 0 {
		pollMs = defaultPollMs
	}

	go func() {
		for {
			led.Set(!led.Get())
			src.Poll(dev)
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
