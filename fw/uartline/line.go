//go:build tinygo

// Package uartline is a board-agnostic half-duplex serial transport over an
// RS485 UART. It owns the UART, the TX/RX alternate-function routing and the
// DE/RE direction pin, and exposes a byte-oriented Send/Receive interface with
// idle-gap framing. Higher layers (Modbus RTU, the JK NW protocol, …) build
// their own frames on top of it, so the same transport serves several device
// firmwares.
package uartline

import (
	"machine"
	"time"
	_ "unsafe"
)

// Default frame timing, applied by New when a Config field is left zero.
const (
	defaultTxSettle    = 2 * time.Millisecond
	defaultRxFirstByte = 120 * time.Millisecond
	defaultRxInterByte = 15 * time.Millisecond
)

// Config describes the RS485 wiring and timing for a Line.
type Config struct {
	// UART is the hardware UART to drive. If nil, machine.DefaultUART is used.
	UART *machine.UART
	// TX and RX are the UART data pins.
	TX machine.Pin
	RX machine.Pin
	// AltFunc is the GPIO alternate-function number that routes the USART to
	// the TX/RX pins (e.g. AF1 for PA2/PA3 = USART1 on PY32F030). The TinyGo
	// py32 UART driver does not route the peripheral to pins itself, so New
	// selects this alternate function on both TX and RX explicitly.
	AltFunc uint8
	// TxEn drives the transceiver's DE/RE direction control (active-high while
	// transmitting).
	TxEn machine.Pin
	// Baud is the serial bit rate (e.g. 9600 or 115200).
	Baud uint32
	// TxSettle is the guard time held before and after driving the bus. Zero
	// selects defaultTxSettle.
	TxSettle time.Duration
	// RxFirstByte is how long to wait for the first response byte. Zero selects
	// defaultRxFirstByte.
	RxFirstByte time.Duration
	// RxInterByte is the idle gap that ends a response frame. Zero selects
	// defaultRxInterByte.
	RxInterByte time.Duration
}

// Line is a configured half-duplex RS485 transport. It is not safe for
// concurrent use; drive it from a single goroutine.
type Line struct {
	uart        *machine.UART
	txEn        machine.Pin
	txSettle    time.Duration
	rxFirstByte time.Duration
	rxInterByte time.Duration
	scratch     [64]byte
}

// New configures the UART, the TX/RX alternate function and the direction pin,
// and returns a ready Line.
func New(cfg Config) *Line {
	txEn := cfg.TxEn
	txEn.Configure(machine.PinConfig{Mode: machine.PinOutput})
	txEn.Low()

	u := cfg.UART
	if u == nil {
		u = machine.DefaultUART
	}
	u.Configure(machine.UARTConfig{
		TX:       cfg.TX,
		RX:       cfg.RX,
		BaudRate: cfg.Baud,
	})

	// The py32 UART.Configure does not map the USART onto the requested pins,
	// so route TX and RX to the peripheral via their alternate function here.
	cfg.TX.Configure(machine.PinConfig{Mode: machine.PinAlternate})
	cfg.TX.SetAltFunc(cfg.AltFunc)
	cfg.RX.Configure(machine.PinConfig{Mode: machine.PinAlternate})
	cfg.RX.SetAltFunc(cfg.AltFunc)

	l := &Line{
		uart:        u,
		txEn:        txEn,
		txSettle:    cfg.TxSettle,
		rxFirstByte: cfg.RxFirstByte,
		rxInterByte: cfg.RxInterByte,
	}
	if l.txSettle == 0 {
		l.txSettle = defaultTxSettle
	}
	if l.rxFirstByte == 0 {
		l.rxFirstByte = defaultRxFirstByte
	}
	if l.rxInterByte == 0 {
		l.rxInterByte = defaultRxInterByte
	}
	return l
}

// FirstByteTimeout reports the configured wait for the first response byte.
func (l *Line) FirstByteTimeout() time.Duration { return l.rxFirstByte }

// InterByteTimeout reports the configured idle gap that ends a frame.
func (l *Line) InterByteTimeout() time.Duration { return l.rxInterByte }

// FlushRX drains any pending received bytes so a stale reply cannot be read as
// the response to the next request.
func (l *Line) FlushRX() {
	for {
		n, _ := l.uart.Read(l.scratch[:])
		if n == 0 {
			return
		}
	}
}

// Send drives the bus (DE/RE high), writes data with a settle guard on both
// sides, then releases the bus so the reply can be received.
func (l *Line) Send(data []byte) {
	l.FlushRX()
	l.txEn.High()
	time.Sleep(l.txSettle)
	_, _ = l.uart.Write(data)
	time.Sleep(l.txSettle)
	l.txEn.Low()
}

// Receive reads a reply frame into buf using idle-gap framing. It waits up to
// the first-byte timeout for the first byte, then keeps reading until an
// inter-byte idle gap ends the frame, buf is full, or (when stopAt > 0) at
// least stopAt bytes have arrived. It returns the number of bytes read.
//
// stopAt lets fixed-length protocols (Modbus RTU) return as soon as the whole
// frame is in without waiting out an extra idle gap; pass 0 for
// variable-length protocols that end on the idle gap alone.
func (l *Line) Receive(buf []byte, stopAt int) int {
	total := 0
	deadline := monoNanos() + int64(l.rxFirstByte)
	for total < len(buf) {
		n, _ := l.uart.Read(buf[total:])
		if n > 0 {
			total += n
			if stopAt > 0 && total >= stopAt {
				break
			}
			deadline = monoNanos() + int64(l.rxInterByte)
			continue
		}
		if monoNanos() >= deadline {
			break
		}
		time.Sleep(250 * time.Microsecond)
	}
	return total
}

//go:linkname monoNanos runtime.nanotime
func monoNanos() int64
