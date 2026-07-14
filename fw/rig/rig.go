//go:build tinygo

// Package rig is the shared node runtime for the DIN RS485 boards. It owns the
// fixed BOB pin mapping, the BleRiot node lifecycle and the poll loop, so each
// device firmware only supplies a Source: which tags it exposes and how to
// poll them. One node drives one slave, so there is no addressing map to
// decode; the only runtime knob is the poll interval.
package rig

import (
	"machine"
	"runtime"
	"time"

	"github.com/burgrp/bleriot/lib/node/pan211x"
	"github.com/burgrp/bleriot/lib/shared/config"
	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

// BOB breakout board (PY32F030 + PAN211x) pin mapping. The debug console lives
// on USART1 (BOB's PROG-header UART, PB6/PB7, AF0) so a USB-serial adapter on
// that header shows the `--serial uart` log. The RS485 bus runs on USART2
// (PA2/PA3, AF4) with DE/RE on PA4. The RF and LED pins follow
// bleriot/example/bob. See the PY32F030/F003 datasheet port-A/port-B AF maps
// (PA2=USART2_TX/AF4, PA3=USART2_RX/AF4, PB6=USART1_TX/AF0, PB7=USART1_RX/AF0).
const (
	UartTx  = machine.PA2  // USART2_TX (breakout header)
	UartRx  = machine.PA3  // USART2_RX (breakout header)
	UartAF  = 4            // USART2 alternate function for PA2/PA3
	TxEn    = machine.PA4  // RS485 DE/RE direction (header GPIO)
	Led     = machine.PB0  // red LED
	DebugTx = machine.PB6  // USART1_TX (PROG header) for the --serial uart console
	DebugAF = 0            // USART1 alternate function for PB6
	RfData  = machine.PA7  // PAN211x DATA (SPI1_MOSI), bidirectional
	RfSck   = machine.PA9  // PAN211x SCK  (SPI1_SCK)
	RfCs    = machine.PA10 // PAN211x CSN  (SPI1_NSS), active-low

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
